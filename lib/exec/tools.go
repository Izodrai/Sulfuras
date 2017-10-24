package exec

import (
	"../config/utils"
	"../db"
	"../log"
	"../tools"
	"./api"
	"errors"
	"time"
	"encoding/json"
	"strconv"
)

func savedBids(ch_bid chan tools.Bid, bids map[int]tools.SavedBids) {

	for b := range ch_bid {
		var ok bool
		var bids_of_s tools.SavedBids

		_ = b.Feed(b.Symbol)

		if bids_of_s, ok = bids[b.Symbol.Id]; !ok {
			bids_of_s.ById = make(map[int]tools.Bid)
			bids_of_s.ByDate = make(map[time.Time]tools.Bid)
		}

		bids_of_s.AddBid(b)
		bids[b.Symbol.Id] = bids_of_s
	}
}

func requester(ch_req_to_exec chan tools.Request, api_c *utils.API) {

	var closed bool
	var buffer []tools.Request

	go func() {
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

func allowedToExe(api_c *utils.API, tNow time.Time) bool {
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

func untilNextStep(api_c *utils.API, tTime []time.Time, symbol tools.Symbol) {

	dDiff := time.Now().Sub(tTime[0])

	var dStepTempo time.Duration

	if dDiff >= api_c.StepRetrieve/2 {
		dStepTempo = api_c.StepRetrieve
	} else {
		dStepTempo = api_c.StepRetrieve - dDiff
	}

	// log.Info("Symbol ", symbol.Id, "	( ", symbol.Name, " ) OK | duration : ", dDiff, "	| next retrieve in : ", dStepTempo)

	var tot, s1_2, s2_3, s3_4 string

	tot = strconv.FormatFloat(dDiff.Seconds(), 'f', 3, 64)
	s1_2 = strconv.FormatFloat(tTime[1].Sub(tTime[0]).Seconds(), 'f', 3, 64)
	s2_3 = strconv.FormatFloat(tTime[2].Sub(tTime[1]).Seconds(), 'f', 3, 64)
	s3_4 = strconv.FormatFloat(tTime[3].Sub(tTime[2]).Seconds(), 'f', 3, 64)

	var s = "-> "+ symbol.Name+ " ("+strconv.Itoa(symbol.Id)+") "

	if symbol.Id < 10 {
		s += " "
	}

	s += "OK, total dur: "+tot+"s "

	if dDiff.Seconds() < 10 {
		s += " "
	}

	s+= "| resp: "+s1_2+"s "
	if tTime[1].Sub(tTime[0]).Seconds() < 10 {
		s += " "
	}

	s+= "| feed: "+s2_3+"s "
	if tTime[2].Sub(tTime[1]).Seconds() < 10 {
		s += " "
	}

	s+= "| calc: "+s3_4+"s "
	if tTime[3].Sub(tTime[2]).Seconds() < 10 {
		s += " "
	}
	s+= "next retrieve in : " + strconv.FormatFloat(dStepTempo.Seconds(), 'f', 3, 64)+"s"

	log.Info(s)

	time.Sleep(dStepTempo)
}

func loadLastBidsForSymbol(api_c *utils.API, symbol tools.Symbol, ch_bid chan tools.Bid) error {
	var err error
	var bs, bsf []tools.Bid

	if err = db.LoadLastBidsForSymbol(api_c, symbol, &bs); err != nil {
		return errors.New("Cannot load data for : " + symbol.Name + " - " + err.Error())
	}

	for _,b := range bs {
		if err = b.Feed(symbol); err != nil {
			return err
		}
		bsf = append(bsf, b)
	}

	if err = initCheckCalc(api_c, &bsf, ch_bid); err != nil {
		return err
	}

	return nil
}

func marshal(bids []tools.Bid) []tools.Bid {

	var err error
	var resp_bids []tools.Bid

	for _, res_b := range bids {
		if res_b.Calculations_s != "{}" {
			err = json.Unmarshal([]byte(res_b.Calculations_s), &res_b.Calculations)
			if err != nil {
				log.Error("For : ", res_b, " - ", err.Error())
				continue
			}
		}
		resp_bids = append(resp_bids, res_b)
	}
	return resp_bids
}