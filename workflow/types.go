package workflow

type Atom struct {
	Command     string  `json:"command"`
	Description *string `json:"description"`
}

type Config struct {
	Args           map[string][]string `json:"args"`
	StoreVariables bool                `json:"store_variables"`
}

type Workflow struct {
	Name        string  `json:"name"`
	Steps       []Atom  `json:"steps"`
	Description *string `json:"description"`
	Actions     map[string]Atom
	Config      map[string]interface{} `json:"config"`
}

type WorkflowMapper struct {
	WorkflowDir string
}
