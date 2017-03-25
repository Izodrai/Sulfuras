package config

import (
	"encoding/json"
	"io/ioutil"
)

type API struct {
	Url string `json:"Url"`
	Actions []string `json:"Actions"`
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

	return nil
}
