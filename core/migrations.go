package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tesh254/go-migraine/utils"
)

type MigrationData struct {
	FileName      string    `db:"file_name" json:"file_name"`
	Checksum      string    `db:"checksum" json:"checksum"`
	TransactionID int       `db:"transaction_id" json:"transaction_id"`
	AppliedAt     time.Time `db:"applied_at" json:"applied_at"`
}

type MRow struct {
	ID            int       `db:"id" json:"id"`
	FileName      string    `db:"file_name" json:"file_name"`
	Checksum      string    `db:"checksum" json:"checksum"`
	TransactionID int       `db:"transaction_id" json:"transaction_id"`
	AppliedAt     time.Time `db:"applied_at" json:"applied_at"`
}

func (c *Core) createMigrationsTable() {
	var config Config
	var fs FS
	migrationsSql := `
		create table if not exists migrations (
			id serial primary key,
			file_name varchar(255) not null,
			checksum varchar(255) not null,
			transaction_id int not null,
			applied_at timestamp with time zone default current_timestamp
		);
	`
	_, err := c.Db.Exec(migrationsSql)

	if err != nil {
		log.Fatalln(":::migrations::: | failed to create migrations table: ", err)
	}

	doesFolderExist := fs.checkIfMigrationFolderExists()
	migrationsPath := "migrations"
	config.updateMigrationsConfig(doesFolderExist, true, &migrationsPath)
	log.Println(":::migrations::: | table created successfully ", utils.CHECK)
}

func (c *Core) createMigration(name string) {
	var config Config
	var fs FS
	name = utils.FormatString(name)
	unixTime := time.Now().Unix()
	prevConfig := config.getConfig()

	filename := fmt.Sprintf("%d_%s.sql", unixTime, name)
	migrationsFolder := fmt.Sprintf("%s/%s", fs.getCurrentDirectory(), *prevConfig.MigrationsPath)

	content := []byte("begin;\ncommit;")

	filepath := migrationsFolder + "/" + filename

	err := os.WriteFile(filepath, content, 0644)

	if err != nil {
		log.Fatalln(":::migrations::: | error creating file: ", err)
	}

	log.Printf(":::migrations::: | %s%s%s created successfully %s", utils.BOLD, filename, utils.RESET, utils.CHECK)
}

func (c *Core) checkForMigration(query string) bool {
	queryStrip := utils.StripText(query)
	queryChecksum := utils.GenerateChecksum(queryStrip)

	checkQuery := `select exists(select 1 from migrations where checksum = $1)`

	var exists bool
	err := c.Db.QueryRow(checkQuery, queryChecksum).Scan(&exists)

	if err != nil {
		log.Fatalf(":::migrations::: | error checking for migration: %v\n", err)
	}

	return exists
}

func (c *Core) runAllMigrations() {
	var config Config
	var fs FS
	migrationsPath := fmt.Sprintf("%s/%s", fs.getCurrentDirectory(), *config.getConfig().MigrationsPath)

	files, err := os.ReadDir(migrationsPath)

	if err != nil {
		log.Fatal(":::migrations::: | error reading migrations directory: ", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()

		filePath := filepath.Join(migrationsPath, filename)
		content, err := os.ReadFile(filePath)

		if err != nil {
			log.Fatalf(":::migrations::: | error reading file '%s': '%v'\n", filename, err)
		}

		c.runMigration(filename, utils.GenerateChecksum(utils.StripText(string(content))), string(content))
	}
}

func (c *Core) saveMigration(data MigrationData) {
	db := c.Db
	insertQuery := "INSERT INTO migrations (file_name, checksum, transaction_id, applied_at) values ($1,$2,$3,$4)"

	_, err := db.Exec(insertQuery, data.FileName, data.Checksum, data.TransactionID, time.Now().Format(time.RFC3339))

	if err != nil {
		log.Fatalln(":::migrations::: error saving migration to the database: ", err)
	}

	log.Printf(":::migrations::: %s%s%s just applied %s", utils.BOLD, data.FileName, utils.RESET, utils.CHECK)
}

func (c *Core) runMigration(fileName string, checksum string, migrationQuery string) {
	db := c.Db
	if migrationQuery == "begin;\ncommit;" {
		log.Println(":::migrations::: migration contains no sql")
		return
	}

	exists := c.checkForMigration(migrationQuery)

	if exists {
		log.Printf(":::migrations::: %s%s%s applied %s", utils.BOLD, fileName, utils.RESET, utils.CHECK)
		return
	}

	tx, err := db.Begin()

	if err != nil {
		log.Fatalln(err)
	}

	_, err = tx.Exec(migrationQuery)

	if err != nil {
		fmt.Println(migrationQuery)
		log.Fatalln(":::migrations::: error running migration: ", err)
	}

	var txID int
	err = tx.QueryRow("select txid_current();").Scan(&txID)
	if err != nil {
		log.Fatalln(":::migrations::: error getting transaction ID: ", err)
	}

	var migrationData MigrationData = MigrationData{
		FileName:      fileName,
		Checksum:      checksum,
		TransactionID: txID,
	}

	c.saveMigration(migrationData)
}

func (c *Core) rollback() {
	db := c.Db
	migration := c.getLastMigration()

	query := fmt.Sprintf(`
		select pg_terminate_backend (pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.backend_xid = '%d';
	`, migration.TransactionID)

	_, err := db.Exec(query)

	if err != nil {
		log.Fatalln(":::migrations::: unable to rollback")
	}

	c.deleteMigration(migration.ID, migration.FileName)

	log.Printf(":::migrations::: %s%s%s rolled back successfully %s", utils.BOLD, migration.FileName, utils.RESET, utils.CHECK)
}

func (c *Core) getLastMigration() *MRow {
	query := `select * from migrations order by applied_at desc limit 1`

	row := c.Db.QueryRow(query)

	var migration MRow

	err := row.Scan(&migration.ID, &migration.FileName, &migration.Checksum, &migration.TransactionID, &migration.AppliedAt)

	if err != nil {
		log.Fatalln(":::migrations::: unable to retrieve last migration")
	}

	return &migration
}

func (c *Core) deleteMigration(id int, filename string) {
	var config Config
	var fs FS
	db := c.Db
	query := `delete from migrations where id = $1`

	_, err := db.Exec(query, id)

	if err != nil {
		log.Fatalln(":::migrations::: unable to delete migration record")
	}

	migrationsPath := fmt.Sprintf("%s/%s", fs.getCurrentDirectory(), *config.getConfig().MigrationsPath)

	filePath := fmt.Sprintf(migrationsPath, filename)

	err = os.Remove(filePath)

	if err != nil {
		log.Fatalln(":::migrations::: unable to delete migrations file")
	}
}
