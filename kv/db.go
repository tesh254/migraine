package kv

import (
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
)

func InitDB(appName string) (*badger.DB, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(homeDir, "."+appName+"_db")
	err = os.MkdirAll(dbPath, 0755)
	if err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions(dbPath)
	return badger.Open(opts)
}
