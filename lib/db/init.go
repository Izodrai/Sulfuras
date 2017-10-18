package db

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"../config/utils"
)

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
