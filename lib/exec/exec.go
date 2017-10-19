package exec

import (
	"../config/utils"
	"../log"
	"../tools"
	"./api"
	"./decision"
	"os"
	"strconv"
	"time"
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

		if symbol.Id != 1 {
			continue
		}
		log.GreenInfo(symbol.Name + " (" + strconv.Itoa(symbol.Id) + ")	- running...")

		if err = loadLastBidsForSymbol(api_c, symbol, ch_bid); err != nil {
			log.Error("Cannot load data for : ", symbol.Name, " - ", err.Error())
			continue
		}

		go dataRetrieve(api_c, symbol, ch_req_to_exec, ch_bid, bids[symbol.Id])
	}

	go decision.DecisionMaker(api_c, ch_req_to_exec, bids)

	for {
		time.Sleep(24 * time.Hour)
	}

	return nil
}

func dataRetrieve(api_c *utils.API, symbol tools.Symbol, ch_req_to_exec chan tools.Request, ch_bid chan tools.Bid, saved_bids map[int]tools.Bid) {

	for {
		var tNow = time.Now()

		if !allowedToExe(api_c, tNow) {
			continue
		}

		var req_feed tools.Request
		req_feed.Resp = make(chan tools.Response)
		req_feed.Symbol = symbol

		var dt_last time.Time
		var min_id, max_id int

		for _, b := range saved_bids {
			if dt_last.Before(b.Bid_at) {
				dt_last = b.Bid_at
			}
			if b.Id > max_id {
				max_id = b.Id
			}
			if b.Id < min_id {
				min_id = b.Id
			}
		}

		req_feed.URL_request = api.RequestFeedSymbol(api_c, symbol, dt_last.Add(-3*time.Hour))

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

		var resp_bids, bids []tools.Bid
		var upd_bids = make(map[int]tools.Bid)

		for _, xtb_b := range resp_feed.Bids {
			xtb_b.Feed(symbol)
			resp_bids = append(resp_bids, xtb_b)
		}

		for i := min_id; i <= max_id; i++ {

			if mysql_b, ok := saved_bids[i]; ok {
				for _, xtb_b := range resp_bids {
					if xtb_b.Bid_at == mysql_b.Bid_at {
						if mysql_b.Last_bid != xtb_b.Last_bid {
							mysql_b.Last_bid = xtb_b.Last_bid
							upd_bids[mysql_b.Id] = mysql_b
						}
					}
				}
				bids = append(bids, mysql_b)
			}
		}

		for _, xtb_b := range resp_bids {

			var exist bool

			for _, mysql_b := range saved_bids {
				if xtb_b.Bid_at == mysql_b.Bid_at {
					exist = true
					break
				}
			}

			if exist {
				continue
			}
			bids = append(bids, xtb_b)
		}

		calc(api_c, marshal(bids), ch_bid, upd_bids)

		os.Exit(0)

		untilNextStep(api_c, tNow, symbol)
	}
}
