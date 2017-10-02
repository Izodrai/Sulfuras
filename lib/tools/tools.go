package tools

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Res_error struct {
	IsAnError    bool
	MessageError string
}

type Symbol struct {
	Id           int     `json:"Id"`
	Name         string  `json:"Name"`
	Description  string  `json:"Description"`
	State        string  `json:"State"`
	Lot_max_size float64 `json:"Lot_max_size"`
	Lot_min_size float64 `json:"Lot_min_size"`
}

type Bid struct {
	Id             int
	Symbol         Symbol
	Bid_at_s       string `json:"Bid_at"`
	Bid_at         time.Time
	Last_bid       float64
	Calculations_s string `json:"Calculations"`
	Calculations   map[string]float64
}

type Trade struct {
	Id            int
	Symbol        Symbol
	Trade_type    int
	Volume        float64
	Opened_reason string
	Closed_reason string
}

func (b *Bid) Base64Calculations() string {
	by, _ := json.Marshal(b.Calculations)
	return string(base64.StdEncoding.EncodeToString(by))
}

type Response struct {
	ResError Res_error `json:"Error"`
	Error 	 error
	Bids     []Bid
	Symbols  []Symbol
	Trades   []Trade
}

type Request struct {
	URL_request string
	Symbol 		Symbol
	Resp 		chan Response
}