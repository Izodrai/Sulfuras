package main

import (
	"./lib/config"
	"./lib/log"
	"./lib/tools"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"os"
	"path"
	"strconv"
	l "log"
)

func api_request(conf config.Config, symbol tools.Symbol, bid_to_update tools.Bid, t int) (tools.Response, error) {

	var err error
	var data []byte
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}}
	c := make(chan error, 1)
	var resp *http.Response
	var req_url string

	switch t {
		case 1:
      req_url = conf.API.Url + "feed_symbol_from_last_insert/" + strconv.Itoa(symbol.Id) + "/"
		case 2:
			req_url = conf.API.Url + "set_calculation/" + strconv.Itoa(bid_to_update.Id) + "/" + bid_to_update.Base64Calculations()
  }

	go func() {
		resp, err = http.Get(req_url)
		c <- err
	}()

	select {
	case err := <-c:
		if err != nil {
			return res, err
		}
	case <-time.After(time.Second * 350):
		return res, errors.New("HTTP source timeout")
	}

	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return res, err
	}

	if res.Error.IsAnError {
		return res, errors.New(res.Error.MessageError)
	}

	return res, nil
}

const vers_algo = "v0.0.4"

func main() {
	var err error

	var conf config.Config

	if len(os.Args) != 2 {
		l.Println(log.RED+ "Invalid Argument(s)"+ log.STOP)
		l.Println(log.RED+ "Usuel : ./market-binary config_file"+ log.STOP)
		os.Exit(1)
	}

	if err = conf.LoadConfig(path.Join(os.Args[1])); err != nil {
		l.Println(log.RED+err.Error()+log.STOP)
		os.Exit(1)
	}

	if err = log.InitLog(true, conf); err != nil {
		l.Println(log.RED+err.Error()+log.STOP)
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("")

	log.YellowInfo("Running Sulfuras with : ", vers_algo)

	fmt.Println("")

	log.WhiteInfo("Start current retrieve")
	log.Info("#############################")

	for _, symbol := range conf.API.Symbols {
		go retrieve_symbol(&conf, symbol)
		time.Sleep(5 * time.Second)
	}

	for {
		time.Sleep(24 * time.Hour)
	}
}

func retrieve_symbol(conf *config.Config, symbol tools.Symbol) {

	var err error
	var dStep = 2 * time.Minute + 30 * time.Second

	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}}

	for {
		var tNow = time.Now().UTC()

		if per, ok := conf.API.RetrievePeriode[tNow.Weekday()]; ok {
			if per.Deactivated {
				log.Info(tNow.Weekday(), " is deactivated")
				y, m, d := tNow.Date()
				tTomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC)
				log.Info("tNow (", tNow.Format("2006-01-02 15:04:05"), "), tTomorrow (", tTomorrow.Format("2006-01-02 15:04:05"), ")")
				log.Info("Sleep ", tTomorrow.Sub(tNow).String(), " before tomorrow")
				time.Sleep(tTomorrow.Sub(tNow))
				continue
			}

			if per.Limited {
				y, m, d := tNow.Date()
				tStart := time.Date(y, m, d, per.Start_h, per.Start_m, per.Start_s, 0, time.UTC)
				tEnd := time.Date(y, m, d, per.End_h, per.End_m, per.End_s, 0, time.UTC)
				tTomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC)

				if tNow.Before(tStart) {
					log.Info(tNow.Weekday(), " is limited")
					log.Info("tNow (", tNow.Format("2006-01-02 15:04:05"), ") before tStart (", tStart.Format("2006-01-02 15:04:05"), ")")
					log.Info("Sleep : ", tStart.Sub(tNow).String())
					time.Sleep(tStart.Sub(tNow))
					continue
				}

				if tNow.After(tEnd) {
					log.Info(tNow.Weekday(), " is limited")
					log.Info("tNow (", tNow.Format("2006-01-02 15:04:05"), ") after tEnd (", tEnd.Format("2006-01-02 15:04:05"), "), tTomorrow (", tTomorrow.Format("2006-01-02 15:04:05"), ")")
					log.Info("Sleep ", tTomorrow.Sub(tNow).String(), " before tomorrow")
					time.Sleep(tTomorrow.Sub(tNow))
					continue
				}
			}
		}

	  res, err = api_request(*conf, symbol, tools.Bid{}, 1)

		if err != nil {
			switch {
			case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
				log.Error("For : ", symbol.Name, " - ", err.Error())
				continue
			case err.Error() == "No data to retrieve in this range":
				log.Error("For : ", symbol.Name, " - ", err.Error())
				continue
			default:
				log.FatalError(err)
				continue
			}
		}

		var res_bids []tools.Bid

		for _,res_b := range res.Bids {
			if res_b.Calculations_s != "{}" {
				err = json.Unmarshal([]byte(res_b.Calculations_s), &res_b.Calculations)
				if err != nil {
					log.Error("For : ", res_b, " - ", err.Error())
					continue
				}
			}
			res_bids = append(res_bids, res_b)
		}

		for _,b_to_update := range calculate_bids(conf, res_bids) {
			log.Info("Update calculations for -> ", b_to_update)
			res, err = api_request(*conf, tools.Symbol{}, b_to_update, 2)
			if err != nil {
				switch {
				case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
					log.Error(err.Error())
					continue
				default:
					log.FatalError(err)
					continue
				}
			}
		}

		dDiff := time.Now().UTC().Sub(tNow)

		var dStepTempo time.Duration

		if dDiff >= dStep/2 {
			dStepTempo = dStep
		} else {
			dStepTempo = dStep-dDiff
		}

		log.Info(res.Error.MessageError, " | duration : ", dDiff, " | next retrieve in : ", dStepTempo)

		time.Sleep(dStepTempo)
	}
}

