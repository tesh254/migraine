package main

import (
	"log"
	"time"

	"github.com/tesh254/migraine/cmd"
	"github.com/tesh254/migraine/internal/storage/sqlite"
	"github.com/tesh254/migraine/internal/version"
)

var BuildVersion string
var BuildCommit string
var BuildDate string

func init() {
	// Set up build version info from ldflags in the version package
	if BuildVersion != "" {
		version.BuildVersion = BuildVersion
	}
	if BuildCommit != "" {
		version.BuildCommit = BuildCommit
	}
	if BuildDate != "" {
		version.BuildDate = BuildDate
	}
}

func main() {
	// Run migration from old Badger storage to new SQLite storage
	if err := sqlite.RunInitialMigration(); err != nil {
		log.Printf("Warning: failed to run migration: %v", err)
		// Continue execution even if migration fails
	}

	// Set up auto DB closure after 5 minutes of inactivity
	sqlite.AutoCloseStorageService(5 * time.Minute)

	// Execute the command
	cmd.Execute()

	// Ensure the DB is closed on exit
	sqlite.CloseStorageService()
}
