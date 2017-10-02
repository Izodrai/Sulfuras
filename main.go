package main

import (
	"./lib/config"
	"./lib/exec"
	"./lib/log"
	"./lib/tools"
	l "log"
	"os"
	"path"
	"strconv"
	"time"
)

const vers_algo = "v0.0.6"

func main() {

	var err error
	var mode int
	var conf config.Config

	if len(os.Args) != 2 && len(os.Args) != 3 {
		l.Println(log.RED + "Invalid Argument(s)" + log.STOP)
		l.Println(log.RED + "Usuel : ./market-binary config_file mode (optional)" + log.STOP)
		l.Println(log.RED + "Mode : (1) rattrap" + log.STOP)
		os.Exit(1)
	}

	if err = conf.Load(path.Join(os.Args[1])); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	if err = log.InitLog(true, conf); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	if len(os.Args) == 3 {
		if mode, err = strconv.Atoi(os.Args[2]); err != nil {
			log.Error(log.RED + err.Error() + log.STOP)
			os.Exit(1)
		}

		if mode != 1 {
			log.Error(log.RED + "Bad mode ! " + os.Args[2] + log.STOP)
			os.Exit(1)
		}
	}

	log.SkipLines(4)

	if mode == 1 {
		log.YellowInfo("Running Sulfuras-Rattrap")
	} else {
		log.YellowInfo("Running Sulfuras version : ", vers_algo)
	}

	log.SkipLines(2)
	log.WhiteInfo("> Symbol Configuration")
	log.SkipLines(1)

	for _, s := range conf.API.Symbols {
		log.Info("\tSymbol id : ", s.Id, " (", s.Name, ") | Status : ", s.State)
	}

	log.SkipLines(1)
	log.WhiteInfo("> Days Configuration")
	log.SkipLines(1)

	for i := 0; i < 7; i++ {
		p := conf.API.RetrievePeriode[time.Weekday(i)]
		log.WhiteInfo(">> ", time.Weekday(i).String())
		log.Info("\tDeactivated : ", p.Deactivated)
		log.Info("\tLimited     : ", p.Limited)
		log.Info("\tStart time  : ", p.Start)
		log.Info("\tEnd time    : ", p.End)
		log.SkipLines(1)
	}

	log.Info("##########")
	log.SkipLines(2)
	log.WhiteInfo("Start current retrieve")
	log.Info("#############################")

	/*
	if mode == 1 {

		for _, symbol := range conf.API.Symbols {
			log.WhiteInfo(symbol.Name)
			exec.RattrapCalcSymbol(&conf.API, symbol)
			time.Sleep(1 * time.Second)
		}
		time.Sleep(10 * time.Second)

		log.Info("#############################")
		log.WhiteInfo("End rattrap")
		log.Info("#############################")
		os.Exit(0)

	} else {*/

	var bids = make(map[int]map[int]tools.Bid)

	if err = exec.ExecNotInactivSymbols(&conf.API, bids); err != nil {
		log.Error(err)
	}
		/*
			for {
				time.Sleep(24 * time.Hour)
			}*/
	//}
}
