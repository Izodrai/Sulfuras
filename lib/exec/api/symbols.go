package api

import (
	"../../tools"
	"../../config/utils"
	"../../db"
)

func GetSymbolsStatus(api_c *utils.API) error {

	var err error
	var symbols []tools.Symbol

	if err = db.LoadSymbolStatus(api_c, &symbols); err != nil {
		return err
	}

	for _, s := range symbols {

		api_c.AllSymbols = append(api_c.AllSymbols, s)

		if s.State != "inactive" {
			api_c.Symbols = append(api_c.Symbols, s)
		} else {
			api_c.InactivSymbols = append(api_c.InactivSymbols, s)
		}

		if s.State == "active" {
			api_c.ActivSymbols = append(api_c.ActivSymbols, s)
		}

		if s.State == "standby" {
			api_c.StandbySymbols = append(api_c.StandbySymbols, s)
		}
	}

	return nil
}