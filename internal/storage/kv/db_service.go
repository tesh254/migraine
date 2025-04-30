package kv

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
)

// DBService manages database access with proper locking
type DBService struct {
	dbPath       string
	mu           sync.Mutex
	db           *badger.DB
	isOpen       bool
	dbReadOnly   bool
	timeout      time.Duration
	lastActivity time.Time
}

var (
	// Global service instance (singleton pattern)
	globalService *DBService
	once          sync.Once
)

// GetDBService returns the singleton instance of DBService
func GetDBService() *DBService {
	once.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("Failed to get home directory: %v", err))
		}

		dbPath := filepath.Join(homeDir, ".migraine_db")
		globalService = &DBService{
			dbPath:       dbPath,
			isOpen:       false,
			timeout:      5 * time.Second,
			lastActivity: time.Now(),
		}
	})
	return globalService
}

// WithTimeout sets custom timeout for DB operations
func (s *DBService) WithTimeout(duration time.Duration) *DBService {
	s.timeout = duration
	return s
}

// databaseExists checks if the database file exists
func (s *DBService) databaseExists() bool {
	manifestPath := filepath.Join(s.dbPath, "MANIFEST")
	_, err := os.Stat(manifestPath)
	return !os.IsNotExist(err)
}

// initializeEmptyDatabase creates an empty database if it doesn't exist
func (s *DBService) initializeEmptyDatabase() error {
	return s.operationWithTimeout(func(store *Store) error {
		return nil
	}, true, false)
}

// openDB opens the database if not already open
func (s *DBService) openDB(readOnly bool) error {
	if s.isOpen {
		return nil
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(s.dbPath, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	if readOnly {
		// Check if database files exist
		manifestPath := filepath.Join(s.dbPath, "MANIFEST")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return fmt.Errorf("no database found at %s, cannot open in read-only mode", s.dbPath)
		}
	}

	// Configure BadgerDB options
	opts := badger.DefaultOptions(s.dbPath)

	opts.ReadOnly = readOnly

	// Optimize for concurrent access
	opts.NumCompactors = 2
	opts.BlockCacheSize = 50 << 20 // 50MB
	opts.IndexCacheSize = 20 << 20 // 20MB

	logsDir := filepath.Join(s.dbPath, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	var err error

	logFile, err := os.OpenFile(
		filepath.Join(logsDir, "badger.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("couldn't open log file: %v", err)
	}

	opts.Logger = &BadgerLogger{log.New(logFile, "", log.LstdFlags)}

	s.db, err = badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	s.isOpen = true
	s.dbReadOnly = readOnly
	return nil
}

// closeDB closes the database if open
func (s *DBService) closeDB() {
	if s.isOpen && s.db != nil {
		s.db.Close()
		s.isOpen = false
		s.dbReadOnly = false
		s.db = nil
	}
}

// ReadOperation performs a read operation with timeout
func (s *DBService) ReadOperation(operation func(*Store) error) error {
	s.mu.Lock()
	dbExists := s.databaseExists()
	s.mu.Unlock()

	if !dbExists {
		if err := s.initializeEmptyDatabase(); err != nil {
			return fmt.Errorf("failed to initialize store: %v", err)
		}
	}

	return s.operationWithTimeout(operation, false, true)
}

// WriteOperation performs a write operation with timeout
func (s *DBService) WriteOperation(operation func(*Store) error) error {
	return s.operationWithTimeout(operation, true, false)
}

// operationWithTimeout executes a database operation with a timeout
func (s *DBService) operationWithTimeout(operation func(*Store) error, isWrite bool, readOnly bool) error {
	// Create a channel to signal completion
	done := make(chan error, 1)

	// Launch the operation in a goroutine
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if err := s.openDB(readOnly); err != nil {
			done <- err
			return
		}

		// Update last activity time
		s.lastActivity = time.Now()

		// Create a store and execute the operation
		store := New(s.db)
		err := operation(store)

		// Don't close DB here - it stays open for subsequent operations
		done <- err
	}()

	// Wait for operation to complete or timeout
	select {
	case err := <-done:
		return err
	case <-time.After(s.timeout):
		return fmt.Errorf("database operation timed out after %v", s.timeout)
	}
}

// ReadOperationWithContext performs a read operation with context and timeout
func (s *DBService) ReadOperationWithContext(ctx context.Context, operation func(*Store) error) error {
	// Create a channel to signal completion
	done := make(chan error, 1)

	// Launch the operation in a goroutine
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if err := s.openDB(true); err != nil {
			done <- err
			return
		}

		// Update last activity time
		s.lastActivity = time.Now()

		// Create a store and execute the operation
		store := New(s.db)
		err := operation(store)

		// Don't close DB here - it stays open for subsequent operations
		done <- err
	}()

	// Wait for operation to complete, context cancelation, or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(s.timeout):
		return fmt.Errorf("database operation timed out after %v", s.timeout)
	}
}

// ReadOperationWithRetry performs a read operation with retries
func (s *DBService) ReadOperationWithRetry(operation func(*Store) error, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := s.ReadOperation(operation)
		if err == nil {
			return nil
		}

		lastErr = err
		retryDelay := time.Duration(100*(1<<i)) * time.Millisecond // Exponential backoff
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("operation failed after %d retries: %v", maxRetries, lastErr)
}

// CloseDB explicitly closes the database - call this when the app is shutting down
func (s *DBService) CloseDB() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeDB()
}

// AutoCloseDB sets up automatic DB closing after a period of inactivity
func (s *DBService) AutoCloseDB(inactivityPeriod time.Duration) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			s.mu.Lock()
			if s.isOpen && time.Since(s.lastActivity) > inactivityPeriod {
				s.closeDB()
			}
			s.mu.Unlock()
		}
	}()
}
