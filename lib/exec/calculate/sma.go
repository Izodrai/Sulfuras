package calculate

import (
	gma "github.com/RobinUS2/golang-moving-average"
	"../../config/utils"
	"../../tools"
	"strconv"
)

func smaCalc(api *utils.API, org_bids, chng_bids *[]tools.Bid) {

	var sma_conf = make(map[int]*gma.MovingAverage)

	for _, co_sma := range api.Calculations.Sma {
		sma_conf[co_sma] = gma.New(co_sma)
	}

	var new_org_bids []tools.Bid

	for _, org_b := range *org_bids {
		var change bool
		var new_b = org_b

		if new_b.Calculations == nil {
			new_b.Calculations = make(map[string]float64)
		}

		for co_sma,ma := range sma_conf {

			var ib = "sma_"+strconv.Itoa(co_sma)

			ma.Add(new_b.Last_bid)

			if calc, ok := org_b.Calculations[ib]; ok {
				if calc != ma.Avg() {
					change = true
				}
			} else {
				change = true
			}

			new_b.Calculations[ib] = ma.Avg()
		}

		if change {
			*chng_bids = append(*chng_bids, org_b)
		}

		new_org_bids = append(new_org_bids, org_b)
	}

	*org_bids = new_org_bids
}