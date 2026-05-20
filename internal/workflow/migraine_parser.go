package workflow

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

type MigraineParser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
}

func LoadMigraineWorkflow(path string) (*YAMLWorkflow, error) {
	p, err := NewMigraineParser(path)
	if err != nil {
		return nil, err
	}
	wf, err := p.Parse()
	if err != nil {
		return nil, err
	}
	yamlWf := ConvertInternalToYAML(wf, "")
	yamlWf.Path = path
	return yamlWf, nil
}

func NewMigraineParser(path string) (*MigraineParser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	
	return NewMigraineParserFromReader(file)
}

func NewMigraineParserFromReader(r io.Reader) (*MigraineParser, error) {
	l := NewLexer(r)
	p := &MigraineParser{lexer: l}
	
	// Read two tokens to prime the parser
	p.nextToken()
	p.nextToken()
	
	return p, nil
}

func (p *MigraineParser) nextToken() {
	p.curToken = p.peekToken
	tok, err := p.lexer.NextToken()
	if err != nil {
		// If error is EOF, it's fine. If other error, we might want to handle it.
		// For now, we'll just let the parser fail on unexpected tokens.
		// Ideally we should store the error.
	}
	p.peekToken = tok
}

func (p *MigraineParser) Parse() (*Workflow, error) {
	wf := &Workflow{
		Config: Config{
			Variables: make(map[string]interface{}),
		},
		Actions: make(map[string]Atom),
	}

	for p.curToken.Type != TokenEOF {
		if p.curToken.Type != TokenIdent {
			return nil, fmt.Errorf("expected identifier at start of block, got %v", p.curToken)
		}

		blockName := p.curToken.Literal
		p.nextToken() // consume block name

		if p.curToken.Type != TokenLBrace {
			return nil, fmt.Errorf("expected { after %s, got %v", blockName, p.curToken)
		}
		p.nextToken() // consume {

		switch blockName {
		case "metadata":
			if err := p.parseMetadata(wf); err != nil {
				return nil, err
			}
		case "variables":
			if err := p.parseVariables(wf); err != nil {
				return nil, err
			}
		case "workflow":
			if err := p.parseWorkflow(wf); err != nil {
				return nil, err
			}
		case "config":
			if err := p.parseConfig(wf); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown block: %s", blockName)
		}

		// Optional comma after block? Not in example.
		// But we expect } to be consumed by parse* methods.
		// Wait, my previous logic was: parse* consumes content inside { ... }.
		// So here we expect to be after }.
		// Let's check parse* implementations.
		
		// Actually, standard pattern is: parseBlock consumes { and }.
		// But here I consumed { before switch.
		// So parse* should consume content and }.
	}

	return wf, nil
}

func (p *MigraineParser) parseMetadata(wf *Workflow) error {
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		key, val, err := p.parseKeyValue()
		if err != nil {
			return err
		}
		
		switch key {
		case "name":
			if s, ok := val.(string); ok {
				wf.Name = s
			}
		case "desc", "description":
			if s, ok := val.(string); ok {
				wf.Description = &s
			}
		}
	}
	if p.curToken.Type != TokenRBrace {
		return fmt.Errorf("expected } after metadata")
	}
	p.nextToken() // consume }
	return nil
}

func (p *MigraineParser) parseVariables(wf *Workflow) error {
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		key, val, err := p.parseKeyValue()
		if err != nil {
			return err
		}
		wf.Config.Variables[key] = val
	}
	if p.curToken.Type != TokenRBrace {
		return fmt.Errorf("expected } after variables")
	}
	p.nextToken() // consume }
	return nil
}

func (p *MigraineParser) parseWorkflow(wf *Workflow) error {
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		if p.curToken.Type != TokenIdent {
			return fmt.Errorf("expected identifier in workflow block, got %v", p.curToken)
		}
		
		section := p.curToken.Literal
		p.nextToken()

		if section == "actions" {
			if p.curToken.Type != TokenLBrace {
				return fmt.Errorf("expected { after actions, got %v", p.curToken)
			}
			p.nextToken() // consume {
			
			if err := p.parseActions(wf); err != nil {
				return err
			}
			// parseActions consumes }
			continue
		}

		// pre_checks or steps
		if p.curToken.Type != TokenLBracket {
			// Maybe it's not a list?
			// The example shows: pre_checks [ ... ]
			return fmt.Errorf("expected [ after %s, got %v", section, p.curToken)
		}
		p.nextToken() // consume [

		atoms, err := p.parseAtomList()
		if err != nil {
			return err
		}

		switch section {
		case "pre_checks":
			wf.PreChecks = atoms
		case "steps":
			wf.Steps = atoms
		default:
			return fmt.Errorf("unknown workflow section: %s", section)
		}
		
		// parseAtomList consumes ]
	}
	if p.curToken.Type != TokenRBrace {
		return fmt.Errorf("expected } after workflow")
	}
	p.nextToken() // consume }
	return nil
}

