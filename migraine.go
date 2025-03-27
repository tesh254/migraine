package main

import (
	"time"

	"github.com/tesh254/migraine/cmd"
	"github.com/tesh254/migraine/internal/storage/kv"
)

func main() {
	// Set up auto DB closure after 5 minutes of inactivity
	kv.GetDBService().AutoCloseDB(5 * time.Minute)

	// Execute the command
	cmd.Execute()

	// Ensure the DB is closed on exit
	kv.GetDBService().CloseDB()
}
