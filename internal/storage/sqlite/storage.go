package sqlite

// StorageService provides a unified interface for all storage operations
type StorageService struct {
	dbService     *DBService
	workflowStore *WorkflowStore
	vaultStore    *VaultStore
	runStore      *RunStore
}

func NewStorageService(dbService *DBService) (*StorageService, error) {
	service := &StorageService{
		dbService:     dbService,
		workflowStore: NewWorkflowStore(dbService),
		vaultStore:    NewVaultStore(dbService),
		runStore:      NewRunStore(dbService),
	}

	return service, nil
}

func (s *StorageService) WorkflowStore() *WorkflowStore {
	return s.workflowStore
}

func (s *StorageService) VaultStore() *VaultStore {
	return s.vaultStore
}

func (s *StorageService) RunStore() *RunStore {
	return s.runStore
}

func (s *StorageService) Close() error {
	return s.dbService.Close()
}

// GetDB returns the underlying database connection for direct queries if needed
func (s *StorageService) GetDB() *DBService {
	return s.dbService
}
