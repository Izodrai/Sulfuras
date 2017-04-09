package config

import (
	"encoding/json"
	"io/ioutil"
	"../tools"
	"time"
)

type API struct {
	Url string `json:"Url"`
	Symbols_s []string `json:"Symbols"`
	Symbols []tools.Symbol
}

type Config struct {
	API API `json:"API"`
}

func (c *Config) LoadConfig(configFile string) error {

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		return err
	}

	for _,s_name := range c.API.Symbols_s {
		c.API.Symbols = append(c.API.Symbols, tools.Symbol{0,s_name,"", time.Time{}})
	}

	return nil
}
