package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tesh254/migraine/constants"
	"github.com/tesh254/migraine/utils"
)

type MigrationData struct {
	FileName  string    `db:"file_name" json:"file_name"`
	Checksum  string    `db:"checksum" json:"checksum"`
	AppliedAt time.Time `db:"applied_at" json:"applied_at"`
}

type MRow struct {
	ID        int       `db:"id" json:"id"`
	FileName  string    `db:"file_name" json:"file_name"`
	Checksum  string    `db:"checksum" json:"checksum"`
	AppliedAt time.Time `db:"applied_at" json:"applied_at"`
}

func (c *Core) createMigrationsTable() {
	var config Config
	var fs FS
	migrationsSql := `
		create table if not exists _migraine_migrations (
			id serial primary key,
			file_name varchar(255) not null,
			checksum varchar(255) not null,
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
	log.Println(":::migrations::: | migrations stash created successfully ", utils.CHECK)
}

func (c *Core) createMigration(name string) {
	var config Config
	var fs FS
	name = utils.FormatString(name)
	unixTime := time.Now().Unix()
	prevConfig := config.getConfig()

	filename := fmt.Sprintf("%d_%s.sql", unixTime, name)
	migrationsFolder := fmt.Sprintf("%s/%s", fs.getCurrentDirectory(), *prevConfig.MigrationsPath)

	content := []byte(constants.MIGRATION_CONTENT)

	filepath := migrationsFolder + "/" + filename

	err := os.WriteFile(filepath, content, 0644)

	if err != nil {
		log.Fatalln(":::migrations::: | error creating file: ", err)
	}

	log.Printf(":::migrations::: | %s%s%s created successfully %s", utils.BOLD, filename, utils.RESET, utils.CHECK)
}

func (c *Core) checkForMigration(query string, filename string) bool {
	queryStrip := utils.StripText(query)
	queryChecksum := utils.GenerateChecksum(queryStrip)

	checkQuery := `select exists(select 1 from _migraine_migrations where checksum = $1 or file_name = $2)`

	var exists bool
	err := c.Db.QueryRow(checkQuery, queryChecksum, filename).Scan(&exists)

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

		c.runMigration(filename, utils.GenerateChecksum(utils.StripText(string(content))))
	}
}

func (c *Core) saveMigration(data MigrationData) {
	db := c.Db
	insertQuery := "INSERT INTO _migraine_migrations (file_name, checksum, applied_at) values ($1,$2,$3)"

	_, err := db.Exec(insertQuery, data.FileName, data.Checksum, time.Now().Format(time.RFC3339))

	if err != nil {
		log.Fatalln(":::migrations::: error saving migration to the database: ", err)
	}

	log.Printf(":::migrations::: %s%s%s just applied %s", utils.BOLD, data.FileName, utils.RESET, utils.CHECK)
}

func (c *Core) runMigration(fileName string, checksum string) {
	db := c.Db
	var fs FS
	migrationQuery := fs.migrationSqlFileParser(fileName, false)

	if len(utils.StripText(migrationQuery)) == 0 {
		log.Println(":::migrations::: migration contains no SQL")
		return
	}

	exists := c.checkForMigration(migrationQuery, fileName)

	if exists {
		log.Printf(":::migrations::: %s%s%s already applied %s", utils.BOLD, fileName, utils.RESET, utils.CHECK)
		return
	}

	tx, err := db.Begin()

	if err != nil {
		log.Fatalf(":::migrations::: error beginning transaction: %v\n", err)
		return
	}

	// run migrations
	_, err = tx.Exec(migrationQuery)
	if err != nil {
		log.Fatalf(":::migrations::: error running migration: %v\n", err)
		tx.Rollback()
		return
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatalf(":::migrations::: error committing transaction: %v\n", err)
		tx.Rollback()
		return
	}

	var migrationData MigrationData = MigrationData{
		FileName: fileName,
		Checksum: checksum,
	}

	c.saveMigration(migrationData)
}

func (c *Core) rollback(filename string) {
	var fs FS
	db := c.Db

	// Begin a transaction block
	tx, err := db.Begin()
	if err != nil {
		log.Fatalln(":::migrations::: unable to begin transaction:", err)
	}

	query := fs.migrationSqlFileParser(filename, true)
	// query := fmt.Sprintf(`rollback to savepoint %s;`, savepoint)

	_, err = tx.Exec(query)
	if err != nil {
		// Roll back the transaction and handle the error
		tx.Rollback()
		log.Fatalln(":::migrations::: unable to rollback:", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Fatalln(":::migrations::: unable to commit transaction:", err)
	}
}

func (c *Core) rollbackLastMigration() {
	migration := c.getLastMigration()

	if migration == nil {
		log.Println(":::migrations::: No migration to rollback.")
		return
	}

	c.rollback(migration.FileName)
	c.deleteMigration(migration.ID, migration.FileName)

	log.Printf(":::migrations::: %s%s%s rolled back successfully %s", utils.BOLD, migration.FileName, utils.RESET, utils.CHECK)
}

func (c *Core) getLastMigration() *MRow {
	query := `select * from _migraine_migrations order by applied_at desc limit 1`

	row := c.Db.QueryRow(query)

	var migration MRow

	err := row.Scan(&migration.ID, &migration.FileName, &migration.Checksum, &migration.AppliedAt)

	if err != nil {
		log.Fatalln(":::migrations::: unable to retrieve last migration ", err)
	}

	return &migration
}

func (c *Core) deleteMigration(id int, filename string) {
	var config Config
	var fs FS
	db := c.Db
	query := `delete from _migraine_migrations where id = $1`

	migrationsDir := config.getConfig().MigrationsPath

	if migrationsDir == nil {
		log.Fatalf(":::migrations::: | `%smigrations%s` folder does not exist, please initialize repository\n", utils.BOLD, utils.RESET)
	}

	migrationsFilePath := fmt.Sprintf("%s/%s/%s", fs.getCurrentDirectory(), *migrationsDir, filename)

	err := os.Remove(migrationsFilePath)

	if err != nil {
		log.Fatalln(":::migrations::: unable to delete migrations file")
	}

	_, err = db.Exec(query, id)

	if err != nil {
		log.Fatalln(":::migrations::: unable to delete migration record")
	}
}
