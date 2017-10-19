package exec

import (
	"time"
	"../log"
	"../tools"
	"../config/utils"
	"errors"
	"../db"
	"./api"
)

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

	log.Info("Symbol ", symbol.Id, "	( ", symbol.Name, " ) OK | duration : ", dDiff, "	| next retrieve in : ", dStepTempo)

	time.Sleep(dStepTempo)
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