func (p *MigraineParser) parseActions(wf *Workflow) error {
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenComma {
			p.nextToken()
			continue
		}

		if p.curToken.Type != TokenIdent {
			return fmt.Errorf("expected action name, got %v", p.curToken)
		}
		actionName := p.curToken.Literal
		p.nextToken()

		if p.curToken.Type != TokenLBrace {
			return fmt.Errorf("expected { after action name, got %v", p.curToken)
		}
		p.nextToken() // consume {

		atom, err := p.parseAtom()
		if err != nil {
			return err
		}
		wf.Actions[actionName] = atom

		// parseAtom consumes }? No, parseAtom parses content inside { ... }.
		// We need to consume } here.
		if p.curToken.Type != TokenRBrace {
			return fmt.Errorf("expected } after action block %s, got %v", actionName, p.curToken)
		}
		p.nextToken() // consume }
	}
	if p.curToken.Type != TokenRBrace {
		return fmt.Errorf("expected } after actions")
	}
	p.nextToken() // consume }
	return nil
}

func (p *MigraineParser) parseConfig(wf *Workflow) error {
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		key, val, err := p.parseKeyValue()
		if err != nil {
			return err
		}
		
		switch key {
		case "store_variables":
			if b, ok := val.(bool); ok {
				wf.Config.StoreVariables = b
			}
		case "store_logs":
			if b, ok := val.(bool); ok {
				wf.Config.StoreLogs = b
			}
		case "background":
			if b, ok := val.(bool); ok {
				wf.Config.Background = b
			}
		case "global":
			if b, ok := val.(bool); ok {
				wf.Config.Global = b
			}
		}
	}
	if p.curToken.Type != TokenRBrace {
		return fmt.Errorf("expected } after config")
	}
	p.nextToken() // consume }
	return nil
}

func (p *MigraineParser) parseKeyValue() (string, interface{}, error) {
	if p.curToken.Type != TokenIdent {
		return "", nil, fmt.Errorf("expected key, got %v", p.curToken)
	}
	key := p.curToken.Literal
	p.nextToken()

	if p.curToken.Type != TokenAssign {
		return "", nil, fmt.Errorf("expected = after key %s, got %v", key, p.curToken)
	}
	p.nextToken()

	var val interface{}
	switch p.curToken.Type {
	case TokenString:
		val = p.curToken.Literal
	case TokenBool:
		b, _ := strconv.ParseBool(p.curToken.Literal)
		val = b
	case TokenNumber:
		// Not used in example but good to have
		f, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		val = f
	default:
		return "", nil, fmt.Errorf("expected value for key %s, got %v", key, p.curToken)
	}
	p.nextToken()

	// Optional comma
	if p.curToken.Type == TokenComma {
		p.nextToken()
	}

	return key, val, nil
}

func (p *MigraineParser) parseAtomList() ([]Atom, error) {
	var atoms []Atom
	for p.curToken.Type != TokenRBracket && p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenComma {
			p.nextToken()
			continue
		}
		
		if p.curToken.Type != TokenLBrace {
			return nil, fmt.Errorf("expected { for atom, got %v", p.curToken)
		}
		p.nextToken() // consume {

		atom, err := p.parseAtom()
		if err != nil {
			return nil, err
		}
		atoms = append(atoms, atom)

		if p.curToken.Type != TokenRBrace {
			return nil, fmt.Errorf("expected } after atom, got %v", p.curToken)
		}
		p.nextToken() // consume }
		
		if p.curToken.Type == TokenComma {
			p.nextToken()
		}
	}
	
	if p.curToken.Type != TokenRBracket {
		return nil, fmt.Errorf("expected ] at end of list")
	}
	p.nextToken() // consume ]
	
	return atoms, nil
}

func (p *MigraineParser) parseAtom() (Atom, error) {
	var atom Atom
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		key, val, err := p.parseKeyValue()
		if err != nil {
			return atom, err
		}
		
		switch key {
		case "cmd":
			if s, ok := val.(string); ok {
				atom.Command = s
			}
		case "desc":
			if s, ok := val.(string); ok {
				atom.Description = &s
			}
		case "on_fail":
			if s, ok := val.(string); ok {
				atom.OnFail = s
			}
		case "on_success":
			if s, ok := val.(string); ok {
				atom.OnSuccess = s
			}
		}
	}
	return atom, nil
}
