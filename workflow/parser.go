package workflow

import (
	"encoding/json"
	"fmt"
)

type TemplateParser struct {
	rawJSON string
}

func NewTemplateParser(jsonStr string) *TemplateParser {
	return &TemplateParser{
		rawJSON: jsonStr,
	}
}

func (tp *TemplateParser) ParseToWorkflow() (*Workflow, error) {
	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(tp.rawJSON), &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	workflow := &Workflow{}

	if name, ok := rawMap["name"].(string); ok {
		workflow.Name = name
	}

	if desc, ok := rawMap["description"].(string); ok {
		workflow.Description = &desc
	}

	if stepsRaw, ok := rawMap["steps"].([]interface{}); ok {
		workflow.Steps = make([]Atom, 0, len(stepsRaw))
		for _, stepRaw := range stepsRaw {
			if stepMap, ok := stepRaw.(map[string]interface{}); ok {
				atom := Atom{}

				if cmd, ok := stepMap["command"].(string); ok {
					atom.Command = cmd
				}

				if desc, ok := stepMap["description"].(string); ok {
					atom.Description = &desc
				}

				workflow.Steps = append(workflow.Steps, atom)
			}
		}
	}

	if actionsRaw, ok := rawMap["actions"].(map[string]interface{}); ok {
		workflow.Actions = make(map[string]Atom)
		for key, actionRaw := range actionsRaw {
			if actionMap, ok := actionRaw.(map[string]interface{}); ok {
				atom := Atom{}

				if cmd, ok := actionMap["command"].(string); ok {
					atom.Command = cmd
				}

				if desc, ok := actionMap["description"].(string); ok {
					atom.Description = &desc
				}

				workflow.Actions[key] = atom
			}
		}
	}

	if configRaw, ok := rawMap["config"].(map[string]interface{}); ok {
		config := Config{}

		// Parse variables
		if vars, ok := configRaw["variables"].(map[string]interface{}); ok {
			config.Variables = vars
		}

		// Parse store_variables
		if storeVars, ok := configRaw["store_variables"].(bool); ok {
			config.StoreVariables = storeVars
		}

		workflow.Config = config
	}

	return workflow, nil
}

func (tp *TemplateParser) ValidateWorkflow(wk *Workflow) error {
	if wk.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(wk.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	for i, step := range wk.Steps {
		if step.Command == "" {
			return fmt.Errorf("step %d must have a command", i+1)
		}
	}

	return nil
}
