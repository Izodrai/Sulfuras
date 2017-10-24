package calculate

import (
	gma "github.com/RobinUS2/golang-moving-average"
	"../../config/utils"
	"../../tools"
	"strconv"
)

func calcEma(api *utils.API, org_bids, chng_bids *[]tools.Bid) {

	var ema_conf= make(map[int]*gma.MovingAverage)

	for _, co_ema := range api.Calculations.Ema {
		ema_conf[co_ema] = gma.New(co_ema)
	}

	for _, org_b := range *org_bids {
		var change bool

		var calc_b = org_b
		calc_b.Calculations = make(map[string]float64)

		for co_ema,ma := range ema_conf {

			var ib = "ema_"+strconv.Itoa(co_ema)

			ma.Add(org_b.Last_bid)

			if calc, ok := org_b.Calculations[ib]; ok {
				if calc != ma.Avg() {
					change = true
				}
			} else {
				change = true
			}

			calc_b.Calculations[ib] = ma.Avg()
		}

		if change {
			*chng_bids = append(*chng_bids, calc_b)
		}
	}
}
