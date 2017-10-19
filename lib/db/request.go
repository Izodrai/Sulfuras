package db

import (
	"database/sql"
	"time"
	"../tools"
	"../config/utils"
	"encoding/json"
)

func LoadLastBidsForSymbol(api_c *utils.API, symbol tools.Symbol, bids *[]tools.Bid) error {
	var err error
	var rows *sql.Rows

	if rows, err = api_c.Database.Query("SELECT id, bid_at, last_bid, calculations FROM stock_values WHERE bid_at >= ? AND symbol_id = ?", api_c.From, symbol.Id); err != nil {
		return err
	}

	for rows.Next() {
		var id int
		var bid_at string
		var last_bid float64
		var calculations []byte

		if err = rows.Scan(&id, &bid_at, &last_bid, &calculations); err != nil {
			return err
		}

		*bids = append(*bids,tools.Bid{id, symbol, bid_at, time.Time{}, last_bid, string(calculations), map[string]float64{}})
	}

	return  nil
}

func LoadSymbolStatus(api_c *utils.API, symbols *[]tools.Symbol) error {
	var err error
	var rows *sql.Rows

	if rows, err = api_c.Database.Query("SELECT id, reference, description, lot_max_size, lot_min_size, state FROM symbols"); err != nil {
		return err
	}

	for rows.Next() {
		var id int
		var reference, description, state string
		var lot_max_size, lot_min_size float64

		if err = rows.Scan(&id, &reference, &description, &lot_max_size, &lot_min_size, &state); err != nil {
			return err
		}

		*symbols = append(*symbols,tools.Symbol{id, reference, description, state, lot_max_size, lot_min_size})
	}

	return  nil
}

func UpdateCalculation(api_c *utils.API, bid *tools.Bid) error {

	var err error
	var stmt *sql.Stmt

	if stmt, err = api_c.Database.Prepare("update stock_values set calculations=? where id=?"); err != nil {
		return err
	}

	by, _ := json.Marshal(bid.Calculations)

	if _, err = stmt.Exec(string(by), bid.Id); err != nil {
		return err
	}

	return nil
}

