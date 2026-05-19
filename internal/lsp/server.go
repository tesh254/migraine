package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/tesh254/migraine/internal/workflow"
)

type Server struct {
	mu      sync.Mutex
	docs    map[string]string
	version int
}

func NewServer() *Server {
	return &Server{
		docs: make(map[string]string),
	}
}

func (s *Server) Handle(ctx context.Context, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "initialize":
		return s.handleInitialize(params)
	case "initialized":
		return nil, nil
	case "textDocument/didOpen":
		return nil, s.handleDidOpen(params)
	case "textDocument/didChange":
		return nil, s.handleDidChange(params)
	case "textDocument/didClose":
		return nil, s.handleDidClose(params)
	case "textDocument/diagnostic":
		return s.handleDiagnostic(params)
	case "textDocument/completion":
		return s.handleCompletion(params)
	case "textDocument/hover":
		return s.handleHover(params)
	case "textDocument/documentSymbol":
		return s.handleDocumentSymbol(params)
	case "textDocument/semanticTokens/full":
		return s.handleSemanticTokens(params)
	case "shutdown":
		return nil, nil
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

type InitializeParams struct {
	RootURI string `json:"rootUri"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync       TextDocumentSyncOptions   `json:"textDocumentSync"`
	CompletionProvider     CompletionOptions         `json:"completionProvider"`
	HoverProvider          bool                      `json:"hoverProvider"`
	DocumentSymbolProvider bool                      `json:"documentSymbolProvider"`
	SemanticTokensProvider SemanticTokensOptions      `json:"semanticTokensProvider"`
	DiagnosticProvider     DiagnosticOptions         `json:"diagnosticProvider"`
}

type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
}

type SemanticTokensOptions struct {
	Legend    SemanticTokensLegend `json:"legend"`
	Full      bool                 `json:"full"`
	Range     bool                 `json:"range"`
}

type DiagnosticOptions struct {
	Identifier           string `json:"identifier"`
	InterFileDependencies bool  `json:"interFileDependencies"`
	WorkspaceDiagnostics bool  `json:"workspaceDiagnostics"`
}

func (s *Server) handleInitialize(params json.RawMessage) (interface{}, error) {
	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: TextDocumentSyncOptions{
				OpenClose: true,
				Change:    1,
			},
			CompletionProvider: CompletionOptions{
				TriggerCharacters: []string{"=", " ", "{"},
			},
			HoverProvider:          true,
			DocumentSymbolProvider: true,
			SemanticTokensProvider: SemanticTokensOptions{
				Legend: SemanticTokensLegend{
					TokenTypes: []string{
						"namespace", "type", "class", "enum", "interface",
						"struct", "typeParameter", "parameter", "variable",
						"property", "enumMember", "event", "function", "member",
						"macro", "modifier", "comment", "string", "number",
						"regexp", "operator", "decorator",
					},
					TokenModifiers: []string{
						"declaration", "definition", "readonly", "static",
						"deprecated", "abstract", "async", "modification",
						"documentation", "defaultLibrary",
					},
				},
				Full:  true,
				Range: false,
			},
			DiagnosticProvider: DiagnosticOptions{
				Identifier:            "migraine-lsp",
				InterFileDependencies: false,
				WorkspaceDiagnostics: false,
			},
		},
	}, nil
}

type TextDocumentItem struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
	Text    string `json:"text"`
}

type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

func (s *Server) handleDidOpen(params json.RawMessage) error {
	var p DidOpenParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	s.mu.Lock()
	s.docs[p.TextDocument.URI] = p.TextDocument.Text
	s.mu.Unlock()
	return nil
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type DidChangeParams struct {
	TextDocument   TextDocumentIdentifier `json:"textDocument"`
	ContentChanges []ContentChangeEvent  `json:"contentChanges"`
}

type ContentChangeEvent struct {
	Text string `json:"text"`
}

func (s *Server) handleDidChange(params json.RawMessage) error {
	var p DidChangeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	s.mu.Lock()
	if len(p.ContentChanges) > 0 {
		s.docs[p.TextDocument.URI] = p.ContentChanges[0].Text
	}
	s.mu.Unlock()
	return nil
}

type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

func (s *Server) handleDidClose(params json.RawMessage) error {
	var p DidCloseParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.docs, p.TextDocument.URI)
	s.mu.Unlock()
	return nil
}

type DiagnosticParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DiagnosticResult struct {
	Items []DiagnosticItem `json:"items"`
}

type DiagnosticItem struct {
	Range    Range       `json:"range"`
	Severity int         `json:"severity"`
	Source   string      `json:"source"`
	Message  string      `json:"message"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func (s *Server) handleDiagnostic(params json.RawMessage) (interface{}, error) {
	var p DiagnosticParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	s.mu.Lock()
	text, ok := s.docs[p.TextDocument.URI]
	s.mu.Unlock()

	if !ok {
		return DiagnosticResult{}, nil
	}

	diagnostics := validateDocument(text)
	return DiagnosticResult{Items: diagnostics}, nil
}

func validateDocument(text string) []DiagnosticItem {
	parser, err := workflow.NewMigraineParserFromReader(strings.NewReader(text))
	if err != nil {
		return []DiagnosticItem{
			{
				Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 1}},
				Severity: 1,
				Source:   "migraine-lsp",
				Message:  fmt.Sprintf("Failed to create parser: %v", err),
			},
		}
	}

	_, err = parser.Parse()
	if err != nil {
		line := 0
		char := 0
		diag := DiagnosticItem{
			Range: Range{
				Start: Position{Line: line, Character: char},
				End:   Position{Line: line, Character: char + 1},
			},
			Severity: 1,
			Source:   "migraine-lsp",
			Message:  err.Error(),
		}
		return []DiagnosticItem{diag}
	}
	return nil
}

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position    Position               `json:"position"`
}

