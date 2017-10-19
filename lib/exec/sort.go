package exec

import (
	"../log"
	"../tools"
	"encoding/json"
)

func marshal(bids []tools.Bid) []tools.Bid{

	var err error
	var resp_bids []tools.Bid

	for _, res_b := range bids {
		if res_b.Calculations_s != "{}" {
			err = json.Unmarshal([]byte(res_b.Calculations_s), &res_b.Calculations)
			if err != nil {
				log.Error("For : ", res_b, " - ", err.Error())
				continue
			}
		}
		resp_bids = append(resp_bids, res_b)
	}
	return resp_bids
}