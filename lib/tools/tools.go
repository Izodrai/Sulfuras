package tools

import (
	"encoding/json"
	"strings"
	"time"
	"math"
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
	Symbol_name    string `json:"Symbol"`
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

type Response struct {
	ResError Res_error `json:"Error"`
	Error    error
	Bids     []Bid
	Symbols  []Symbol
	Trades   []Trade
}

type Request struct {
	URL_request string
	Symbol      Symbol
	Resp        chan Response
}

type SavedBids struct {
	LastDate 	time.Time
	ById		map[int]Bid
	ByDate		map[time.Time]Bid
}

func (sb *SavedBids) AddBid(b Bid) {
	sb.ById[b.Id] = b
	sb.ByDate[b.Bid_at] = b
	if b.Bid_at.After(sb.LastDate) {
		sb.LastDate = b.Bid_at
	}
}

func (sb *SavedBids) SortBidsByDateAscFrom(tFrom time.Time) []Bid {

	var bids []Bid
	var tNow = time.Now()

	var _, m, _ = tNow.Clock()
	var tNewFrom = tFrom

	if mod := math.Mod(float64(m), 5); mod != 0 {
		for math.Mod(float64(m), 5) != 0 {
			tNewFrom = tNewFrom.Add(-1 * time.Minute)
			_, m, _ = tNewFrom.Clock()
		}
	} else {
		tNewFrom = tFrom
	}

	for tNewFrom.Before(tNow) {
		if b, ok := sb.ByDate[tNewFrom]; ok {
			bids = append(bids, b)
		}
		tNewFrom = tNewFrom.Add(5 * time.Minute)
	}

	return bids
}

func (b *Bid) Feed(symbol Symbol) error {

	if b.Symbol_name == symbol.Name {
		symbol.Description = ""
		b.Symbol = symbol
	}

	if err := b.ParseAPITime(); err != nil {
		return err
	}

	if b.Calculations_s == "" {
		b.Calculations_s = "{}"
	}

	if err := b.UnmarshalCalculation(); err != nil {
		return err
	}

	return nil
}

func (b *Bid) ParseAPITime() error {
	var err error

	b.Bid_at_s = strings.Replace(b.Bid_at_s, "Z", "", -1)
	b.Bid_at_s = strings.Replace(b.Bid_at_s, "T", " ", -1)

	b.Bid_at, err = time.Parse("2006-01-02 15:04:05", b.Bid_at_s)
	if err != nil {
		return err
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
	Host  string
	Name  string
	Login string
	Pass  string
	Port  string
}

func (d *Database) DSN() string {
	return d.Login + ":" + d.Pass + "@tcp(" + d.Host + ":" + d.Port + ")/" + d.Name + "?charset=utf8"
}
