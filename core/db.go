package core

import (
	"log"

	"database/sql"

	_ "github.com/lib/pq"
)

func (c *Core) connection() *sql.DB {
	var db *sql.DB
	var config Config

	configuration := config.getConfig()

	var dbUrl string

	if configuration.DbUrl != nil {
		dbUrl = *configuration.DbUrl
	} else {
		log.Fatalln(":::db::: url is not present please initialize migrations update")
	}

	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		log.Fatalf(":::db::: | unable to reach the database %v\n", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatalf(":::db::: | unable to connect to the datbase %v\n", err)
	}

	c.Db = db

	return db
}
