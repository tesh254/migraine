package kv

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
)

// BadgerLogger implements badger.Logger interface
type BadgerLogger struct {
	*log.Logger
}

func (l *BadgerLogger) Errorf(f string, v ...interface{})   { l.Printf("ERROR: "+f, v...) }
func (l *BadgerLogger) Warningf(f string, v ...interface{}) { l.Printf("WARNING: "+f, v...) }
func (l *BadgerLogger) Infof(f string, v ...interface{})    { l.Printf("INFO: "+f, v...) }
func (l *BadgerLogger) Debugf(f string, v ...interface{})   { l.Printf("DEBUG: "+f, v...) }

func InitDB(appName string) (*badger.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create the application directory
	dbPath := filepath.Join(homeDir, "."+appName+"_db")
	err = os.MkdirAll(dbPath, 0755)
	if err != nil {
		return nil, err
	}

	// Create logs directory
	logsDir := filepath.Join(dbPath, "logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		return nil, err
	}

	// Set up the log file
	logFile, err := os.OpenFile(
		filepath.Join(logsDir, "badger.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't open log file: %v", err)
	}

	// Configure Badger options
	opts := badger.DefaultOptions(dbPath)

	// Create our custom logger that implements the badger.Logger interface
	logger := &BadgerLogger{Logger: log.New(logFile, "", log.LstdFlags)}
	opts.Logger = logger

	return badger.Open(opts)
}
