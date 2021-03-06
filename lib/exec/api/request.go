package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"../../config/utils"
	"../../tools"
	"strings"
)

func RequestFeedSymbol(api *utils.API, symbol tools.Symbol, tFrom int64) string {
	return api.Url + "get_xtb_bids/" + api.Token + "/" + symbol.Name + "/" + strconv.FormatInt(tFrom, 10) + "/"
}

func RequestOpenTrade(api *utils.API, trade tools.Trade) string {
	return api.Url + "open_trade/" + strconv.Itoa(trade.Symbol.Id) + "/" + strconv.Itoa(trade.Type) + "/" + strconv.FormatFloat(trade.Volume, 'f', -1, 64) + "/" + trade.Opened_reason
}

func RequestCloseTrade(api *utils.API, trade tools.Trade) string {
	return api.Url + "close_trade/" + strconv.Itoa(trade.Id) + "/" + trade.Closed_reason
}

func RequestGetOpenTrades(api *utils.API) string {
	return api.Url + "get_open_trades/"
}

func Request(req_url string, api *utils.API) tools.Response {

	var err error
	var data []byte
	var res = tools.Response{tools.Res_error{true, "init -> " + req_url}, nil, []tools.Bid{}, []tools.Symbol{}, []tools.Trade{}}
	c := make(chan error, 1)
	var resp *http.Response

	go func() {
		resp, err = http.Get(req_url)
		c <- err
	}()

	select {
	case err := <-c:
		if err != nil {
			res.Error = err
			return res
		}
	case <-time.After(90 * time.Second):
		res.Error = errors.New("HTTP timeout")
		return res
	}

	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		res.Error = err
		return res
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		if strings.Contains(string(data), "Error 403 - This web app is stopped.</h1>") {
			res.Error = errors.New("Azure app disconected")
			return res
		}
		res.Error = err
		return res
	}

	if res.ResError.IsAnError {
		res.Error = errors.New(res.ResError.MessageError + " ||| " + string(data))
		return res
	}

	return res
}
