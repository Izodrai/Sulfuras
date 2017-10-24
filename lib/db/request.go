package db

import (
	"../config/utils"
	"../tools"
	"database/sql"
	"encoding/json"
	"time"
	"errors"
)


func LoadConfig(api_c *utils.API) ([]byte, error) {

	var conf []byte

	mutex.Lock()

	if err := api_c.Database.QueryRow("SELECT configuration FROM configuration").Scan(&conf); err != nil {
		return []byte{}, errors.New("LoadConfig : " + err.Error())
	}

	mutex.Unlock()

	return conf, nil
}

func LoadSymbolStatus(api_c *utils.API, symbols *[]tools.Symbol) error {
	var err error
	var rows *sql.Rows

	mutex.Lock()

	if rows, err = api_c.Database.Query("SELECT id, reference, description, lot_max_size, lot_min_size, state FROM symbols"); err != nil {
		return errors.New("LoadSymbolStatus Query : " + err.Error())
	}

	for rows.Next() {
		var id int
		var reference, description, state string
		var lot_max_size, lot_min_size float64

		if err = rows.Scan(&id, &reference, &description, &lot_max_size, &lot_min_size, &state); err != nil {
			return errors.New("LoadSymbolStatus Scan : " + err.Error())
		}

		*symbols = append(*symbols, tools.Symbol{id, reference, description, state, lot_max_size, lot_min_size})
	}

	mutex.Unlock()

	return nil
}

func LoadLastBidsForSymbol(api_c *utils.API, symbol tools.Symbol, bids *[]tools.Bid) error {
	var err error
	var rows *sql.Rows

	mutex.Lock()

	if rows, err = api_c.Database.Query("SELECT id, bid_at, last_bid, calculations FROM stock_values WHERE bid_at >= ? AND symbol_id = ?", api_c.From, symbol.Id); err != nil {
		return errors.New("LoadLastBidsForSymbol Query : " + err.Error())
	}

	for rows.Next() {
		var id int
		var bid_at string
		var last_bid float64
		var calculations []byte

		if err = rows.Scan(&id, &bid_at, &last_bid, &calculations); err != nil {
			return errors.New("LoadLastBidsForSymbol Scan : " + err.Error())
		}

		symbol.Description = ""

		*bids = append(*bids, tools.Bid{id, symbol.Name, symbol, bid_at, time.Time{}, last_bid, string(calculations), map[string]float64{}})
	}

	mutex.Unlock()

	return nil
}

func InsertOrUpdateBid(api_c *utils.API, bid *tools.Bid) error {

	var err error

	mutex.Lock()

	by, _ := json.Marshal(bid.Calculations)

	if _, err = api_c.Database.Exec("INSERT INTO stock_values (`symbol_id`, `bid_at`, `last_bid`, `calculations`) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE `last_bid`= ?, `calculations` = ?", bid.Symbol.Id, bid.Bid_at, bid.Last_bid, string(by), bid.Last_bid, string(by)); err != nil {
		return errors.New("InsertOrUpdateBid Prepare : " + err.Error())
	}
	mutex.Unlock()

	return nil
}