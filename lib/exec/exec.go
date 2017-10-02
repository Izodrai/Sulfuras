package exec

import (
	"../log"
	"time"
	"./api"
	"../config/utils"
	"../tools"
	"./calculate"
	"encoding/json"
	"strconv"
)

func ExecNotInactivSymbols(api_c *utils.API, bids map[int]map[int]tools.Bid) error {

	//var err error
	//var open_trades = make(map[int]tools.Trade)

	ch_req_to_exec := make(chan tools.Request)
	ch_bid := make(chan tools.Bid)

	/*if err = api.GetOpenedTrades(api_c, open_trades); err != nil {
		return err
	}*/

	go requester(ch_req_to_exec, api_c)

	go savedBids(ch_bid, bids)

	for _, symbol := range api_c.Symbols {
		// TODELETE
		if symbol.Id != 1 && symbol.Id != 41 {
			continue
		}
		log.Info(symbol.Name + " ("+strconv.Itoa(symbol.Id)+") - process to execute")
		////////////
		go retrieveDataForSymbol(api_c, symbol, ch_req_to_exec, ch_bid, bids)
	}

	for {
		time.Sleep(24 * time.Hour)
	}

	return nil
}

func savedBids(ch_bid chan tools.Bid, bids map[int]map[int]tools.Bid) {

	for b := range ch_bid {
		var ok bool
		var bids_of_s map[int]tools.Bid

		if bids_of_s, ok = bids[b.Symbol.Id]; !ok {
			bids_of_s = make(map[int]tools.Bid)
		}

		if existing_b, ok := bids_of_s[b.Id]; !ok {
			bids_of_s[b.Id] = b
			bids[b.Symbol.Id] = bids_of_s
		} else {
			if b.Last_bid != existing_b.Last_bid  {
				bids_of_s[b.Id] = b
				bids[b.Symbol.Id] = bids_of_s
			}
		}
	}
}

func retrieveDataForSymbol(api_c *utils.API, symbol tools.Symbol, ch_req_to_exec chan tools.Request, ch_bid chan tools.Bid, bids map[int]map[int]tools.Bid){

	var err error

reload:
	var req_get tools.Request
	req_get.Resp = make(chan tools.Response)
	req_get.Symbol = symbol
	req_get.URL_request = api.RequestGetDataForSymbol(api_c, symbol, api_c.From, api_c.To)

	ch_req_to_exec <- req_get

	resp_get := <- req_get.Resp

	if resp_get.Error != nil {
		switch {
		case resp_get.Error.Error() == "Unable to connect to any of the specified MySQL hosts.":
			log.Error("For : ", symbol.Name, " - ", resp_get.Error.Error())
			time.Sleep(api_c.StepRetrieve)
			goto reload
		case resp_get.Error.Error() == "No data to retrieve in this range":
			log.Error("For : ", symbol.Name, " - ", resp_get.Error.Error())
			time.Sleep(api_c.StepRetrieve)
			goto reload
		default:
			log.FatalError(resp_get.Error)
			time.Sleep(api_c.StepRetrieve)
			goto reload
		}
	}

	for _, res_b := range resp_get.Bids {
		ch_bid <- res_b
	}

	////////////////////////////////////////

	for {
		var tNow = time.Now()

		if !allowedToExe(api_c, tNow) {
			continue
		}

		var req_feed tools.Request
		req_feed.Resp = make(chan tools.Response)
		req_feed.Symbol = symbol
		req_feed.URL_request = api.RequestFeedSymbol(api_c, symbol)

		ch_req_to_exec <- req_feed

		resp_feed := <- req_feed.Resp

		if resp_feed.Error != nil {
			switch {
			case resp_feed.Error.Error() == "Unable to connect to any of the specified MySQL hosts.":
				log.Error("For : ", symbol.Name, " - ", resp_feed.Error.Error())
				time.Sleep(api_c.StepRetrieve)
				continue
			case resp_feed.Error.Error() == "No data to retrieve in this range":
				log.Error("For : ", symbol.Name, " - ", resp_feed.Error.Error())
				time.Sleep(api_c.StepRetrieve)
				continue
			default:
				log.FatalError(resp_feed.Error)
				time.Sleep(api_c.StepRetrieve)
				continue
			}
		}

		var resp_bids []tools.Bid

		for _, res_b := range resp_feed.Bids {
			if res_b.Calculations_s != "{}" {
				err = json.Unmarshal([]byte(res_b.Calculations_s), &res_b.Calculations)
				if err != nil {
					log.Error("For : ", res_b, " - ", err.Error())
					continue
				}
			}
			resp_bids = append(resp_bids, res_b)
		}

		calc_bid := calculate.CalculateBids(api_c, resp_bids)

		for _, b_to_update := range calc_bid {
			var req_update tools.Request
			req_update.Resp = make(chan tools.Response)
			req_update.Symbol = symbol
			req_update.URL_request = api.RequestSetCalculation(api_c, b_to_update)

			ch_req_to_exec <- req_update

			resp_update := <- req_update.Resp

			if resp_update.Error != nil {
				switch {
				case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
					log.Error(err.Error())
					continue
				default:
					log.FatalError(err)
					continue
				}
			}

			ch_bid <- b_to_update
		}

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



