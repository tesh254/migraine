package core

import "database/sql"

type Core struct {
	DbUrl string
	Db    *sql.DB
}
