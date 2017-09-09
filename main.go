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
	var mode int
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

	if err = api.Get_symbols_status(&conf); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	if err = log.InitLog(true, conf); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	if len(os.Args) == 3 {
		if mode, err = strconv.Atoi(os.Args[2]); err != nil {
			log.Error(log.RED + err.Error() + log.STOP)
			os.Exit(1)
		}

		if mode != 1 {
			log.Error(log.RED + "Bad mode ! " + os.Args[2] + log.STOP)
			os.Exit(1)
		}
	}

	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")

	if mode == 1 {
		log.YellowInfo("Running Sulfuras-Rattrap")
	} else {
		log.YellowInfo("Running Sulfuras version : ", vers_algo)
	}

	fmt.Println("")
	fmt.Println("")
	log.WhiteInfo("> Symbol Configuration")
	fmt.Println("")

	for _, s := range conf.API.Symbols {
		log.Info("\tSymbol id : ", s.Id, " (", s.Name, ") | Status : ", s.State)
	}

	fmt.Println("")
	log.WhiteInfo("> Days Configuration")
	fmt.Println("")

	for i := 0; i < 7; i++ {
		p := conf.API.RetrievePeriode[time.Weekday(i)]
		log.WhiteInfo(">> ", time.Weekday(i).String())
		log.Info("\tDeactivated : ", p.Deactivated)
		log.Info("\tLimited     : ", p.Limited)
		log.Info("\tStart time  : ", p.Start)
		log.Info("\tEnd time    : ", p.End)
		fmt.Println("")
	}

	log.Info("##########")
	fmt.Println("")
	fmt.Println("")
	log.WhiteInfo("Start current retrieve")
	log.Info("#############################")

	if mode == 1 {

		for _, symbol := range conf.API.Symbols {
			log.WhiteInfo(symbol.Name)
			rattrap_calc_symbol(&conf, symbol)
			time.Sleep(1 * time.Second)
		}
		time.Sleep(10 * time.Second)

		log.Info("#############################")
		log.WhiteInfo("End rattrap")
		log.Info("#############################")
		os.Exit(0)

	} else {

		for _, symbol := range conf.API.Symbols {
			go exec_symbol(&conf, symbol)
			time.Sleep(5 * time.Second)
		}

		for {
			time.Sleep(24 * time.Hour)
		}

	}
}

func exec_symbol(conf *config.Config, symbol tools.Symbol) {

	var err error
	var dStep = 2*time.Minute + 30*time.Second
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	var open_trades = make(map[int]tools.Trade)

	get_opened_trades(conf, open_trades)

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

		res, err = api.Api_request(*conf, symbol, tools.Bid{}, tools.Trade{}, 1, time.Time{}, time.Time{})

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

		calc_bid := calculate.Calculate_bids(conf, res_bids)
		for _, b_to_update := range calc_bid {
			res, err = api.Api_request(*conf, tools.Symbol{}, b_to_update, tools.Trade{}, 2, time.Time{}, time.Time{})
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

		if symbol.State != "standby" {
			trade_decision(conf, symbol, calc_bid, open_trades)
		}

		log.Info("Symbol ", symbol.Id, " ( ", symbol.Name, " ) OK | duration : ", dDiff, " | next retrieve in : ", dStepTempo)

		time.Sleep(dStepTempo)
	}
}

func trade_decision(conf *config.Config, symbol tools.Symbol, calc_bid []tools.Bid, open_trades map[int]tools.Trade) {

	if calc_bid == nil || len(calc_bid) == 0 {
		close_all_trade(conf, open_trades, "calc_bid == nil or 0")
	}

	var last_bid = calc_bid[len(calc_bid)-1]

	//log.Info(symbol.Name, " -> ", last_bid.Id, " | ", last_bid.Bid_at_s, " | ", last_bid.Calculations)

	var ok bool
	var sma_6, sma_12 float64

	if sma_6, ok = last_bid.Calculations["sma_6"]; !ok {
		close_all_trade(conf, open_trades, "sma_6 is nil")
	}

	if sma_12, ok = last_bid.Calculations["sma_12"]; !ok {
		close_all_trade(conf, open_trades, "sma_12 is nil")
	}

	var diff_sma_12_6 = sma_12 - sma_6

	if diff_sma_12_6 >= 20 {
		var trade = tools.Trade{0, symbol, 1, symbol.Lot_min_size, "bid", ""}
		open_trad(conf, trade, open_trades)
		close_ask_trade(conf, open_trades)
	} else if diff_sma_12_6 <= -20 {
		var trade = tools.Trade{0, symbol, 0, symbol.Lot_min_size, "ask", ""}
		open_trad(conf, trade, open_trades)
		close_bid_trade(conf, open_trades)
	}
}

func open_trad(conf *config.Config, trade tools.Trade, open_trades map[int]tools.Trade) {
	var err error
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	res, err = api.Api_request(*conf, tools.Symbol{}, tools.Bid{}, trade, 5, time.Time{}, time.Time{})
	if err != nil {
		switch {
		case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
			log.Error(err.Error())
			return
		default:
			log.FatalError(err)
			return
		}
	}

	if len(res.Trades) != 1 {
		log.Error("len(res.Trades) != 1 -> ", len(res.Trades))
		return
	}

	open_trades[res.Trades[0].Id] = res.Trades[0]
}

func get_opened_trades(conf *config.Config, open_trades map[int]tools.Trade) {
	var err error
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	res, err = api.Api_request(*conf, tools.Symbol{}, tools.Bid{}, tools.Trade{}, 7, time.Time{}, time.Time{})
	if err != nil {
		switch {
		case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
			log.Error(err.Error())
			return
		default:
			log.FatalError(err)
			return
		}
	}

	if len(res.Trades) != 0 {
		for _, t := range res.Trades {
			open_trades[t.Id] = t
		}
	}
}

func close_trade(conf *config.Config, trade_to_close tools.Trade, open_trades map[int]tools.Trade) {
	var err error
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	res, err = api.Api_request(*conf, tools.Symbol{}, tools.Bid{}, trade_to_close, 6, time.Time{}, time.Time{})
	if err != nil {
		switch {
		case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
			log.Error(err.Error())
			return
		default:
			log.FatalError(err)
			return
		}
	}
	delete(open_trades, trade_to_close.Id)
}

func close_all_trade(conf *config.Config, open_trades map[int]tools.Trade, reason string) {
	for _, trade_to_close := range open_trades {
		trade_to_close.Closed_reason = reason
		close_trade(conf, trade_to_close, open_trades)
	}
}

func close_ask_trade(conf *config.Config, open_trades map[int]tools.Trade) {
	for _, trade_to_close := range open_trades {
		if trade_to_close.Trade_type == 0 {
			trade_to_close.Closed_reason = "bid"
			close_trade(conf, trade_to_close, open_trades)
		}
	}
}

func close_bid_trade(conf *config.Config, open_trades map[int]tools.Trade) {
	for _, trade_to_close := range open_trades {
		if trade_to_close.Trade_type == 1 {
			trade_to_close.Closed_reason = "ask"
			close_trade(conf, trade_to_close, open_trades)
		}
	}
}

func rattrap_calc_symbol(conf *config.Config, symbol tools.Symbol) {

	var err error
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	res, err = api.Api_request(*conf, symbol, tools.Bid{}, tools.Trade{}, 3, conf.API.From, conf.API.To)

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
		res, err = api.Api_request(*conf, tools.Symbol{}, b_to_update, tools.Trade{}, 2, time.Time{}, time.Time{})
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
