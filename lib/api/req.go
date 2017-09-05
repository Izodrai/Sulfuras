package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"../config"
	"../tools"
)

func Api_request(conf config.Config, symbol tools.Symbol, bid_to_update tools.Bid, t int, tfrom, tnow time.Time) (tools.Response, error) {

	var err error
	var data []byte
	var res = tools.Response{tools.Res_error{true, "init"}, []tools.Bid{}}
	c := make(chan error, 1)
	var resp *http.Response
	var req_url string

	switch t {
	case 1:
		req_url = conf.API.Url + "feed_symbol_from_last_insert/" + strconv.Itoa(symbol.Id) + "/"
	case 2:
		req_url = conf.API.Url + "set_calculation/" + strconv.Itoa(bid_to_update.Id) + "/" + bid_to_update.Base64Calculations()
	case 3:
		req_url = conf.API.Url + "get_data_for_symbol/" + strconv.Itoa(symbol.Id) + "/" + tfrom.Format("2006-01-02") + "/" + tnow.Format("2006-01-02")
	}

	go func() {
		resp, err = http.Get(req_url)
		c <- err
	}()

	select {
	case err := <-c:
		if err != nil {
			return res, err
		}
	case <-time.After(time.Second * 350):
		return res, errors.New("HTTP source timeout")
	}

	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return res, err
	}

	if res.Error.IsAnError {
		return res, errors.New(res.Error.MessageError)
	}

	return res, nil
}