func calculate_bids(conf *config.Config, res_bids []tools.Bid) []tools.Bid {

	var calc_bids []tools.Bid

	calc_bids = calc_sma(conf, res_bids)

	return sort_calc(conf, res_bids, calc_bids)
}

func calc_sma(conf *config.Config, res_bids []tools.Bid) []tools.Bid{

	var calc_bids []tools.Bid
	var sma_conf = make(map[int][]float64)

	for _,co_sma := range conf.API.Calculations.SMA {
		sma_conf[co_sma] = []float64{}
	}

	for i,res_b := range res_bids {

		var b tools.Bid
		b.Id = res_b.Id
		b.Symbol = res_b.Symbol
		b.Bid_at_s = res_b.Bid_at_s
		b.Last_bid = res_b.Last_bid
		b.Calculations_s = res_b.Calculations_s
		b.Calculations = make(map[string]float64)

		calc_bids = append(calc_bids, res_b)

		for co_sma,_ := range sma_conf {
			sma_conf[co_sma] = append(sma_conf [co_sma], res_b.Last_bid)

			if len(sma_conf[co_sma]) > co_sma {
				sma_conf[co_sma] = append(sma_conf[co_sma][:0], sma_conf[co_sma][1:]...)
			} else if len(sma_conf[co_sma]) < co_sma {
				continue
			}

			var ma float64
			for _,last_b := range sma_conf[co_sma] {
				ma = ma + last_b
			}
			ma = ma/float64(co_sma)

			b.Calculations["sma_"+strconv.Itoa(co_sma)] = ma
		}

		calc_bids[i] = b
	}

	return calc_bids
}

func sort_calc(conf *config.Config, res_bids, calc_bids []tools.Bid) []tools.Bid {

	var bids_to_update []tools.Bid

	for i, calc_b := range calc_bids {

		var b = res_bids[i]
		b.Calculations = map[string]float64{}

		var diff = false

		for t, val := range res_bids[i].Calculations {
			b.Calculations[t]=val
		}

		for t, calc_val := range calc_b.Calculations {
			if res_val, ok := b.Calculations[t]; ok {
				if calc_val == res_val {
					continue
				}
			}
			b.Calculations[t]=calc_val
			diff = true
		}

		if !diff {
			continue
		}

		bids_to_update = append(bids_to_update,b)
	}

	return bids_to_update
}
