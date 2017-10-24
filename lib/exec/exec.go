package exec

import (
	"../config/utils"
	"../log"
	"../tools"
	"./api"
	"./decision"
	"strconv"
	"time"
	//"errors"
	//"../db"
	"errors"
	"os"
)

func RattrapNotInactivSymbols(api_c *utils.API, bids map[int]tools.SavedBids) error {

	/*
	var err error
	ch_bid := make(chan tools.Bid)
	var upd_bids = make(map[int]tools.Bid)

	go savedBids(ch_bid, bids)

	for _, symbol := range api_c.Symbols_t {
		log.GreenInfo(symbol.Name + " (" + strconv.Itoa(symbol.Id) + ")	- running...")

		var bs []tools.Bid

		if err = db.LoadLastBidsForSymbol(api_c, symbol, &bs); err != nil {
			return errors.New("Cannot load data for : " + symbol.Name + " - " + err.Error())
		}

		calc(api_c, marshal(bs), ch_bid, upd_bids)
	}
	*/

	return errors.New("not yet implemented")
}

func ExecNotInactivSymbols(api_c *utils.API, bids map[int]tools.SavedBids) error {

	var err error
	var open_trades = make(map[int]tools.Trade)

	ch_req_to_exec := make(chan tools.Request)
	ch_bid := make(chan tools.Bid)
	ch_symbol := make(chan tools.Symbol)

	if err = api.GetOpenedTrades(api_c, open_trades); err != nil {
		return err
	}

	go requester(ch_req_to_exec, api_c)

	go savedBids(ch_bid, bids)

	for _, symbol := range api_c.Symbols_t {

		/*if symbol.Id != 1 {
			continue
		}*/

		log.WhiteInfo(symbol.Name + " (" + strconv.Itoa(symbol.Id) + ")	- loading data...")

		if err = loadLastBidsForSymbol(api_c, symbol, ch_bid); err != nil {
			log.Error("Cannot load data for : ", symbol.Name, " - ", err.Error())
			continue
		}

		log.GreenInfo(symbol.Name + " (" + strconv.Itoa(symbol.Id) + ")	- running...")

		go dataRetrieve(api_c, symbol, ch_req_to_exec, ch_bid, bids[symbol.Id], ch_symbol)
	}

	go decision.DecisionMaker(api_c, ch_req_to_exec, bids, ch_symbol)

	return nil
}

func dataRetrieve(api_c *utils.API, symbol tools.Symbol, ch_req_to_exec chan tools.Request, ch_bid chan tools.Bid, saved_bids tools.SavedBids, ch_symbol chan tools.Symbol) {

	for {
		var tTime = []time.Time{}
		var tNow = time.Now()

		////////
		// 1 init
		tTime = append(tTime, tNow)
		//
		////////

		if !allowedToExe(api_c, tNow) {
			continue
		}

		var req_feed tools.Request
		req_feed.Resp = make(chan tools.Response)
		req_feed.Symbol = symbol

		req_feed.URL_request = api.RequestFeedSymbol(api_c, symbol, saved_bids.LastDate.Add(-24*time.Hour))

		ch_req_to_exec <- req_feed

		resp_feed := <-req_feed.Resp

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
			case resp_feed.Error.Error() == "Azure app disconected":
				log.Error("For : ", symbol.Name, " - ", resp_feed.Error.Error(), " - quota exceded ?")

				// TODO SEND EMAIL

				time.Sleep(api_c.StepRetrieve * 2)
				continue
			default:
				log.FatalError(resp_feed.Error)
				time.Sleep(api_c.StepRetrieve)
				continue
			}
		}

		////////
		// 2 resp
		tTime = append(tTime, time.Now())
		//
		////////

		var resp_bids  []tools.Bid
		var upd_bids = make(map[time.Time]interface{})

		for _, xtb_b := range resp_feed.Bids {

			xtb_b.Feed(symbol)

			if mysql_b, ok := saved_bids.ByDate[xtb_b.Bid_at]; ok {
				if xtb_b.Last_bid != mysql_b.Last_bid {
					upd_bids[xtb_b.Bid_at] = nil
					resp_bids = append(resp_bids, xtb_b)
				} else {
					resp_bids = append(resp_bids, mysql_b)
				}
			} else {
				upd_bids[xtb_b.Bid_at] = nil
				resp_bids = append(resp_bids, xtb_b)
			}

		}

		////////
		// 3 feed
		tTime = append(tTime, time.Now())
		//
		////////

		if err := checkCalc(api_c, &resp_bids, upd_bids, ch_bid); err != nil {
			log.FatalError(err)
			os.Exit(0)
		}

		////////
		// 4 calc
		tTime = append(tTime, time.Now())
		//
		////////

		//ch_symbol <- symbol

		untilNextStep(api_c, tTime, symbol)
	}
}
