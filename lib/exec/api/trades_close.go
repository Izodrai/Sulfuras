package api
/*
import (
	"time"
	"../../tools"
	"../../log"
	"../../config/utils"
)

func CloseTrade(api utils.API, trade_to_close tools.Trade) {
	var err error
	//var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	_, err = RequestCloseTrade(api, tools.Symbol{}, tools.Bid{}, trade_to_close, time.Time{}, time.Time{})
	if err != nil {
		log.Error(err.Error())
		return
	}
	//DO NOT FORGET
}

func CloseAllTrade(api utils.API, open_trades map[int]tools.Trade, reason string) {
	for _, trade_to_close := range open_trades {
		trade_to_close.Closed_reason = reason
		CloseTrade(api, trade_to_close)
	}
}

func CloseAskTrade(api utils.API, open_trades map[int]tools.Trade) {
	for _, trade_to_close := range open_trades {
		if trade_to_close.Trade_type == 0 {
			trade_to_close.Closed_reason = "bid"
			CloseTrade(api, trade_to_close)
		}
	}
}

func CloseBidTrade(api utils.API, open_trades map[int]tools.Trade) {
	for _, trade_to_close := range open_trades {
		if trade_to_close.Trade_type == 1 {
			trade_to_close.Closed_reason = "ask"
			CloseTrade(api, trade_to_close)
		}
	}
}*/