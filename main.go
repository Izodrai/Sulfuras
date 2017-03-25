package main

import (
	"./lib/config"
	"./lib/log"
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

	fmt.Println(configFile)

}