type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"`
	Documentation string `json:"documentation,omitempty"`
}

type CompletionResult struct {
	Items []CompletionItem `json:"items"`
}

func (s *Server) handleCompletion(params json.RawMessage) (interface{}, error) {
	var p CompletionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	items := []CompletionItem{}

	blockKeywords := []CompletionItem{
		{Label: "metadata", Kind: 6, Documentation: "Workflow metadata block (name, description)"},
		{Label: "variables", Kind: 6, Documentation: "Variable definitions block"},
		{Label: "workflow", Kind: 6, Documentation: "Workflow definition block (steps, pre_checks, actions)"},
		{Label: "config", Kind: 6, Documentation: "Configuration block (store_variables, store_logs, background, global)"},
	}

	workflowKeywords := []CompletionItem{
		{Label: "steps", Kind: 6, Documentation: "Ordered list of execution steps"},
		{Label: "pre_checks", Kind: 6, Documentation: "Pre-flight checks before running steps"},
		{Label: "actions", Kind: 6, Documentation: "Named reusable actions that can be triggered by hooks"},
	}

	atomKeywords := []CompletionItem{
		{Label: "cmd", Kind: 6, Documentation: "Command to execute"},
		{Label: "desc", Kind: 6, Documentation: "Human-readable description"},
		{Label: "on_fail", Kind: 6, Documentation: "Action or command to run on failure (e.g. 'action:name' or 'run:cmd')"},
		{Label: "on_success", Kind: 6, Documentation: "Action or command to run on success (e.g. 'action:name' or 'run:cmd')"},
	}

	configKeywords := []CompletionItem{
		{Label: "store_variables", Kind: 6, Documentation: "Persist resolved variables between runs"},
		{Label: "store_logs", Kind: 6, Documentation: "Store execution logs"},
		{Label: "background", Kind: 6, Documentation: "Run workflow in the background"},
		{Label: "global", Kind: 6, Documentation: "Make workflow available globally"},
	}

	metadataKeywords := []CompletionItem{
		{Label: "name", Kind: 6, Documentation: "Workflow name"},
		{Label: "desc", Kind: 6, Documentation: "Workflow description"},
	}

	valueHints := []CompletionItem{
		{Label: "true", Kind: 12, Documentation: "Boolean true"},
		{Label: "false", Kind: 12, Documentation: "Boolean false"},
		{Label: "args:", Kind: 15, Documentation: "Resolve from CLI argument (e.g. args:APP_NAME)"},
		{Label: "env:", Kind: 15, Documentation: "Resolve from environment variable (e.g. env:HOME)"},
		{Label: "vault:", Kind: 15, Documentation: "Resolve from migraine vault (e.g. vault:SECRET_KEY)"},
		{Label: "action:", Kind: 15, Documentation: "Reference a named action in on_fail/on_success (e.g. action:notify)"},
		{Label: "run:", Kind: 15, Documentation: "Run a command in on_fail/on_success (e.g. 'run:echo done')"},
	}

	items = append(items, blockKeywords...)
	items = append(items, workflowKeywords...)
	items = append(items, atomKeywords...)
	items = append(items, configKeywords...)
	items = append(items, metadataKeywords...)
	items = append(items, valueHints...)

	return CompletionResult{Items: items}, nil
}

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position    Position               `json:"position"`
}

