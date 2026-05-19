package lsp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tesh254/migraine/internal/workflow"
)

func TestValidateDocument_Valid(t *testing.T) {
	text := `
metadata {
    name = "test-workflow"
}
workflow {
    steps [
        {
            cmd = "echo hello"
        }
    ]
}
config {
    store_logs = false
}
`
	diags := validateDocument(text)
	if len(diags) != 0 {
		for _, d := range diags {
			t.Errorf("Unexpected diagnostic: %s", d.Message)
		}
	}
}

func TestValidateDocument_InvalidBlock(t *testing.T) {
	text := `
unknown_block {
    foo = "bar"
}
`
	diags := validateDocument(text)
	if len(diags) == 0 {
		t.Fatal("Expected diagnostic for unknown block, got none")
	}
	if !strings.Contains(diags[0].Message, "unknown block") {
		t.Errorf("Expected 'unknown block' error, got: %s", diags[0].Message)
	}
}

func TestValidateDocument_SyntaxError(t *testing.T) {
	text := `
metadata {
    name = 
}
`
	diags := validateDocument(text)
	if len(diags) == 0 {
		t.Fatal("Expected diagnostic for syntax error, got none")
	}
}

func TestCompletion_ReturnsItems(t *testing.T) {
	s := NewServer()
	params, _ := json.Marshal(CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.mg"},
		Position:     Position{Line: 0, Character: 0},
	})
	s.docs["file:///test.mg"] = ""

	result, err := s.handleCompletion(params)
	if err != nil {
		t.Fatalf("handleCompletion error: %v", err)
	}
	cr := result.(CompletionResult)
	if len(cr.Items) == 0 {
		t.Fatal("Expected completion items, got none")
	}

	labels := make(map[string]bool)
	for _, item := range cr.Items {
		labels[item.Label] = true
	}

	expected := []string{"metadata", "variables", "workflow", "config",
		"steps", "pre_checks", "actions",
		"cmd", "desc", "on_fail", "on_success",
		"store_variables", "store_logs", "background", "global",
		"true", "false", "args:", "env:", "vault:", "action:", "run:"}

	for _, e := range expected {
		if !labels[e] {
			t.Errorf("Missing completion item: %s", e)
		}
	}
}

func TestHover_KnownKeyword(t *testing.T) {
	s := NewServer()
	s.docs["file:///test.mg"] = "metadata {\n    name = \"test\"\n}\n"

	params, _ := json.Marshal(HoverParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.mg"},
		Position:     Position{Line: 0, Character: 2},
	})
	result, err := s.handleHover(params)
	if err != nil {
		t.Fatalf("handleHover error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected hover result for 'metadata', got nil")
	}
	hr := result.(HoverResult)
	if !strings.Contains(hr.Contents.Value, "metadata") {
		t.Errorf("Hover content should mention 'metadata', got: %s", hr.Contents.Value)
	}
}

func TestHover_UnknownWord(t *testing.T) {
	s := NewServer()
	s.docs["file:///test.mg"] = "xyzzy {\n}\n"

	params, _ := json.Marshal(HoverParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.mg"},
		Position:     Position{Line: 0, Character: 2},
	})
	result, _ := s.handleHover(params)
	if result != nil {
		t.Errorf("Expected nil hover for unknown word, got: %v", result)
	}
}

func TestDocumentSymbols(t *testing.T) {
	text := `
metadata {
    name = "test"
}
workflow {
    steps [
        {
            cmd = "echo"
        }
    ]
}
config {
    store_logs = true
}
`
	symbols := extractDocumentSymbols(text)
	names := make(map[string]bool)
	for _, sym := range symbols {
		names[sym.Name] = true
	}

	for _, expected := range []string{"metadata", "workflow", "config"} {
		if !names[expected] {
			t.Errorf("Missing document symbol: %s", expected)
		}
	}
}

func TestTokenizeDocument(t *testing.T) {
	text := `metadata {
    name = "test"
}
`
	tokens := tokenizeDocument(text)
	if len(tokens) == 0 {
		t.Fatal("Expected semantic tokens, got none")
	}
	if len(tokens)%5 != 0 {
		t.Errorf("Token data length should be multiple of 5, got %d", len(tokens))
	}
}

func TestSemanticTokens_BlockNames(t *testing.T) {
	text := `metadata {
    name = "test"
}
`
	tokens := tokenizeDocument(text)

	foundMetadata := false
	for i := 0; i < len(tokens); i += 5 {
		tokenType := tokens[i+3]
		if tokenType == 3 {
			foundMetadata = true
		}
	}
	if !foundMetadata {
		t.Error("Expected 'metadata' to be tokenized as block/section (type 3)")
	}
}

