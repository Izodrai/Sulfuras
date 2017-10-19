package exec

import (
	"../config/utils"
	"../db"
	"../log"
	"../tools"
	"./calculate"
)

func calc(api_c *utils.API, resp_bids []tools.Bid, ch_bid chan tools.Bid, upd_bids map[int]tools.Bid) {

	var err error

	calc_bid := calculate.CalculateBids(api_c, resp_bids, upd_bids)

	for _, b_to_update := range calc_bid {

		log.Info(b_to_update)

		if err = db.InsertOrUpdateBid(api_c, &b_to_update); err != nil {
			log.Error(err)
			continue
		}

		ch_bid <- b_to_update
	}
}