type HoverResult struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

var hoverDocs = map[string]string{
	"metadata":      "## metadata block\nDefines workflow metadata: `name` and `desc` (description).",
	"variables":     "## variables block\nDefine variables resolved at runtime.\n\nPrefixes:\n- `args:VAR` — from CLI flags\n- `env:VAR` — from environment\n- `vault:VAR` — from migraine vault",
	"workflow":      "## workflow block\nContains `pre_checks`, `steps`, and `actions`.",
	"config":        "## config block\nConfiguration options:\n- `store_variables` (bool)\n- `store_logs` (bool)\n- `background` (bool)\n- `global` (bool)",
	"pre_checks":    "## pre_checks\nPre-flight checks that run before steps. Each check is an atom with `cmd`, optional `desc`, `on_fail`, `on_success`.",
	"steps":         "## steps\nOrdered execution steps. Each step is an atom with `cmd`, optional `desc`, `on_fail`, `on_success`.",
	"actions":       "## actions\nNamed reusable actions triggered by `on_fail` or `on_success` hooks.\n\nReference with `action:name` in hook fields.",
	"cmd":           "## cmd\nThe shell command to execute. Supports template variables: `{{var_name}}`.",
	"desc":          "## desc\nHuman-readable description displayed during execution.",
	"on_fail":       "## on_fail\nHook executed when the step/check fails. Use `action:name` to reference an action, or `run:command` for inline.",
	"on_success":    "## on_success\nHook executed when the step/check succeeds. Use `action:name` to reference an action, or `run:command` for inline.",
	"store_variables": "`store_variables` (bool): Persist resolved variables between runs.",
	"store_logs":      "`store_logs` (bool): Store execution logs for later review.",
	"background":      "`background` (bool): Run the workflow in the background.",
	"global":           "`global` (bool): Make the workflow available across all projects.",
}

func (s *Server) handleHover(params json.RawMessage) (interface{}, error) {
	var p HoverParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	s.mu.Lock()
	text, ok := s.docs[p.TextDocument.URI]
	s.mu.Unlock()

	if !ok {
		return nil, nil
	}

	word := getWordAtPosition(text, p.Position.Line, p.Position.Character)
	if word == "" {
		return nil, nil
	}

	if docs, ok := hoverDocs[word]; ok {
		return HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: docs,
			},
		}, nil
	}

	return nil, nil
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

func (s *Server) handleDocumentSymbol(params json.RawMessage) (interface{}, error) {
	var p DocumentSymbolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	s.mu.Lock()
	text, ok := s.docs[p.TextDocument.URI]
	s.mu.Unlock()

	if !ok {
		return []DocumentSymbol{}, nil
	}

	return extractDocumentSymbols(text), nil
}

type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type SemanticTokensResult struct {
	Data []int `json:"data"`
}

func (s *Server) handleSemanticTokens(params json.RawMessage) (interface{}, error) {
	var p SemanticTokensParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	s.mu.Lock()
	text, ok := s.docs[p.TextDocument.URI]
	s.mu.Unlock()

	if !ok {
		return SemanticTokensResult{}, nil
	}

	tokens := tokenizeDocument(text)
	return SemanticTokensResult{Data: tokens}, nil
}

func getWordAtPosition(text string, line, character int) string {
	lines := splitLines(text)
	if line >= len(lines) {
		return ""
	}
	l := lines[line]
	if character >= len(l) {
		return ""
	}

	start := character
	for start > 0 && isIdentChar(rune(l[start-1])) {
		start--
	}

	end := character
	for end < len(l) && isIdentChar(rune(l[end])) {
		end++
	}

	if start >= end {
		return ""
	}
	return l[start:end]
}

func isIdentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-'
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func tokenizeDocument(text string) []int {
	lexer := workflow.NewLexer(strings.NewReader(text))
	var data []int
	prevLine := 0
	prevChar := 0

	for {
		tok, err := lexer.NextToken()
		if err != nil {
			break
		}
		if tok.Type == workflow.TokenEOF {
			break
		}

		deltaLine := tok.Line - 1 - prevLine
		var deltaStart int
		if deltaLine == 0 {
			deltaStart = tok.Column - 1 - prevChar
		} else {
			deltaStart = tok.Column - 1
		}

		length := len(tok.Literal)
		if length == 0 {
			length = 1
		}

		var tokenType int
		switch tok.Type {
		case workflow.TokenIdent:
			tokenType = identTokenType(tok.Literal)
		case workflow.TokenString:
			tokenType = 13 // string
		case workflow.TokenNumber:
			tokenType = 15 // number
		case workflow.TokenBool:
			tokenType = 16 // regexp (using for boolean literals)
		case workflow.TokenLBrace, workflow.TokenRBrace:
			tokenType = 18 // operator (for braces/punctuation)
		case workflow.TokenLBracket, workflow.TokenRBracket:
			tokenType = 18
		case workflow.TokenAssign:
			tokenType = 18
		case workflow.TokenComma:
			tokenType = 18
		default:
			continue
		}

		data = append(data, deltaLine, deltaStart, length, tokenType, 0)

		prevLine = tok.Line - 1
		prevChar = tok.Column - 1
	}

	return data
}

var blockNames = map[string]bool{
	"metadata": true, "variables": true, "workflow": true, "config": true,
}

var sectionNames = map[string]bool{
	"pre_checks": true, "steps": true, "actions": true,
}

var propertyNames = map[string]bool{
	"cmd": true, "desc": true, "description": true,
	"on_fail": true, "on_success": true,
	"store_variables": true, "store_logs": true,
	"background": true, "global": true,
	"name": true,
}

func identTokenType(literal string) int {
	if blockNames[literal] || sectionNames[literal] {
		return 3 // class - for block/section names
	}
	if propertyNames[literal] {
		return 8 // property
	}
	return 0 // namespace
}

func extractDocumentSymbols(text string) []DocumentSymbol {
	lexer := workflow.NewLexer(strings.NewReader(text))
	var symbols []DocumentSymbol
	var currentBlock string
	blockStartLine := 0

	for {
		tok, err := lexer.NextToken()
		if err != nil {
			break
		}
		if tok.Type == workflow.TokenEOF {
			if currentBlock != "" {
				symbols = append(symbols, DocumentSymbol{
					Name:           currentBlock,
					Kind:           symbolKind(currentBlock),
					Range:          Range{Start: Position{Line: blockStartLine, Character: 0}, End: Position{Line: tok.Line - 1, Character: 0}},
					SelectionRange: Range{Start: Position{Line: blockStartLine, Character: 0}, End: Position{Line: blockStartLine, Character: len(currentBlock)}},
				})
			}
			break
		}

		if tok.Type == workflow.TokenIdent {
			if blockNames[tok.Literal] || sectionNames[tok.Literal] {
				if currentBlock != "" {
					symbols = append(symbols, DocumentSymbol{
						Name:           currentBlock,
						Kind:           symbolKind(currentBlock),
						Range:          Range{Start: Position{Line: blockStartLine, Character: 0}, End: Position{Line: tok.Line - 1, Character: 0}},
						SelectionRange: Range{Start: Position{Line: blockStartLine, Character: 0}, End: Position{Line: blockStartLine, Character: len(currentBlock)}},
					})
				}
				currentBlock = tok.Literal
				blockStartLine = tok.Line - 1
			}
		}
	}

	return symbols
}

func symbolKind(name string) int {
	switch name {
	case "metadata", "config", "variables":
		return 7 // Module
	case "workflow":
		return 6 // Class
	case "steps", "pre_checks", "actions":
		return 12 // Function
	default:
		return 12
	}
}

func RunStdio() {
	server := NewServer()
	RunStdioServer(server)
}