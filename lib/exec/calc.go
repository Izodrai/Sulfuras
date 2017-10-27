package exec

import (
	"../config/utils"
	"../db"
	"../tools"
	"./calculate"
	"../log"
)

func initCheckCalc(api_c *utils.API, org_bids *[]tools.Bid, ch_bid chan tools.Bid) error {

	var err error
	var chng_bids []tools.Bid

	calculate.Calc(api_c, org_bids, &chng_bids)
	for _, b_to_update := range chng_bids {
		if err = db.InsertOrUpdateBid(api_c, &b_to_update); err != nil {
			return err
		}
	}

	for _, b := range *org_bids {
		ch_bid <- b
	}

	return nil
}

func checkCalc(api_c *utils.API, org_bids *[]tools.Bid, upd_bids map[int64]interface{}, ch_bid chan tools.Bid) error {

	var chng_bids []tools.Bid

	calculate.Calc(api_c, org_bids, &chng_bids)

	for _, chng_b := range chng_bids {
		if _, ok := upd_bids[chng_b.Bid_at_ts]; ok {
			//log.Info(chng_b)

			if err := db.InsertOrUpdateBid(api_c, &chng_b); err != nil {
				log.Error(err)
				return err
			}
			ch_bid <- chng_b
		}
	}

	return nil
}