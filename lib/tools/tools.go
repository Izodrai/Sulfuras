package tools

import (
	"time"
	"encoding/json"
	"encoding/base64"
)

type Symbol struct {
	Id          int    `json:"Id"`
	Name        string `json:"Name"`
}

type Res_error struct {
	IsAnError    bool
	MessageError string
}

type Bid struct {
	Id       int
	Symbol   Symbol
	Bid_at_s string `json:"Bid_at"`
	Bid_at   time.Time
	Last_bid float64
	Calculations_s string `json:"Calculations"`
	Calculations map[string]float64
}

func (b *Bid) Base64Calculations() string {
	by, _ := json.Marshal(b.Calculations)
	return string(base64.StdEncoding.EncodeToString(by))
}

type Response struct {
	Error Res_error
	Bids  []Bid
}
