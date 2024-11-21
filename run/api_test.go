package run

import (
	"bytes"
	"testing"
)

func TestFormattedWriter_Write(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantLen int
		wantErr bool
	}{
		{
			name:    "simple write",
			input:   []byte("test message"),
			wantLen: 12,
			wantErr: false,
		},
		{
			name:    "empty write",
			input:   []byte{},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			fw := NewFormattedWriter(buf)
			gotLen, err := fw.Write(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormattedWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLen != tt.wantLen {
				t.Errorf("FormattedWriter.Write() = %v, want %v", gotLen, tt.wantLen)
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "valid command",
			command: "echo 'test'",
			wantErr: false,
		},
		{
			name:    "invalid command",
			command: "nonexistentcommand",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecuteCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
