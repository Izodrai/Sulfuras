package api

import (
	"../../config/utils"
	"../../db"
	"../../tools"
)

func GetSymbolsStatus(api_c *utils.API) error {

	var err error
	var symbols []tools.Symbol

	if err = db.LoadSymbolStatus(api_c, &symbols); err != nil {
		return err
	}

	for _, s := range symbols {

		api_c.AllSymbols[s.Id] = s

		if s.State != "inactive" {
			api_c.Symbols[s.Id] = s
			api_c.Symbols_t = append(api_c.Symbols_t, s)
		} else {
			api_c.InactivSymbols[s.Id] = s
		}

		if s.State == "active" {
			api_c.ActivSymbols[s.Id] = s
		}

		if s.State == "standby" {
			api_c.StandbySymbols[s.Id] = s
		}
	}

	return nil
}
