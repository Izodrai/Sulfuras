package calculate

import (
	"../../config/utils"
	"../../tools"
)

func CalculateBids(api *utils.API, res_bids []tools.Bid) []tools.Bid {

	var calc_bids, calc_1 []tools.Bid

	calc_1 = calcSma(api, res_bids)

	calc_bids = calcEma(api, calc_1)

	return sortCalc( res_bids, calc_bids)
}

func sortCalc( res_bids, calc_bids []tools.Bid) []tools.Bid {

	var bids_to_update []tools.Bid

	for i, calc_b := range calc_bids {

		var b = res_bids[i]
		b.Calculations = map[string]float64{}

		var diff = false

		for t, val := range res_bids[i].Calculations {
			b.Calculations[t] = val
		}

		for t, calc_val := range calc_b.Calculations {
			if res_val, ok := b.Calculations[t]; ok {
				if calc_val == res_val {
					continue
				}
			}
			b.Calculations[t] = calc_val
			diff = true
		}

		if !diff {
			continue
		}

		bids_to_update = append(bids_to_update, b)
	}

	return bids_to_update
}
