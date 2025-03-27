package workflow

import (
	"testing"
)

func TestTemplateParser_ParseToWorkflow(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid workflow",
			json: `{
				"name": "Test Workflow",
				"description": "Test Description",
				"steps": [
					{
						"command": "echo 'test'",
						"description": "test step"
					}
				],
				"actions": {
					"test": {
						"command": "echo 'test action'",
						"description": "test action"
					}
				},
				"config": {
					"variables": {},
					"store_variables": true
				}
			}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := NewTemplateParser(tt.json)
			workflow, err := tp.ParseToWorkflow()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && workflow == nil {
				t.Error("ParseToWorkflow() returned nil workflow when no error expected")
			}
		})
	}
}

func TestTemplateParser_ValidateWorkflow(t *testing.T) {
	tests := []struct {
		name    string
		wk      *Workflow
		wantErr bool
	}{
		{
			name: "valid workflow",
			wk: &Workflow{
				Name: "Test",
				Steps: []Atom{
					{Command: "echo 'test'"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			wk: &Workflow{
				Steps: []Atom{
					{Command: "echo 'test'"},
				},
			},
			wantErr: true,
		},
		{
			name: "no steps",
			wk: &Workflow{
				Name: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TemplateParser{}
			err := tp.ValidateWorkflow(tt.wk)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
