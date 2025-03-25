package kv

// SetValue is a convenience function for setting a value using the global DB service
func SetValue(key string, value interface{}) error {
	return GetDBService().WriteOperation(func(store *Store) error {
		return store.Set(key, value)
	})
}

// GetValue is a convenience function for getting a value using the global DB service
func GetValue(key string, value interface{}) error {
	return GetDBService().ReadOperation(func(store *Store) error {
		return store.Get(key, value)
	})
}

// DeleteValue is a convenience function for deleting a value using the global DB service
func DeleteValue(key string) error {
	return GetDBService().WriteOperation(func(store *Store) error {
		return store.Delete(key)
	})
}

// ListValues is a convenience function for listing values using the global DB service
func ListValues(prefix string) ([]string, error) {
	var result []string
	err := GetDBService().ReadOperation(func(store *Store) error {
		var err error
		result, err = store.List(prefix)
		return err
	})
	return result, err
}

// GetWorkflowSafe retrieves a workflow safely through the DB service
func GetWorkflowSafe(id string) (*Workflow, error) {
	var workflow *Workflow

	err := GetDBService().ReadOperation(func(store *Store) error {
		workflowStore := NewWorkflowStore(store)
		wf, err := workflowStore.GetWorkflow(id)
		if err != nil {
			return err
		}
		workflow = wf
		return nil
	})

	if err != nil {
		return nil, err
	}
	return workflow, nil
}

// ListWorkflowsSafe lists workflows safely through the DB service
func ListWorkflowsSafe() ([]Workflow, error) {
	var workflows []Workflow

	err := GetDBService().ReadOperation(func(store *Store) error {
		workflowStore := NewWorkflowStore(store)
		wfs, err := workflowStore.ListWorkflows()
		if err != nil {
			return err
		}
		workflows = wfs
		return nil
	})

	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// CreateWorkflowSafe creates a workflow safely through the DB service
func CreateWorkflowSafe(id string, workflow Workflow) error {
	return GetDBService().WriteOperation(func(store *Store) error {
		workflowStore := NewWorkflowStore(store)
		return workflowStore.CreateWorkflow(id, workflow)
	})
}

// DeleteWorkflowSafe deletes a workflow safely through the DB service
func DeleteWorkflowSafe(id string) error {
	return GetDBService().WriteOperation(func(store *Store) error {
		workflowStore := NewWorkflowStore(store)
		return workflowStore.DeleteWorkflow(id)
	})
}

// GetTemplateSafe retrieves a template safely through the DB service
func GetTemplateSafe(slug string) (*TemplateItem, error) {
	var template *TemplateItem

	err := GetDBService().ReadOperation(func(store *Store) error {
		templateStore := NewTemplateStoreManager(store)
		tmpl, err := templateStore.GetTemplate(slug)
		if err != nil {
			return err
		}
		template = tmpl
		return nil
	})

	if err != nil {
		return nil, err
	}
	return template, nil
}

// ListTemplatesSafe lists templates safely through the DB service
func ListTemplatesSafe() ([]TemplateItem, error) {
	var templates []TemplateItem

	err := GetDBService().ReadOperation(func(store *Store) error {
		templateStore := NewTemplateStoreManager(store)
		tmpls, err := templateStore.ListTemplates()
		if err != nil {
			return err
		}
		templates = tmpls
		return nil
	})

	if err != nil {
		return nil, err
	}
	return templates, nil
}

// CreateTemplateSafe creates a template safely through the DB service
func CreateTemplateSafe(template TemplateItem) error {
	return GetDBService().WriteOperation(func(store *Store) error {
		templateStore := NewTemplateStoreManager(store)
		return templateStore.CreateTemplate(template)
	})
}

// DeleteTemplateSafe deletes a template safely through the DB service
func DeleteTemplateSafe(slug string) error {
	return GetDBService().WriteOperation(func(store *Store) error {
		templateStore := NewTemplateStoreManager(store)
		return templateStore.DeleteTemplate(slug)
	})
}
