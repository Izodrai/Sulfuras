package tools

import (
	"encoding/json"
	"time"
	"errors"
	"strconv"
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
	Bid_at_ts      int64 `json:"Bid_at"`
	Bid_at         time.Time
	Last_bid       float64
	Calculations_s string `json:"Calculations"`
	Calculations   map[string]float64
}

type Trade struct {
	Id            int
	Symbol        Symbol
	Type	      int
	Opened_value  float64
	Closed_value  float64
	Volume        float64
	Opened_reason string
	Closed_reason string
	Open          bool
	Bids 	      []Bid
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
	LastDate 	int64
	ById		map[int]Bid
	ByDate		map[int64]Bid
}

func (sb *SavedBids) AddBid(b Bid) {
	sb.ById[b.Id] = b
	sb.ByDate[b.Bid_at_ts] = b
	if b.Bid_at_ts > sb.LastDate {
		sb.LastDate = b.Bid_at_ts
	}
}

func (sb *SavedBids) SortBidsByDateAscFrom(tFrom int64) []Bid {

	var bids []Bid
	var tNow = time.Now().Unix()

	for i := tFrom; i <= tNow; i++ {
		if b, ok := sb.ByDate[i]; ok {
			bids = append(bids, b)
		}
	}

	return bids
}

type SavedTrades struct {
	OpenTrades map[int]Trade
	ClosedTrades map[int]Trade
}

func (st *SavedTrades) CloseTrade(id int) error {
	var ok bool
	var trade Trade

	if trade, ok = st.OpenTrades[id]; !ok {
		if _, ok = st.ClosedTrades[id]; !ok {
			return errors.New("This trade " + strconv.Itoa(id) + " do not exist")
		}
		return errors.New("This trade " + strconv.Itoa(id) + " is already close")
	}

	st.ClosedTrades[id] = trade

	delete(st.OpenTrades, id)

	return nil
}

func (b *Bid) Feed(symbol Symbol) error {

	if b.Symbol_name == symbol.Name {
		symbol.Description = ""
		b.Symbol = symbol
	}

	b.ParseAPITime()

	if b.Calculations_s == "" {
		b.Calculations_s = "{}"
	}

	if err := b.UnmarshalCalculation(); err != nil {
		return err
	}

	return nil
}

func (b *Bid) ParseAPITime() {
	b.Bid_at = time.Unix(b.Bid_at_ts, 0)
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
