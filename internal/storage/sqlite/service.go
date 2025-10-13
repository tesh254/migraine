package sqlite

import (
	"sync"
	"time"
)

var (
	// Global service instance (singleton pattern)
	globalService *StorageService
	once          sync.Once
)

// GetStorageService returns the singleton instance of StorageService
func GetStorageService() *StorageService {
	once.Do(func() {
		dbService, err := NewDBService("migraine")
		if err != nil {
			panic(err)
		}

		storageService, err := NewStorageService(dbService)
		if err != nil {
			panic(err)
		}

		globalService = storageService
	})
	return globalService
}

// SetStorageService allows setting a custom service instance (useful for testing)
func SetStorageService(service *StorageService) {
	globalService = service
}

// CloseStorageService closes the global storage service
func CloseStorageService() error {
	if globalService != nil {
		return globalService.Close()
	}
	return nil
}

// AutoCloseStorageService closes the storage service after a period of inactivity
func AutoCloseStorageService(inactivityPeriod time.Duration) {
	// In SQLite, we don't need to keep connections open like with Badger,
	// but we'll keep this for API compatibility
	go func() {
		time.Sleep(inactivityPeriod)
		CloseStorageService()
	}()
}
