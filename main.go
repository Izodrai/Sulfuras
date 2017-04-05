package main

import (
	"encoding/json"
	"./lib/config"
	"./lib/log"
	"./lib/tools"
	"net/http"
	"errors"
	"io/ioutil"
	"time"
	"fmt"
)

func init() {
	log.InitLog(false)
}

func api_request(conf config.Config, symbol tools.Symbol, tRetrieve time.Time) (tools.Response, error) {

	var err error
	var data []byte
	var res = tools.Response{tools.Res_error{true,"init"},[]tools.Bid{}}
	c := make(chan error, 1)
	var resp *http.Response

	go func() {
		resp, err = http.Get(conf.API.Url+"get_symbol/"+symbol.Name+"/"+tRetrieve.Format("2006-01-02"))
		c <- err
	}()

	select {
	case err := <-c:
		if err != nil {
			return res, err
		}
	case <-time.After(time.Second * 350):
		return res, errors.New("HTTP source timeout")
	}

	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return res, err
	}

	if res.Error.IsAnError {
		return res, errors.New(res.Error.MessageError)
	}

	return res, nil
}

func retrieve_max_import(conf config.Config, symbols *[]tools.Symbol) error {
	var err error
	var res tools.Response
	var tempo_symbols []tools.Symbol

	for _, symbol := range *symbols {

		log.Info("Retrieve last data insert for ",symbol.Name)

		/* VTempo */
		var i = -1
		var ct int
		t:= time.Now().UTC()
		tRetrieve := t.AddDate(0,0,i)

		retry_api_request:

		// Requete sur l'api avec tRetrieve
		if res, err = api_request(conf, symbol, tRetrieve); err != nil {
			return err
		}

		// Si CT >= 30 donc 31 jours, on considère que la base et vide et on prend cette valeur à importer
		if ct >= 30 {
			symbol.Last_insert = tRetrieve
			tempo_symbols = append(tempo_symbols, symbol)
			continue
		}
		ct++

		// Si l'api ne retourne aucun bid ça veut dire que le plus ancien n'est pas dans cette plage de temps, on l'augmente donc de 10
		if len(res.Bids) == 0 {
			i -= 1
			tRetrieve = t.AddDate(0,0,i)
			goto retry_api_request
		}

		// Une fois qu'il y a au moins un bid, on cherche le plus récent et on prend cette date comme référence
		for _,bid := range res.Bids {
			if bid.Bid_at, err = time.Parse("2006-01-02T15:04:05", bid.Bid_at_s); err != nil {
				return err
			}

			if bid.Bid_at.After(symbol.Last_insert) {
				symbol.Last_insert = bid.Bid_at
			}
		}

		tempo_symbols = append(tempo_symbols, symbol)

		/* Vfinal */
		/*
		go func() {
			resp, err = http.Get(conf.API.Url+"get_last_bid_for_symbol/"+symbol.Name)
			c <- err
		}()

		select {
		case err := <-c:
			if err != nil {
				return err
			}
		case <-time.After(time.Second * 350):
			return errors.New("HTTP source timeout")
		}

		defer resp.Body.Close()

		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		fmt.Println(string(data))
		*/
	}

	*symbols = tempo_symbols

	return nil
}

func main() {
	var err error
	var conf config.Config
	var res = tools.Response{tools.Res_error{true,"init"},[]tools.Bid{}}

	var configFile string = "config.json"

	fmt.Println("")
	fmt.Println("")

	log.YellowInfo("Running Sulfuras")

	if err = conf.LoadConfig(configFile); err != nil {
		log.FatalError(err)
		return
	}

	fmt.Println("")

	var symbols []tools.Symbol
	symbols = append(symbols, tools.Symbol{ 41,"EURUSD","", time.Time{}})

	if err = retrieve_max_import(conf, &symbols); err != nil {
		log.FatalError(err)
		return
	}

	for {
		for i, symbol := range symbols {
			c := make(chan error, 1)
			var resp *http.Response
			var data []byte

			log.Info("Retrieve data for ",symbol.Name," between ",symbol.Last_insert.Format("2006-01-02")," and ",time.Now().UTC().Format("2006-01-02 15:04:05"), " (UTC)")

			go func() {
				resp, err = http.Get(conf.API.Url+"update_symbol/"+symbol.Name+"/"+symbol.Last_insert.Format("2006-01-02"))
				c <- err
			}()

			select {
			case err := <-c:
				if err != nil {
					log.FatalError(err)
					return
				}
			case <-time.After(time.Second * 350):
				log.FatalError("HTTP source timeout")
				return
			}

			defer resp.Body.Close()

			data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.FatalError(err)
				return
			}

			err = json.Unmarshal(data, &res)
			if err != nil {
				log.FatalError(err)
				continue
			}

			if res.Error.IsAnError {
				log.FatalError(res.Error.MessageError)
				continue
			}

			symbols[i].Last_insert = time.Now().UTC()

			fmt.Println(res.Error.MessageError)

			time.Sleep(1 * time.Minute)
		}
	}
}
