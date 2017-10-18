package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"../tools"
	"./utils"
	"../exec/api"
)

func (c *Config) Load(configFile string) error {

	var err error
	var file []byte

	if file, err = ioutil.ReadFile(configFile); err != nil {
		return err
	}

	if err = json.Unmarshal(file, c); err != nil {
		return err
	}

	if err =c.setRetrievePeriode(); err != nil {
		return err
	}

	if c.API.From, err = time.Parse("2006-01-02", c.API.From_s); err != nil {
		return err
	}

	if c.API.To_s == "now" {
		c.API.To = time.Now().AddDate(0, 0, 1).UTC()
	} else {
		if c.API.To, err = time.Parse("2006-01-02", c.API.To_s); err != nil {
			return err
		}
	}

	if c.API.StepRetrieve, err = time.ParseDuration(c.API.StepRetrieve_s); err != nil {
		return err
	}

	c.API.Symbols = []tools.Symbol{}

	if err = api.GetSymbolsStatus(&c.API); err != nil {
		return err
	}

	return nil
}

func (c *Config) setRetrievePeriode() error {

	var err error

	c.API.RetrievePeriode = make(map[time.Weekday]utils.Periode)

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

	return nil;
}