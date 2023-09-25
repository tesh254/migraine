package core

import (
	"fmt"
	"log"

	"database/sql"

	_ "github.com/lib/pq"
)

func (c *Core) connection() *sql.DB {
	var db *sql.DB

	fmt.Println(c.DbUrl)

	db, err := sql.Open("postgres", c.DbUrl)

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
