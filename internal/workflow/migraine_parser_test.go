package workflow

import (
	"os"
	"strings"
	"testing"
)

func TestMigraineParser_Integration(t *testing.T) {
	script := `
metadata {
    name = "integration-test-workflow"
    desc = "Test workflow description"
}

variables {
    project_path = "args:APP_NAME"
    editor = "args:EDITOR"
    debug = true
}

workflow {
    pre_checks [
        {
            cmd = "echo 'checking...'"
            desc = "pre-check description"
            on_fail = "action:fail_handler"
        }
    ]

    steps [
        {
            cmd = "echo 'step 1'"
            desc = "step 1 description"
        },
        {
            cmd = "echo 'step 2'"
        }
    ]

    actions {
        fail_handler {
            cmd = "echo 'failed'"
            desc = "handler for failure"
        }
        
        custom_action {
            cmd = "echo 'custom'"
            on_success = "run:echo 'success'"
        }
    }
}

config {
    store_logs = true
    background = false
    global = true
}
`

	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Verify Metadata
	if wf.Name != "integration-test-workflow" {
		t.Errorf("Expected name 'integration-test-workflow', got '%s'", wf.Name)
	}
	if wf.Description == nil || *wf.Description != "Test workflow description" {
		t.Errorf("Expected description 'Test workflow description', got '%v'", wf.Description)
	}

	// Verify Variables
	if val, ok := wf.Config.Variables["project_path"]; !ok || val != "args:APP_NAME" {
		t.Errorf("Expected variable project_path='args:APP_NAME', got '%v'", val)
	}
	if val, ok := wf.Config.Variables["debug"]; !ok || val != true {
		t.Errorf("Expected variable debug=true, got '%v'", val)
	}

	// Verify PreChecks
	if len(wf.PreChecks) != 1 {
		t.Fatalf("Expected 1 pre-check, got %d", len(wf.PreChecks))
	}
	if wf.PreChecks[0].Command != "echo 'checking...'" {
		t.Errorf("Expected pre-check command 'echo 'checking...'', got '%s'", wf.PreChecks[0].Command)
	}
	if wf.PreChecks[0].OnFail != "action:fail_handler" {
		t.Errorf("Expected pre-check on_fail 'action:fail_handler', got '%s'", wf.PreChecks[0].OnFail)
	}

	// Verify Steps
	if len(wf.Steps) != 2 {
		t.Fatalf("Expected 2 steps, got %d", len(wf.Steps))
	}
	if wf.Steps[0].Command != "echo 'step 1'" {
		t.Errorf("Expected step 1 command 'echo 'step 1'', got '%s'", wf.Steps[0].Command)
	}
	if wf.Steps[1].Command != "echo 'step 2'" {
		t.Errorf("Expected step 2 command 'echo 'step 2'', got '%s'", wf.Steps[1].Command)
	}

	// Verify Actions
	if len(wf.Actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(wf.Actions))
	}
	if action, ok := wf.Actions["fail_handler"]; !ok {
		t.Error("Expected action 'fail_handler' not found")
	} else {
		if action.Command != "echo 'failed'" {
			t.Errorf("Expected fail_handler command 'echo 'failed'', got '%s'", action.Command)
		}
	}
	if action, ok := wf.Actions["custom_action"]; !ok {
		t.Error("Expected action 'custom_action' not found")
	} else {
		if action.OnSuccess != "run:echo 'success'" {
			t.Errorf("Expected custom_action on_success 'run:echo 'success'', got '%s'", action.OnSuccess)
		}
	}

	// Verify Config
	if !wf.Config.StoreLogs {
		t.Error("Expected StoreLogs to be true")
	}
	if wf.Config.Background {
		t.Error("Expected Background to be false")
	}
	if !wf.Config.Global {
		t.Error("Expected Global to be true")
	}

	// Verify Conversion to YAMLWorkflow
	yamlWf := ConvertInternalToYAML(wf, "")

	if yamlWf.Name != "integration-test-workflow" {
		t.Errorf("YAMLWorkflow: Expected name 'integration-test-workflow', got '%s'", yamlWf.Name)
	}

	// Check PreChecks conversion
	if len(yamlWf.PreChecks) != 1 {
		t.Fatalf("YAMLWorkflow: Expected 1 pre-check, got %d", len(yamlWf.PreChecks))
	}
	if yamlWf.PreChecks[0].Command != "echo 'checking...'" {
		t.Errorf("YAMLWorkflow: Expected pre-check command, got '%s'", yamlWf.PreChecks[0].Command)
	}

	// Check Config Variables conversion
	if val, ok := yamlWf.Config.Variables["project_path"]; !ok || val != "args:APP_NAME" {
		t.Errorf("YAMLWorkflow: Expected variable project_path, got '%v'", val)
	}
}

