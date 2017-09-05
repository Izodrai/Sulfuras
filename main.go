package main

import (
	"encoding/json"
	"fmt"
	l "log"
	"os"
	"path"
	"strconv"
	"time"

	"./lib/api"
	"./lib/calculate"
	"./lib/config"
	"./lib/log"
	"./lib/tools"
)

const vers_algo = "v0.0.5"

func main() {

	var err error
	var rattrap_mode bool
	var conf config.Config

	if len(os.Args) != 2 && len(os.Args) != 3 {
		l.Println(log.RED + "Invalid Argument(s)" + log.STOP)
		l.Println(log.RED + "Usuel : ./market-binary config_file mode (optional)" + log.STOP)
		l.Println(log.RED + "Mode : (1) rattrap" + log.STOP)
		os.Exit(1)
	}

	if err = conf.LoadConfig(path.Join(os.Args[1])); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	if err = log.InitLog(true, conf); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	if len(os.Args) == 3 {
		var i int
		if i, err = strconv.Atoi(os.Args[2]); err != nil {
			log.Error(log.RED + err.Error() + log.STOP)
			os.Exit(1)
		}

		if i != 1 {
			log.Error(log.RED + "Bad mode ! " + os.Args[2] + log.STOP)
			os.Exit(1)
		} else {
			rattrap_mode = true
		}
	}

	fmt.Println("")
	fmt.Println("")

	if rattrap_mode {
		log.YellowInfo("Running Sulfuras-Rattrap")
	} else {
		log.YellowInfo("Running Sulfuras with : ", vers_algo)
	}

	fmt.Println("")

	log.WhiteInfo("Start current retrieve")
	log.Info("#############################")

	if rattrap_mode {
		for _, symbol := range conf.API.Symbols {
			log.WhiteInfo(symbol.Name)
			rattrap_calc_symbol(&conf, symbol)
			time.Sleep(1 * time.Second)
		}
		log.Info("#############################")
		log.WhiteInfo("End rattrap")
		log.Info("#############################")
		os.Exit(0)
	} else {
		for _, symbol := range conf.API.Symbols {
			go retrieve_symbol(&conf, symbol)
			time.Sleep(5 * time.Second)
		}

		for {
			time.Sleep(24 * time.Hour)
		}
	}
}

func retrieve_symbol(conf *config.Config, symbol tools.Symbol) {

	var err error
	var dStep = 2*time.Minute + 30*time.Second

	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}}

	for {
		var tNow = time.Now()

		if per, ok := conf.API.RetrievePeriode[tNow.Weekday()]; ok {
			if per.Deactivated {
				log.Info(tNow.Weekday(), " is deactivated")
				y, m, d := tNow.Date()
				tTomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, tNow.Location())
				log.Info("tNow (", tNow.Format("2006-01-02 15:04:05"), "), tTomorrow (", tTomorrow.Format("2006-01-02 15:04:05"), ")")
				log.Info("Sleep ", tTomorrow.Sub(tNow).String(), " before tomorrow")
				time.Sleep(tTomorrow.Sub(tNow))
				continue
			}

			if per.Limited {
				y, m, d := tNow.Date()
				tStart := time.Date(y, m, d, per.Start_h, per.Start_m, per.Start_s, 0, tNow.Location())
				tEnd := time.Date(y, m, d, per.End_h, per.End_m, per.End_s, 0, tNow.Location())
				tTomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, tNow.Location())

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

		res, err = api.Api_request(*conf, symbol, tools.Bid{}, 1, time.Time{}, time.Time{})

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

		for _, res_b := range res.Bids {
			if res_b.Calculations_s != "{}" {
				err = json.Unmarshal([]byte(res_b.Calculations_s), &res_b.Calculations)
				if err != nil {
					log.Error("For : ", res_b, " - ", err.Error())
					continue
				}
			}
			res_bids = append(res_bids, res_b)
		}

		for _, b_to_update := range calculate.Calculate_bids(conf, res_bids) {
			res, err = api.Api_request(*conf, tools.Symbol{}, b_to_update, 2, time.Time{}, time.Time{})
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

		dDiff := time.Now().Sub(tNow)

		var dStepTempo time.Duration

		if dDiff >= dStep/2 {
			dStepTempo = dStep
		} else {
			dStepTempo = dStep - dDiff
		}

		log.Info("Data for symbol ", symbol.Id, "(", symbol.Name, ") | duration : ", dDiff, " | next retrieve in : ", dStepTempo)

		time.Sleep(dStepTempo)
	}
}

func rattrap_calc_symbol(conf *config.Config, symbol tools.Symbol) {

	var err error
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}}

	res, err = api.Api_request(*conf, symbol, tools.Bid{}, 3, conf.API.From, conf.API.To)

	if err != nil {
		switch {
		case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
			log.Error("For : ", symbol.Name, " - ", err.Error())
			return
		case err.Error() == "No data to retrieve in this range":
			log.Error("For : ", symbol.Name, " - ", err.Error())
			return
		default:
			log.FatalError(err)
			return
		}
	}

	var res_bids []tools.Bid

	for _, res_b := range res.Bids {
		if res_b.Calculations_s != "{}" {
			err = json.Unmarshal([]byte(res_b.Calculations_s), &res_b.Calculations)
			if err != nil {
				log.Error("For : ", res_b, " - ", err.Error())
				continue
			}
		}
		res_bids = append(res_bids, res_b)
	}

	for _, b_to_update := range calculate.Calculate_bids(conf, res_bids) {
		log.Info("Update calculations for -> ", b_to_update.Id, " | ", b_to_update.Symbol.Name, " | ", b_to_update.Bid_at_s, " | ", b_to_update.Last_bid, " | ", b_to_update.Calculations)
		res, err = api.Api_request(*conf, tools.Symbol{}, b_to_update, 2, time.Time{}, time.Time{})
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
}
