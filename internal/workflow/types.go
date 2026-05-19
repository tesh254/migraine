package workflow

type Atom struct {
	Command     string  `json:"command"`
	Description *string `json:"description"`
	OnFail      string  `json:"on_fail,omitempty"`
	OnSuccess   string  `json:"on_success,omitempty"`
}

type Config struct {
	Variables      map[string]interface{} `json:"variables"`
	StoreVariables bool                   `json:"store_variables"`
	StoreLogs      bool                   `json:"store_logs"`
	Background     bool                   `json:"background"`
	Global         bool                   `json:"global"`
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
