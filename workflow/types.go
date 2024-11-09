package workflow

type Atom struct {
	Command     string  `json:"command"`
	Description *string `json:"description"`
}

type Config struct {
	Variables      map[string]interface{} `json:variables`
	StoreVariables bool                   `json:store_variables`
}

type Workflow struct {
	Name        string  `json:"name"`
	Steps       []Atom  `json:"steps"`
	Description *string `json:"description"`
	Actions     map[string]Atom
	Config      Config `json:"config"`
}

type WorkflowMapper struct {
	WorkflowDir string
}
