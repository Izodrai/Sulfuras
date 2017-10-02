package api

import (
	//"time"
	//"../../tools"
	//"../../config/utils"
)
/*
func OpenTrade(api *utils.API, trade tools.Trade) {
	var err error
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	res, err = RequestOpenTrade(api, tools.Symbol{}, tools.Bid{}, trade, time.Time{}, time.Time{})
	if err != nil {
		log.Error(err.Error())
		return
	}

	if len(res.Trades) != 1 {
		log.Error("len(res.Trades) != 1 -> ", len(res.Trades))
		return
	}
	//DO NOT FORGET
}*/

/*
func GetOpenedTrades(api *utils.API, open_trades map[int]tools.Trade) error {
	var err error
	var res = tools.Response{tools.Res_error{true, "init"},nil,[]tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	if res = Request(RequestGetOpenTrades(api), api); res.Error != nil {
		return err
	}

	if len(res.Trades) != 0 {
		for _, t := range res.Trades {
			open_trades[t.Id] = t
		}
	}
	return nil
}*/

