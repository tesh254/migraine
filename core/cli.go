package core

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/tesh254/migraine/constants"
)

type CLI struct{}

func (cli *CLI) RunCLI() {
	var (
		envName           = flag.String("env", ".env", "Env file name to parse")
		dbVarName         = flag.String("dbVar", "DATABASE_URL", "Database URL environment variable")
		migrationName     = flag.String("new", "", "Name of your migration file")
		migrationsInit    = flag.Bool("init", false, "Initialize go-migraine")
		createMigration   = flag.Bool("migration", false, "Start a migration process")
		runMigrations     = flag.Bool("run", false, "Run all migrations")
		rollbackMigration = flag.Bool("rollback", false, "Rollback last migration")
		help              = flag.Bool("help", false, "Show flag options for migraine")
		version           = flag.Bool("version", false, "Show migraine current installed version")
	)

	flag.Usage = func() {
		fmt.Print(constants.MIGRAINE_ASCII)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Println(constants.CurrentOSWithVersion())
		fmt.Print(constants.MIGRAINE_USAGE)
	}

	flag.Parse()

	if !*migrationsInit && !*createMigration && !*runMigrations && !*rollbackMigration && !*help && !*version {
		flag.Usage()
		return
	}

	var db *sql.DB
	var fs FS
	if !*help && !*version {
		fs.checkIfConfigFileExistsCreateIfNot()
	}
	var config Config
	var core Core
	prevConfig := config.getConfig()
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	switch {
	case *version:
		fmt.Println(constants.CurrentOSWithVersion())
	case *help:
		flag.Usage()
	case *migrationsInit:
		core.getDatabaseURL(*envName, dbVarName)
		fs.checkIfMigrationFolderExists()
		db = core.connection()
		core.createMigrationsTable()
	case *createMigration:
		if *runMigrations && len(*migrationName) == 0 {
			db = core.connection()

			if !prevConfig.IsMigrationsTableCreated {
				core.createMigrationsTable()
			}

			core.runAllMigrations()
		} else if len(*migrationName) > 0 && !*runMigrations {
			db = core.connection()
			core.createMigration(*migrationName)
		} else {
			flag.Usage()
		}
	case *rollbackMigration:
		db = core.connection()
		core.rollbackLastMigration()
	default:
		flag.Usage()
		return
	}
}
