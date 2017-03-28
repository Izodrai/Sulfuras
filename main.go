package main

import (
	"./lib/config"
	"./lib/log"
	"./lib/tools"
	"net/http"
	"io/ioutil"
	"time"
	"fmt"
)

func init() {
	log.InitLog(false)
}

func main() {
	var err error
	var conf config.Config

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
	symbols = append(symbols, tools.Symbol{ "41","EURUSD",""})

	for {

		c := make(chan error, 1)
		var resp *http.Response
		var data []byte
		
		t:= time.Now()
		tUpdate := t.AddDate(0,0,-1)
		
		log.Info("Retrieve data for ",symbols[0].Name," between ",tUpdate.Format("2006-01-02")," and ",t.Format("2006-01-02 15:04:05"), " (Local)") 

		go func() {
			resp, err = http.Get(conf.API.Url+"update_symbol/"+symbols[0].Name+"/"+tUpdate.Format("2006-01-02"))
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

		fmt.Println(string(data))

		time.Sleep(1 * time.Minute)
	}

}
