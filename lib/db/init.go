package db

import (
	"../config/utils"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
	"../log"
	"os"
)

var mutex = &sync.Mutex{}

func Init(api_c *utils.API) error {

	var err error

	if api_c.Database, err = sql.Open("mysql", api_c.Database_Info.DSN()); err != nil {
		return err
	}

	if err = api_c.Database.Ping(); err != nil {
		api_c.Database.Close()
		return err
	}

	return nil
}

func KeepOpen(api_c * utils.API) {
	var err error

	for {
		time.Sleep(10 * time.Minute)

		mutex.Lock()

		if err = api_c.Database.Close(); err != nil {
			log.FatalError("KeepOpen Close:", err.Error())
			os.Exit(0)
		}

		if api_c.Database, err = sql.Open("mysql", api_c.Database_Info.DSN()); err != nil {
			log.FatalError("KeepOpen Open:", err.Error())
			os.Exit(0)
		}

		if err = api_c.Database.Ping(); err != nil {
			api_c.Database.Close()
			log.FatalError("KeepOpen Ping:", err.Error())
			os.Exit(0)
		}

		mutex.Unlock()
	}

}