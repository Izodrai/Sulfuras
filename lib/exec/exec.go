package exec

import (
	"../log"
	//"encoding/json"
	"time"
	"./api"
	"../config/utils"
	"../tools"
	//"./calculate"
	//"./decision"
	//"errors"
)

func ExecNotInactivSymbols(api_c *utils.API) error {

	var err error
	var open_trades = make(map[int]tools.Trade)

	ch_req_to_exec := make(chan tools.Request)


	if err = api.GetOpenedTrades(api_c, open_trades); err != nil {
		return err
	}

	go requester(ch_req_to_exec, api_c)

	for i, symbol := range api_c.Symbols {
		// TODELETE
		if i > 0 {
			break
		}
		////////////
		go retrieveDataForSymbol(api_c, symbol, ch_req_to_exec)
	}



	for {
		time.Sleep(24 * time.Hour)
	}

	return nil
}



func retrieveDataForSymbol(api_c *utils.API, symbol tools.Symbol, ch_req_to_exec chan tools.Request){

	for {
		var tNow = time.Now()

		if !allowedToExe(api_c, tNow) {
			continue
		}

		var req tools.Request
		req.Resp = make(chan tools.Response)
		req.Symbol = symbol
		req.URL_request = api.RequestFeedSymbol(api_c, symbol)
		ch_req_to_exec <- req

		resp := <- req.Resp

		if resp.Error != nil {
			switch {
			case resp.Error.Error() == "Unable to connect to any of the specified MySQL hosts.":
				log.Error("For : ", symbol.Name, " - ", resp.Error.Error())
				continue
			case resp.Error.Error() == "No data to retrieve in this range":
				log.Error("For : ", symbol.Name, " - ", resp.Error.Error())
				continue
			default:
				log.FatalError(resp.Error)
				continue
			}
		}
		log.Info(resp)

		untilNextStep(api_c, tNow, symbol)
	}

}

func requester(ch_req_to_exec chan tools.Request, api_c *utils.API) {

	var closed bool
	var buffer []tools.Request

	go func () {
		for r := range ch_req_to_exec {
			buffer = append(buffer, r)
		}
		closed = true
	}()

	for !closed || len(buffer) != 0 {
		if len(buffer) > 0 {
			req := buffer[0]

			buffer = append(buffer[:0], buffer[1:]...)

			req.Resp <- api.Request(req.URL_request, api_c)

			close(req.Resp)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

/*
func ExecSymbol(api_c *utils.API, symbol tools.Symbol) {

	var err error
	var dStep = 2*time.Minute + 30*time.Second
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	var open_trades = make(map[int]tools.Trade)

	if err = api.GetOpenedTrades(api_c, open_trades); err != nil {
		//return err
	}

	for {
		var tNow = time.Now()

		if !allowedToExe(api_c, tNow) {
			continue
		}

		res, err = api.RequestFeedSymbol(api_c, symbol, tools.Bid{}, tools.Trade{}, time.Time{}, time.Time{})

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

		calc_bid := calculate.CalculateBids(api_c, res_bids)
		for _, b_to_update := range calc_bid {
			res, err = api.RequestSetCalculation(api_c, tools.Symbol{}, b_to_update, tools.Trade{}, time.Time{}, time.Time{})
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

		untilNextStep(api_c, tNow, symbol)

	}
}
*/





func allowedToExe(api_c *utils.API, tNow time.Time) bool{
	if per, ok := api_c.RetrievePeriode[tNow.Weekday()]; ok {
		if per.Deactivated {
			log.Info(tNow.Weekday(), " is deactivated")
			y, m, d := tNow.Date()
			tTomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, tNow.Location())
			log.Info("tNow (", tNow.Format("2006-01-02 15:04:05"), "), tTomorrow (", tTomorrow.Format("2006-01-02 15:04:05"), ")")
			log.Info("Sleep ", tTomorrow.Sub(tNow).String(), " before tomorrow")
			time.Sleep(tTomorrow.Sub(tNow))
			return false
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
				return false
			}

			if tNow.After(tEnd) {
				log.Info(tNow.Weekday(), " is limited")
				log.Info("tNow (", tNow.Format("2006-01-02 15:04:05"), ") after tEnd (", tEnd.Format("2006-01-02 15:04:05"), "), tTomorrow (", tTomorrow.Format("2006-01-02 15:04:05"), ")")
				log.Info("Sleep ", tTomorrow.Sub(tNow).String(), " before tomorrow")
				time.Sleep(tTomorrow.Sub(tNow))
				return false
			}
		}
	}
	return true
}

func untilNextStep(api_c *utils.API, tNow time.Time, symbol tools.Symbol) {
	dDiff := time.Now().Sub(tNow)

	var dStepTempo time.Duration

	if dDiff >= api_c.StepRetrieve/2 {
		dStepTempo = api_c.StepRetrieve
	} else {
		dStepTempo = api_c.StepRetrieve - dDiff
	}

	log.Info("Symbol ", symbol.Id, " ( ", symbol.Name, " ) OK | duration : ", dDiff, " | next retrieve in : ", dStepTempo)

	time.Sleep(dStepTempo)
}



