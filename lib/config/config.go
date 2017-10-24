package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"../exec/api"
	"../tools"
	"./utils"
	"../db"
	"path"
	"os"
)

func (c *Config) Init() error {

	var err error

	if err = c.InitLoad(path.Join(os.Args[1])); err != nil {
		return err
	}

	if err = db.Init(&c.API); err != nil {
		return err
	}

	if err = c.Load(path.Join(os.Args[1])); err != nil {
		return err
	}

	if err = c.LoadSymbolStatus(); err != nil {
		return err
	}

	return nil
}

func (c *Config) InitLoad(configFile string) error {
	var err error
	var file []byte

	if file, err = ioutil.ReadFile(configFile); err != nil {
		return err
	}

	if err = json.Unmarshal(file, c); err != nil {
		return err
	}
	return nil
}

func (c *Config) Load(configFile string) error {

	var err error
	var file []byte

	if c.ProdEnv {
		if file, err = db.LoadConfig(&c.API); err != nil {
			return err
		}

		if err = json.Unmarshal(file, c); err != nil {
			return err
		}
	}

	if err = c.setRetrievePeriode(); err != nil {
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

	c.API.Symbols = make(map[int]tools.Symbol)
	c.API.InactivSymbols = make(map[int]tools.Symbol)
	c.API.ActivSymbols = make(map[int]tools.Symbol)
	c.API.StandbySymbols = make(map[int]tools.Symbol)
	c.API.AllSymbols = make(map[int]tools.Symbol)

	return nil
}

func (c *Config) LoadSymbolStatus() error {
	var err error
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

	return nil
}
