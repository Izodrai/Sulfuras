package exec

import (
	"../log"
	"time"
	"./api"
	"../config/utils"
	"../tools"
	"../db"
	"./calculate"
	"encoding/json"
	"strconv"
	"errors"
)

func ExecNotInactivSymbols(api_c *utils.API, bids map[int]map[int]tools.Bid) error {

	var err error
	//var open_trades = make(map[int]tools.Trade)

	ch_req_to_exec := make(chan tools.Request)
	ch_bid := make(chan tools.Bid)

	/*if err = api.GetOpenedTrades(api_c, open_trades); err != nil {
		return err
	}*/

	go requester(ch_req_to_exec, api_c)

	go savedBids(ch_bid, bids)

	for _, symbol := range api_c.Symbols {
		log.Info(symbol.Name + " ("+strconv.Itoa(symbol.Id)+") - process to execute ( retrieve data )")

		if err = loadLastBidsForSymbol(api_c, symbol, ch_bid); err != nil {
			log.Error("Cannot load data for : ", symbol.Name, " - ", err.Error())
			continue
		}

		go run(api_c, symbol, ch_req_to_exec, ch_bid)
	}

	for {
		time.Sleep(24 * time.Hour)
	}

	return nil
}

func loadLastBidsForSymbol(api_c *utils.API, symbol tools.Symbol, ch_bid chan tools.Bid) error {
	var err error
	var bs []tools.Bid

	if err = db.LoadLastBidsForSymbol(api_c, symbol, &bs); err != nil {
		return errors.New("Cannot load data for : " + symbol.Name + " - " + err.Error())
	}

	for _,b := range bs {
		ch_bid <- b
	}
	return nil
}

func savedBids(ch_bid chan tools.Bid, bids map[int]map[int]tools.Bid) {

	for b := range ch_bid {
		var ok bool
		var bids_of_s map[int]tools.Bid

		_ = b.ParseAPITime()
		_ = b.UnmarshalCalculation()

		if bids_of_s, ok = bids[b.Symbol.Id]; !ok {
			bids_of_s = make(map[int]tools.Bid)
		}

		if existing_b, ok := bids_of_s[b.Id]; !ok {
			bids_of_s[b.Id] = b
			bids[b.Symbol.Id] = bids_of_s
		} else {
			if b.Last_bid != existing_b.Last_bid || b.Calculations_s != existing_b.Calculations_s {
				bids_of_s[b.Id] = b
				bids[b.Symbol.Id] = bids_of_s
			}
		}
	}
}

func run(api_c *utils.API, symbol tools.Symbol, ch_req_to_exec chan tools.Request, ch_bid chan tools.Bid){

	var err error

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

			if err = db.UpdateCalculation(api_c, &b_to_update); err != nil {
				log.Error(err)
				continue
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
		time.Sleep(100 * time.Millisecond)
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



