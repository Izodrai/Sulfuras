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

func api_request(conf config.Config, request string, symbol tools.Symbol, tRetrieve time.Time) (tools.Response, error) {

	var err error
	var data []byte
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}}
	c := make(chan error, 1)
	var resp *http.Response

	var req_url = conf.API.Url + request + "/" + strconv.Itoa(symbol.Id) + "/"

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

	log.WhiteInfo("Start current retrieve")
	log.Info("#############################")

	for i, symbol := range conf.API.Symbols {
		go retrieve_symbol(&conf, symbol, i)
		time.Sleep(5 * time.Second)
	}

	for {
		time.Sleep(24 * time.Hour)
	}
}

func retrieve_symbol(conf *config.Config, symbol tools.Symbol, i int) {

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

	  res, err = api_request(*conf, "feed_symbol_from_last_insert", symbol, time.Time{})

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
