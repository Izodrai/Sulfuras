package exec


import (
	//"../log"
	//"encoding/json"
	//"./api"
	"../config/utils"
	"../tools"
	//"./calculate"
	//"./decision"
)

func RattrapCalcSymbol(api_c *utils.API, symbol tools.Symbol) {
	/*
	var err error
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	res, err = api.RequestGetDataForSymbol(api_c, symbol, api_c.From, api_c.To)

	if err != nil {
		switch {
		case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
			log.Error("For : ", symbol.Name, " - ", err.Error())
			return
		case err.Error() == "No data to retrieve in this range":
			log.Error("For : ", symbol.Name, " - ", err.Error())
			return
		default:
			log.FatalError(err)
			return
		}
	}

	var res_bids []tools.Bid

	for _, res_b := range res.Bids {
		if res_b.Calculations_s != "{}" {
			err = json.Unmarshal([]byte(res_b.Calculations_s), &res_b.Calculations)
			if err != nil {
				log.Error("For : ", res_b, " - ", err.Error())
				continue
			}
		}
		res_bids = append(res_bids, res_b)
	}
*/
	/*for _, b_to_update := range calculate.CalculateBids(api, res_bids) {
		log.Info("Update calculations for -> ", b_to_update.Id, " | ", b_to_update.Symbol.Name, " | ", b_to_update.Bid_at_s, " | ", b_to_update.Last_bid, " | ", b_to_update.Calculations)
		res, err = api.RequestSetCalculation(api_c, tools.Symbol{}, b_to_update, tools.Trade{}, time.Time{}, time.Time{})
		if err != nil {
			switch {
			case err.Error() == "Unable to connect to any of the specified MySQL hosts.":
				log.Error(err.Error())
				continue
			default:
				log.FatalError(err)
				continue
			}
		}
	}*/
}
