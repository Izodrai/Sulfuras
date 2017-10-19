package calculate

import (
	"strconv"

	"../../config/utils"
	"../../tools"
)

func calcSma(api *utils.API, res_bids []tools.Bid) []tools.Bid {

	var calc_bids []tools.Bid
	var sma_conf = make(map[int][]float64)

	for _, co_sma := range api.Calculations.Sma {
		sma_conf[co_sma] = []float64{}
	}

	for i, res_b := range res_bids {

		var b tools.Bid
		b.Id = res_b.Id
		b.Symbol = res_b.Symbol
		b.Bid_at_s = res_b.Bid_at_s
		b.Bid_at = res_b.Bid_at
		b.Last_bid = res_b.Last_bid
		b.Calculations_s = res_b.Calculations_s

		if res_b.Calculations != nil {
			b.Calculations = res_b.Calculations
		} else {
			b.Calculations = make(map[string]float64)
		}

		calc_bids = append(calc_bids, res_b)

		for co_sma, _ := range sma_conf {
			sma_conf[co_sma] = append(sma_conf[co_sma], res_b.Last_bid)

			if len(sma_conf[co_sma]) > co_sma {
				sma_conf[co_sma] = append(sma_conf[co_sma][:0], sma_conf[co_sma][1:]...)
			} else if len(sma_conf[co_sma]) < co_sma {
				continue
			}

			var ma float64
			for _, last_b := range sma_conf[co_sma] {
				ma = ma + last_b
			}
			ma = ma / float64(co_sma)

			b.Calculations["sma_"+strconv.Itoa(co_sma)] = ma
		}

		calc_bids[i] = b
	}

	return calc_bids
}
