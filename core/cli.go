package core

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
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
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if !*migrationsInit && !*createMigration && !*runMigrations && !*rollbackMigration {
		flag.Usage()
		return
	}

	var db *sql.DB
	var fs FS
	fs.checkIfConfigFileExistsCreateIfNot()
	var config Config
	var core Core
	prevConfig := config.getConfig()
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	if *migrationsInit {
		core.getDatabaseURL(*envName, dbVarName)
		fs.checkIfMigrationFolderExists()
		db = core.connection()
		core.createMigrationsTable()
	} else if *createMigration {
		if *runMigrations && len(*migrationName) == 0 {
			db = core.connection()

			if !prevConfig.IsMigrationsTableCreated {
				core.createMigrationsTable()
			}

			core.runAllMigrations()
		} else if len(*migrationName) > 0 && !*runMigrations {
			db = core.connection()
			fmt.Println(*migrationName)
			core.createMigration(*migrationName)
		} else {
			flag.Usage()
		}
	} else if *rollbackMigration {
		db = core.connection()
		core.rollbackLastMigration()
	} else {
		flag.Usage()
		return
	}
}