func TestMigraineParser_SampleFile(t *testing.T) {
	data, err := os.ReadFile("../../examples/deploy-app.mg")
	if err != nil {
		t.Skipf("Sample file not found: %v", err)
	}

	parser, err := NewMigraineParserFromReader(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse sample file: %v", err)
	}

	if wf.Name != "deploy-app" {
		t.Errorf("Expected name 'deploy-app', got '%s'", wf.Name)
	}
	if wf.Description == nil || *wf.Description != "Build and deploy the web application to staging" {
		t.Errorf("Expected description, got '%v'", wf.Description)
	}

	if len(wf.PreChecks) != 2 {
		t.Fatalf("Expected 2 pre-checks, got %d", len(wf.PreChecks))
	}
	if wf.PreChecks[0].OnFail != "action:notify_failure" {
		t.Errorf("Expected pre-check on_fail 'action:notify_failure', got '%s'", wf.PreChecks[0].OnFail)
	}

	if len(wf.Steps) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(wf.Steps))
	}
	if wf.Steps[2].OnSuccess != "action:notify_success" {
		t.Errorf("Expected step 3 on_success 'action:notify_success', got '%s'", wf.Steps[2].OnSuccess)
	}
	if wf.Steps[2].OnFail != "action:rollback" {
		t.Errorf("Expected step 3 on_fail 'action:rollback', got '%s'", wf.Steps[2].OnFail)
	}

	if len(wf.Actions) != 4 {
		t.Fatalf("Expected 4 actions, got %d", len(wf.Actions))
	}
	rollback, ok := wf.Actions["rollback"]
	if !ok {
		t.Fatal("Expected action 'rollback' not found")
	}
	if rollback.OnFail != "action:notify_failure" {
		t.Errorf("Expected rollback on_fail 'action:notify_failure', got '%s'", rollback.OnFail)
	}
	if !wf.Config.StoreVariables {
		t.Error("Expected StoreVariables to be true")
	}
	if !wf.Config.StoreLogs {
		t.Error("Expected StoreLogs to be true")
	}
	if wf.Config.Background {
		t.Error("Expected Background to be false")
	}
	if wf.Config.Global {
		t.Error("Expected Global to be false")
	}
}

func TestMigraineParser_BacktickStrings(t *testing.T) {
	script := `
metadata {
    name = "backtick-test"
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
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if wf.Steps[0].Command != `echo "hello world"` {
		t.Errorf("Expected backtick string content, got '%s'", wf.Steps[0].Command)
	}
}

func TestMigraineParser_Comments(t *testing.T) {
	script := `# This is a comment
metadata {
    # Inline comment
    name = "comment-test"
}
workflow {
    steps [
        {
            # Another comment
            cmd = "echo hello"
        }
    ]
}
config {
    store_logs = false
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if wf.Name != "comment-test" {
		t.Errorf("Expected name 'comment-test', got '%s'", wf.Name)
	}
	if len(wf.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(wf.Steps))
	}
}

func TestMigraineParser_NoPreChecksNoActions(t *testing.T) {
	script := `
metadata {
    name = "minimal"
}
workflow {
    steps [
        {
            cmd = "echo minimal"
        }
    ]
}
config {
    store_logs = false
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(wf.PreChecks) != 0 {
		t.Errorf("Expected 0 pre-checks, got %d", len(wf.PreChecks))
	}
	if len(wf.Actions) != 0 {
		t.Errorf("Expected 0 actions, got %d", len(wf.Actions))
	}
	if len(wf.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(wf.Steps))
	}
	if wf.Steps[0].Command != "echo minimal" {
		t.Errorf("Expected command 'echo minimal', got '%s'", wf.Steps[0].Command)
	}
}

func TestMigraineParser_EscapedQuotes(t *testing.T) {
	script := `
metadata {
    name = "escape-test"
}
variables {
    msg = "hello \"world\""
}
workflow {
    steps [
        {
            cmd = "echo \"quoted\""
        }
    ]
}
config {
    store_logs = false
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if wf.Config.Variables["msg"] != `hello "world"` {
		t.Errorf("Expected escaped quotes resolved, got '%v'", wf.Config.Variables["msg"])
	}
}

func TestMigraineParser_MissingBlock(t *testing.T) {
	script := `
metadata {
    name = "no-workflow-block"
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Should parse with missing blocks, got error: %v", err)
	}
	if wf.Name != "no-workflow-block" {
		t.Errorf("Expected name parsed, got '%s'", wf.Name)
	}
}

func TestMigraineParser_UnknownBlock(t *testing.T) {
	script := `
unknown_block {
    foo = "bar"
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Fatal("Expected error for unknown block, got nil")
	}
}

func TestMigraineParser_EmptyWorkflow(t *testing.T) {
	script := `
metadata {
    name = "empty-workflow"
}
workflow {
    steps []
}
config {
    store_logs = false
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if len(wf.Steps) != 0 {
		t.Errorf("Expected 0 steps for empty steps list, got %d", len(wf.Steps))
	}
}

func TestMigraineParser_MultiplePreChecksAndSteps(t *testing.T) {
	script := `
metadata {
    name = "multi-check"
}
workflow {
    pre_checks [
        {
            cmd = "check 1"
            desc = "first check"
            on_fail = "action:handle_fail_1"
        },
        {
            cmd = "check 2"
            on_success = "action:handle_ok"
        }
    ]
    steps [
        {
            cmd = "step 1"
        },
        {
            cmd = "step 2"
            on_fail = "action:handle_fail_2"
        },
        {
            cmd = "step 3"
            on_success = "action:handle_ok"
            on_fail = "action:handle_fail_3"
        }
    ]
    actions {
        handle_fail_1 {
            cmd = "fail handler 1"
        }
        handle_fail_2 {
            cmd = "fail handler 2"
        }
        handle_fail_3 {
            cmd = "fail handler 3"
        }
        handle_ok {
            cmd = "success handler"
        }
    }
}
config {
    store_logs = true
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(wf.PreChecks) != 2 {
		t.Fatalf("Expected 2 pre-checks, got %d", len(wf.PreChecks))
	}
	if wf.PreChecks[0].OnFail != "action:handle_fail_1" {
		t.Errorf("Expected pre-check 0 on_fail, got '%s'", wf.PreChecks[0].OnFail)
	}
	if wf.PreChecks[1].OnSuccess != "action:handle_ok" {
		t.Errorf("Expected pre-check 1 on_success, got '%s'", wf.PreChecks[1].OnSuccess)
	}

	if len(wf.Steps) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(wf.Steps))
	}
	if wf.Steps[2].OnSuccess != "action:handle_ok" {
		t.Errorf("Expected step 2 on_success, got '%s'", wf.Steps[2].OnSuccess)
	}
	if wf.Steps[2].OnFail != "action:handle_fail_3" {
		t.Errorf("Expected step 2 on_fail, got '%s'", wf.Steps[2].OnFail)
	}

	if len(wf.Actions) != 4 {
		t.Fatalf("Expected 4 actions, got %d", len(wf.Actions))
	}
}

