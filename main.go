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
	l "log"
)

func api_request(conf config.Config, request string, symbol tools.Symbol, tRetrieve time.Time) (tools.Response, error) {

	var err error
	var data []byte
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}}
	c := make(chan error, 1)
	var resp *http.Response

	var req_url string

	if request == "get_symbol" {
		req_url = conf.API.Url + request + "/" + symbol.Name + "/" + tRetrieve.Format("2006-01-02")
	} else if request == "update_symbol" {
		req_url = conf.API.Url + request + "/" + symbol.Name + "/" + symbol.Last_insert.Format("2006-01-02")
	} else {
		return res, errors.New("bad request")
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

func retrieve_max_import(conf config.Config, symbols *[]tools.Symbol) error {
	var err error
	var res tools.Response
	var tempo_symbols []tools.Symbol

	for _, symbol := range *symbols {

		log.Info("Retrieve last data insert for ", symbol.Name)

		var i = -1
		var ct int
		t := time.Now().UTC()
		tRetrieve := t.AddDate(0, 0, i)

	retry_api_request:

		// Requete sur l'api avec tRetrieve
		if res, err = api_request(conf, "get_symbol", symbol, tRetrieve); err != nil {
			return err
		}

		// Si CT >= 30 donc 31 jours, on considère que la base et vide et on prend cette valeur à importer
		if ct >= 30 {
			symbol.Last_insert = tRetrieve
			tempo_symbols = append(tempo_symbols, symbol)
			continue
		}
		ct++

		// Si l'api ne retourne aucun bid ça veut dire que le plus ancien n'est pas dans cette plage de temps, on l'augmente donc de 10
		if len(res.Bids) == 0 {
			i -= 1
			tRetrieve = t.AddDate(0, 0, i)
			goto retry_api_request
		}

		// Une fois qu'il y a au moins un bid, on cherche le plus récent et on prend cette date comme référence
		for _, bid := range res.Bids {
			if bid.Bid_at, err = time.Parse("2006-01-02T15:04:05", bid.Bid_at_s); err != nil {
				return err
			}

			if bid.Bid_at.After(symbol.Last_insert) {
				symbol.Last_insert = bid.Bid_at
			}
		}

		tempo_symbols = append(tempo_symbols, symbol)
	}

	*symbols = tempo_symbols

	return nil
}

const vers_algo = "v0.0.3"

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

	if err = retrieve_max_import(conf, &conf.API.Symbols); err != nil {
		log.FatalError(err)
		return
	}

	fmt.Println("")
	log.WhiteInfo("Max import retrieved :")

	for _, symbol := range conf.API.Symbols {
		log.Info(symbol.Name, " : ", symbol.Last_insert.Format("2006-01-02 15:04:05"), " (UTC)")
	}

	fmt.Println("")
	log.WhiteInfo("Start current retrieve")
	log.Info("#############################")

	for i, symbol := range conf.API.Symbols {
		go retrieve_symbol(&conf, symbol, i)
	}

	for {
		time.Sleep(24 * time.Hour)
	}
}

func retrieve_symbol(conf *config.Config, symbol tools.Symbol, i int) {

	var err error
	var h_last_d = 3
	var m_last_d = 5
	var dStep = 1 * time.Minute

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

		//log.Info("#########")
		var tUpdate = symbol.Last_insert

		h, m, _ := tNow.Clock()

		var change_update_date bool

		switch {
		case h < h_last_d:
				change_update_date = true
				break
		case h == h_last_d && m <= m_last_d:
				change_update_date = true
				break
		}

		if change_update_date {
				tUpdate = tUpdate.AddDate(0, 0, -1)
		}

		//log.Info("Retrieve data for ", symbol.Name, " between ", tUpdate.Format("2006-01-02"), " and ", tNow.Format("2006-01-02 15:04:05"), " (UTC)")

	  res, err = api_request(*conf, "update_symbol", symbol, time.Time{})

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
				return
			}
		}

		conf.API.Symbols[i].Last_insert = time.Now().UTC()

		dDiff := time.Now().UTC().Sub(tNow)

		var dStepTempo time.Duration

		if dDiff >= dStep/2 {
			dStepTempo = dStep
		} else {
			dStepTempo = dStep-dDiff
		}

		log.Info(res.Error.MessageError, " | duration : ", dDiff, " | next retrieve in : ", dStepTempo)

		//log.Info(dDiff, " next -> ", dStepTempo)

		//log.Info("#########")

		time.Sleep(dStepTempo)
	}
}
