package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"../tools"
)

type API struct {
	Url               string             `json:"Url"`
	Symbols           []tools.Symbol     `json:"Symbols"`
	RetrievePeriode_s map[string]Periode `json:"RetrievePeriode"`
	RetrievePeriode   map[time.Weekday]Periode
	Calculations      Calculation
	From_s            string `json:"From"`
	From              time.Time
	To_s              string `json:"To"`
	To                time.Time
}

type Calculation struct {
	SMA  []int
	EMA  []int
	MACD interface{}
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

type Config struct {
	LogFile string `json:"LogFile"`
	API     API    `json:"API"`
}

func (c *Config) LoadConfig(configFile string) error {

	var err error

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		return err
	}

	c.API.RetrievePeriode = make(map[time.Weekday]Periode)

	for day, per := range c.API.RetrievePeriode_s {
		var day_t time.Weekday
		switch day {
		case "Monday":
			day_t = time.Weekday(1)
		case "Tuesday":
			day_t = time.Weekday(2)
		case "Wednesday":
			day_t = time.Weekday(3)
		case "Thursday":
			day_t = time.Weekday(4)
		case "Friday":
			day_t = time.Weekday(5)
		case "Saturday":
			day_t = time.Weekday(6)
		case "Sunday":
			day_t = time.Weekday(0)
		}

		start := strings.Split(per.Start, ":")

		if per.Start_h, err = strconv.Atoi(start[0]); err != nil {
			return err
		}
		if per.Start_m, err = strconv.Atoi(start[1]); err != nil {
			return err
		}
		if per.Start_s, err = strconv.Atoi(start[2]); err != nil {
			return err
		}

		end := strings.Split(per.End, ":")

		if per.End_h, err = strconv.Atoi(end[0]); err != nil {
			return err
		}
		if per.End_m, err = strconv.Atoi(end[1]); err != nil {
			return err
		}
		if per.End_s, err = strconv.Atoi(end[2]); err != nil {
			return err
		}

		c.API.RetrievePeriode[day_t] = per
	}

	if len(c.API.RetrievePeriode) != 7 {
		return errors.New("All days are not valid...")
	}

	c.API.From, err = time.Parse("2006-01-02", c.API.From_s)
	if err != nil {
		return err
	}

	if c.API.To_s == "now" {
		c.API.To = time.Now().AddDate(0, 0, 1).UTC()
	} else {
		c.API.To, err = time.Parse("2006-01-02", c.API.To_s)
		if err != nil {
			return err
		}
	}

	return nil
}