func TestInitialize(t *testing.T) {
	s := NewServer()
	params, _ := json.Marshal(InitializeParams{RootURI: "file:///project"})
	result, err := s.handleInitialize(params)
	if err != nil {
		t.Fatalf("handleInitialize error: %v", err)
	}
	ir := result.(InitializeResult)
	if !ir.Capabilities.HoverProvider {
		t.Error("Expected HoverProvider to be true")
	}
	if !ir.Capabilities.DocumentSymbolProvider {
		t.Error("Expected DocumentSymbolProvider to be true")
	}
	if !ir.Capabilities.SemanticTokensProvider.Full {
		t.Error("Expected SemanticTokens Full to be true")
	}
	if ir.Capabilities.TextDocumentSync.Change != 1 {
		t.Error("Expected TextDocumentSync Change=1 (full sync)")
	}
}

func TestDidOpenAndClose(t *testing.T) {
	s := NewServer()

	openParams, _ := json.Marshal(DidOpenParams{
		TextDocument: TextDocumentItem{URI: "file:///test.mg", Version: 1, Text: "test"},
	})
	if err := s.handleDidOpen(openParams); err != nil {
		t.Fatalf("handleDidOpen error: %v", err)
	}

	s.mu.Lock()
	_, ok := s.docs["file:///test.mg"]
	s.mu.Unlock()
	if !ok {
		t.Fatal("Document should be in docs after open")
	}

	closeParams, _ := json.Marshal(DidCloseParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.mg"},
	})
	if err := s.handleDidClose(closeParams); err != nil {
		t.Fatalf("handleDidClose error: %v", err)
	}

	s.mu.Lock()
	_, ok = s.docs["file:///test.mg"]
	s.mu.Unlock()
	if ok {
		t.Fatal("Document should be removed from docs after close")
	}
}

func TestGetWordAtPosition(t *testing.T) {
	text := "metadata {\n    name = \"test\"\n}"

	tests := []struct {
		line, char int
		expected   string
	}{
		{0, 2, "metadata"},
		{0, 7, "metadata"},
		{1, 4, "name"},
		{1, 5, "name"},
		{1, 10, ""},
	}

	for _, tt := range tests {
		result := getWordAtPosition(text, tt.line, tt.char)
		if result != tt.expected {
			t.Errorf("getWordAtPosition(%d, %d) = %q, want %q", tt.line, tt.char, result, tt.expected)
		}
	}
}

func TestIdentTokenType(t *testing.T) {
	tests := []struct {
		input  string
		expect int
	}{
		{"metadata", 3},
		{"workflow", 3},
		{"variables", 3},
		{"config", 3},
		{"steps", 3},
		{"pre_checks", 3},
		{"actions", 3},
		{"cmd", 8},
		{"desc", 8},
		{"on_fail", 8},
		{"store_logs", 8},
		{"name", 8},
		{"foobar", 0},
	}

	for _, tt := range tests {
		result := identTokenType(tt.input)
		if result != tt.expect {
			t.Errorf("identTokenType(%q) = %d, want %d", tt.input, result, tt.expect)
		}
	}
}

func TestValidateDocument_WithLexerBug(t *testing.T) {
	text := `
metadata {
    name = "test"
}
variables {
    timeout = 300
}
workflow {
    steps [
        {
            cmd = "echo hello"
        }
    ]
}
config {
    store_logs = false
}
`
	diags := validateDocument(text)
	if len(diags) != 0 {
		for _, d := range diags {
			t.Errorf("Unexpected diagnostic: %s", d.Message)
		}
	}
}

func TestValidateDocument_BacktickStrings(t *testing.T) {
	text := `
metadata {
    name = "test"
}
workflow {
    steps [
        {
            cmd = ` + "`" + `echo "hello world"` + "`" + `
        }
    ]
}
config {
    store_logs = false
}
`
	diags := validateDocument(text)
	if len(diags) != 0 {
		for _, d := range diags {
			t.Errorf("Unexpected diagnostic: %s", d.Message)
		}
	}
}

func TestValidateDocument_SampleFile(t *testing.T) {
	text := `# Migraine workflow: Deploy a web application
metadata {
    name = "deploy-app"
    desc = "Build and deploy the web application to staging"
}

variables {
    app_name = "args:APP_NAME"
    env = "args:ENV"
}

workflow {
    pre_checks [
        {
            cmd = ` + "`" + `docker info` + "`" + `
            desc = "Verify Docker daemon is running"
            on_fail = "action:notify_failure"
        }
    ]

    steps [
        {
            cmd = "echo hello"
        }
    ]

    actions {
        notify_failure {
            cmd = "echo failed"
        }
    }
}

config {
    store_variables = true
    store_logs = true
    background = false
    global = false
}
`
	diags := validateDocument(text)
	if len(diags) != 0 {
		for _, d := range diags {
			t.Errorf("Unexpected diagnostic for sample file: %s", d.Message)
		}
	}
}

func TestNewLexerIntegration(t *testing.T) {
	text := `metadata {
    name = "test"
}
`
	lexer := workflow.NewLexer(strings.NewReader(text))

	tok, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("Unexpected lexer error: %v", err)
	}
	if tok.Type != workflow.TokenIdent || tok.Literal != "metadata" {
		t.Errorf("Expected 'metadata' ident token, got %+v", tok)
	}
}