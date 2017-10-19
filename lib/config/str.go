package config

import "./utils"

type Config struct {
	LogFile string    `json:"LogFile"`
	API     utils.API `json:"API"`
}
