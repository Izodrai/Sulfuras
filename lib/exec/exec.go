package exec

import (
	"../log"
	"time"
	"./api"
	"../config/utils"
	"../tools"
	"./decision"
	"strconv"
)

func ExecNotInactivSymbols(api_c *utils.API, bids map[int]map[int]tools.Bid) error {

	var err error
	var open_trades = make(map[int]tools.Trade)

	ch_req_to_exec := make(chan tools.Request)
	ch_bid := make(chan tools.Bid)

	if err = api.GetOpenedTrades(api_c, open_trades); err != nil {
		return err
	}

	go requester(ch_req_to_exec, api_c)

	go savedBids(ch_bid, bids)

	for _, symbol := range api_c.Symbols_t {
		log.GreenInfo(symbol.Name + " ("+strconv.Itoa(symbol.Id)+")	- running...")

		if err = loadLastBidsForSymbol(api_c, symbol, ch_bid); err != nil {
			log.Error("Cannot load data for : ", symbol.Name, " - ", err.Error())
			continue
		}

		go dataRetrieve(api_c, symbol, ch_req_to_exec, ch_bid)
	}

	go decision.DecisionMaker(api_c, ch_req_to_exec, bids)

	for {
		time.Sleep(24 * time.Hour)
	}

	return nil
}

func dataRetrieve(api_c *utils.API, symbol tools.Symbol, ch_req_to_exec chan tools.Request, ch_bid chan tools.Bid){

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

		calc(api_c, marshal(resp_feed.Bids), ch_bid)

		untilNextStep(api_c, tNow, symbol)
	}
}





