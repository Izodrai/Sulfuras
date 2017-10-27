package main

import (
	"./lib/config"
	"./lib/exec"
	"./lib/log"
	"./lib/tools"
	"./lib/web"
	"./lib/db"
	"os"
	"time"
	l "log"
)

const vers_algo = "v0.0.6"

func main() {

	var err error
	var conf config.Config
	var rattrapMode bool

	if len(os.Args) != 2 && len(os.Args) != 3 {
		l.Println(log.RED + "Invalid Argument(s)" + log.STOP)
		l.Println(log.RED + "Usuel 1 : ./market-binary config_file" + log.STOP)
		l.Println(log.RED + "Usuel 2 : ./market-binary config_file rattrap_mode(true/false) " + log.STOP)
		os.Exit(1)
	}

	if err = conf.Init(); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	defer conf.API.Database.Close()

	go db.KeepOpen(&conf.API)

	if err = log.InitLog(true, conf.LogFile); err != nil {
		l.Println(log.RED + err.Error() + log.STOP)
		os.Exit(1)
	}

	if len(os.Args) == 3 {
		if os.Args[2] == "true"{
			rattrapMode = true
		}
	}

	log.SkipLines(4)
	log.YellowInfo("Running Sulfuras version : ", vers_algo)

	if rattrapMode {
		log.SkipLines(2)
		log.YellowInfo("Rattrap Mode")
	}

	log.SkipLines(2)
	if conf.ProdEnv {
		log.Warning("Prod Environnement")
	} else {
		log.WhiteInfo("Dev Environnement")
	}

	//time.Sleep(5 * time.Second)

	log.SkipLines(2)
	log.WhiteInfo("> Symbol Configuration")
	log.SkipLines(1)

	for _, s := range conf.API.Symbols_t {
		if s.State == "active" {
			log.Info("\tSymbol id : ", s.Id, "	(", s.Name, ") | Status : ", log.GREEN, s.State, log.STOP)
		} else {
			log.Info("\tSymbol id : ", s.Id, "	(", s.Name, ") | Status : ", log.WHITE, s.State, log.STOP)
		}
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
	if rattrapMode {
		log.YellowInfo("Start rattrap")
	} else {
		log.YellowInfo("Start current retrieve")
	}
	log.Info("#############################")

	var bids = make(map[int]tools.SavedBids)
	var trades = make(map[int]map[string]tools.SavedTrades)

	go func() {
		if rattrapMode {
			if err = exec.RattrapNotInactivSymbols(&conf.API, bids); err != nil {
				log.FatalError(err)
				os.Exit(1)
			}
		} else {
			if err = exec.ExecNotInactivSymbols(&conf.API, bids, trades); err != nil {
				log.FatalError(err)
				os.Exit(1)
			}
		}
	}()

	if err = web.StartWebServer(bids, trades, &conf.API); err != nil {
		log.FatalError(err)
		os.Exit(1)
	}

	/*for {
		time.Sleep(24*time.Hour)
	}*/
}
