package storage

type WorkflowStore interface {
	CreateWorkflow(id string, workflow interface{}) error
	GetWorkflow(id string) (interface{}, error)
	UpdateWorkflow(id string, workflow interface{}) error
	DeleteWorkflow(id string) error
	ListWorkflows() ([]interface{}, error)
}

type TemplateStore interface {
	CreateTemplate(template interface{}) error
	GetTemplate(slug string) (interface{}, error)
	UpdateTemplate(template interface{}) error
	DeleteTemplate(slug string) error
	ListTemplates() ([]interface{}, error)
}
