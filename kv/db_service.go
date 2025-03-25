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

// openDB opens the database if not already open
func (s *DBService) openDB() error {
	if s.isOpen {
		return nil
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(s.dbPath, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// Configure BadgerDB options
	opts := badger.DefaultOptions(s.dbPath)

	// Optimize for concurrent access
	opts.NumCompactors = 2
	opts.BlockCacheSize = 50 << 20 // 50MB
	opts.IndexCacheSize = 20 << 20 // 20MB

	logsDir := filepath.Join(s.dbPath, "logs")

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
	return nil
}

// closeDB closes the database if open
func (s *DBService) closeDB() {
	if s.isOpen && s.db != nil {
		s.db.Close()
		s.isOpen = false
		s.db = nil
	}
}

// ReadOperation performs a read operation with timeout
func (s *DBService) ReadOperation(operation func(*Store) error) error {
	return s.operationWithTimeout(operation, false)
}

// WriteOperation performs a write operation with timeout
func (s *DBService) WriteOperation(operation func(*Store) error) error {
	return s.operationWithTimeout(operation, true)
}

// operationWithTimeout executes a database operation with a timeout
func (s *DBService) operationWithTimeout(operation func(*Store) error, isWrite bool) error {
	// Create a channel to signal completion
	done := make(chan error, 1)

	// Launch the operation in a goroutine
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if err := s.openDB(); err != nil {
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

		if err := s.openDB(); err != nil {
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
