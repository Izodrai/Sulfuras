package main

import (
	"./lib/config"
	"./lib/exec"
	"./lib/log"
	"./lib/tools"
	"./lib/db"
	//"./lib/web"
	l "log"
	"os"
	"path"
	"time"
)

const vers_algo = "v0.0.6"

func main() {

	var err error
	var conf config.Config

	if len(os.Args) != 2 {
		l.Println(log.RED + "Invalid Argument(s)" + log.STOP)
		l.Println(log.RED + "Usuel : ./market-binary config_file" + log.STOP)
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

	if err = db.Init(&conf.API); err != nil {
		log.FatalError(err)
		os.Exit(1)
	}
	defer conf.API.Database.Close()

	if err = conf.LoadSymbolStatus(); err != nil {
		log.FatalError(err)
		os.Exit(1)
	}

	log.SkipLines(4)
	log.YellowInfo("Running Sulfuras version : ", vers_algo)
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

	var bids = make(map[int]map[int]tools.Bid)

	if err = exec.ExecNotInactivSymbols(&conf.API, bids); err != nil {
		log.Error(err)
	}

	//web.StartWebServer(bids, &conf.API)
}
