package api

import (
	"../../tools"
	"../../config/utils"
)

func GetSymbolsStatus(api *utils.API) error {

	var err error
	var res = tools.Response{tools.Res_error{true, "init_get_symbols_status"}, nil, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}

	if res = Request(RequestGetSymbolStatus(api), api); res.Error != nil {
		return err
	}

	for _, s := range res.Symbols {
		if s.State != "inactive" {
			api.Symbols = append(api.Symbols, s)
		}
	}

	return nil
}