package tools

import (
	"encoding/base64"
	"encoding/json"
	"time"
	"strings"
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

func (b *Bid) ParseAPITime() error {
	var err error

	if strings.ContainsAny(b.Bid_at_s, "T"){
		b.Bid_at, err = time.Parse("2006-01-02T15:04:05",b.Bid_at_s)
		if err != nil {
			return err
		}
	} else {
		b.Bid_at, err = time.Parse("2006-01-02 15:04:05",b.Bid_at_s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bid) UnmarshalCalculation() error {
	err := json.Unmarshal([]byte(b.Calculations_s), &b.Calculations)
	if err != nil {
		return err
	}
	return nil
}

type Database struct {
	Host string
	Name string
	Login string
	Pass string
	Port string
}

func (d *Database) DSN() string{
	return d.Login+":"+d.Pass+"@tcp("+d.Host+":"+d.Port+")/"+d.Name+"?charset=utf8"
}