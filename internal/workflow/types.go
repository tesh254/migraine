package workflow

type Atom struct {
	Command     string  `json:"command"`
	Description *string `json:"description"`
}

type Config struct {
	Variables      map[string]interface{} `json:"variables"`
	StoreVariables bool                   `json:"store_variables"`
}

type Workflow struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	PreChecks   []Atom          `json:"pre_checks"`
	Steps       []Atom          `json:"steps"`
	Description *string         `json:"description"`
	Actions     map[string]Atom `json:"actions"`
	Config      Config          `json:"config"`
	UsesSudo    bool            `json:"uses_sudo"`
}

type WorkflowMapper struct {
	WorkflowDir string
}
