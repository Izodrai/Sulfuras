package tools

import (
	"time"
)

type Symbol struct {
	Id   				int 		`json:"Id"`
	Name 				string 	`json:"Name"`
  Desc 				string 	`json:"Desc"`
	Last_insert time.Time
}

type Res_error struct {
	IsAnError 		bool
	MessageError 	string
}

type Bid struct {
	Id     		int
	Symbol 		Symbol
	Bid_at_s 	string `json:"Bid_at"`
	Bid_at 		time.Time
	Last_bid	float64
}

type Response struct {
	Error 	Res_error
	Bids 		[]Bid
}
