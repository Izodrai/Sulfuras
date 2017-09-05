package calculate

import (
	"strconv"

	"../config"
	"../tools"
)

func calc_ema(conf *config.Config, res_bids []tools.Bid) []tools.Bid {

	var calc_bids []tools.Bid
	var ema_conf = make(map[int][]float64)

	for _, co_ema := range conf.API.Calculations.EMA {
		ema_conf[co_ema] = []float64{}
	}

	for i, res_b := range res_bids {

		var b tools.Bid
		b.Id = res_b.Id
		b.Symbol = res_b.Symbol
		b.Bid_at_s = res_b.Bid_at_s
		b.Last_bid = res_b.Last_bid
		b.Calculations_s = res_b.Calculations_s

		if res_b.Calculations != nil {
			b.Calculations = res_b.Calculations
		} else {
			b.Calculations = make(map[string]float64)
		}

		calc_bids = append(calc_bids, res_b)

		for co_ema, _ := range ema_conf {
			ema_conf[co_ema] = append(ema_conf[co_ema], res_b.Last_bid)

			if len(ema_conf[co_ema]) > co_ema {
				ema_conf[co_ema] = append(ema_conf[co_ema][:0], ema_conf[co_ema][1:]...)
			} else if len(ema_conf[co_ema]) < co_ema {
				continue
			}

			var ma float64
			for _, last_b := range ema_conf[co_ema] {
				ma = ma + last_b
			}
			ma = ma / float64(co_ema)

			b.Calculations["ema_"+strconv.Itoa(co_ema)] = ma
		}

		calc_bids[i] = b
	}

	return calc_bids
}
