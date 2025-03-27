package utils

import (
	"reflect"
	"testing"
)

func TestExtractTemplateVars(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "simple variable",
			content: "Hello {{NAME}}",
			want:    []string{"NAME"},
		},
		{
			name:    "multiple variables",
			content: "{{FIRST}} {{SECOND}} {{THIRD}}",
			want:    []string{"FIRST", "SECOND", "THIRD"},
		},
		{
			name:    "duplicate variables",
			content: "{{VAR}} {{VAR}} {{VAR}}",
			want:    []string{"VAR"},
		},
		{
			name:    "no variables",
			content: "Hello World",
			want:    nil, // Changed from empty slice to nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTemplateVars(tt.content)
			if tt.want == nil && len(got) != 0 {
				t.Errorf("ExtractTemplateVars() = %v, want empty slice", got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractTemplateVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	tests := []struct {
		name    string
		content string
		values  map[string]string
		want    string
		wantErr bool
	}{
		{
			name:    "simple replacement",
			content: "Hello {{NAME}}",
			values:  map[string]string{"NAME": "John"},
			want:    "Hello John",
			wantErr: false,
		},
		{
			name:    "missing variable",
			content: "Hello {{NAME}}",
			values:  map[string]string{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "multiple replacements",
			content: "{{GREETING}} {{NAME}}!",
			values:  map[string]string{"GREETING": "Hello", "NAME": "John"},
			want:    "Hello John!",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceVariables(tt.content, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReplaceVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}
