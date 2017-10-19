package exec

import (
	"../db"
	"../log"
	"../tools"
	"./calculate"
	"../config/utils"
)

func calc(api_c *utils.API, resp_bids []tools.Bid, ch_bid chan tools.Bid) {
	var err error
	calc_bid := calculate.CalculateBids(api_c, resp_bids)

	for _, b_to_update := range calc_bid {

		if err = db.UpdateCalculation(api_c, &b_to_update); err != nil {
			log.Error(err)
			continue
		}

		ch_bid <- b_to_update
	}
}
