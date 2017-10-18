package utils

import (
	"../../tools"
	"time"
	"database/sql"
)

type API struct {
	Url                   string `json:"Url"`
	Symbols               []tools.Symbol
	AllSymbols		      []tools.Symbol
	InactivSymbols		  []tools.Symbol
	ActivSymbols		  []tools.Symbol
	StandbySymbols		  []tools.Symbol
	RetrievePeriode_s     map[string]Periode `json:"RetrievePeriode"`
	RetrievePeriode       map[time.Weekday]Periode
	Calculations          Calculation
	From_s                string `json:"From"`
	From                  time.Time
	To_s                  string `json:"To"`
	To                    time.Time
	StepRetrieve_s 		  string `json:"StepRetrieve"`
	StepRetrieve   		  time.Duration
	Database_Info		  tools.Database `json:"Database"`
	Database			  *sql.DB
}

type Calculation struct {
	Sma       []int   `json:"SMA"`
	SmaVersus [][]int `json:"Sma_versus"`
	Ema       []int   `json:"EMA"`
	EmaVersus [][]int `json:"Ema_versus"`
	Macd      interface{}
}

type Periode struct {
	Deactivated bool   `json:"deactivated"`
	Limited     bool   `json:"limited"`
	Start       string `json:"start"`
	Start_h     int
	Start_m     int
	Start_s     int
	End         string `json:"end"`
	End_h       int
	End_m       int
	End_s       int
}