func TestMigraineParser_StringWithEquals(t *testing.T) {
	script := `
metadata {
    name = "test-equals"
}
variables {
    foo = "bar=baz"
}
workflow {
    steps [
        {
            cmd = "echo ok"
        }
    ]
}
config {
    store_logs = false
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if wf.Config.Variables["foo"] != "bar=baz" {
		t.Errorf("Expected 'bar=baz', got '%v'", wf.Config.Variables["foo"])
	}
}

func TestMigraineParser_ConversionRoundTrip(t *testing.T) {
	script := `
metadata {
    name = "roundtrip"
    desc = "Round trip test"
}
variables {
    path = "/tmp/test"
}
workflow {
    pre_checks [
        {
            cmd = "test -d /tmp"
            desc = "Check tmp dir"
            on_fail = "action:warn"
        }
    ]
    steps [
        {
            cmd = "echo hello"
        }
    ]
    actions {
        warn {
            cmd = "echo warning"
        }
    }
}
config {
    store_logs = true
    store_variables = false
    background = true
    global = false
}
`
	parser, err := NewMigraineParserFromReader(strings.NewReader(script))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	wf, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	yamlWf := ConvertInternalToYAML(wf, "test-id")

	if yamlWf.Name != "roundtrip" {
		t.Errorf("Roundtrip: name mismatch '%s'", yamlWf.Name)
	}
	if len(yamlWf.PreChecks) != 1 {
		t.Errorf("Roundtrip: expected 1 pre-check, got %d", len(yamlWf.PreChecks))
	}
	if len(yamlWf.Steps) != 1 {
		t.Errorf("Roundtrip: expected 1 step, got %d", len(yamlWf.Steps))
	}
	if len(yamlWf.Actions) != 1 {
		t.Errorf("Roundtrip: expected 1 action, got %d", len(yamlWf.Actions))
	}
	if yamlWf.Config.StoreLogs != true {
		t.Errorf("Roundtrip: expected StoreLogs=true")
	}
	if yamlWf.Config.StoreVariables != false {
		t.Errorf("Roundtrip: expected StoreVariables=false")
	}
	if yamlWf.Config.Background != true {
		t.Errorf("Roundtrip: expected Background=true")
	}
	if yamlWf.Config.Global != false {
		t.Errorf("Roundtrip: expected Global=false")
	}

	internalWf, err := ConvertYAMLToInternal(yamlWf)
	if err != nil {
		t.Fatalf("Failed to convert YAML back to internal: %v", err)
	}

	if internalWf.Name != wf.Name {
		t.Errorf("Roundtrip internal: name mismatch '%s' vs '%s'", internalWf.Name, wf.Name)
	}
	if len(internalWf.Steps) != len(wf.Steps) {
		t.Errorf("Roundtrip internal: steps count mismatch %d vs %d", len(internalWf.Steps), len(wf.Steps))
	}
}
