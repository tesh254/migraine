package workflow

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenBool
	TokenLBrace   // {
	TokenRBrace   // }
	TokenLBracket // [
	TokenRBracket // ]
	TokenAssign   // =
	TokenComma    // ,
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

type Lexer struct {
	reader *bufio.Reader
	line   int
	column int
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewReader(r),
		line:   1,
		column: 0,
	}
}

func (l *Lexer) read() rune {
	r, _, err := l.reader.ReadRune()
	if err != nil {
		return 0
	}
	l.column++
	if r == '\n' {
		l.line++
		l.column = 0
	}
	return r
}

func (l *Lexer) unread() {
	if err := l.reader.UnreadRune(); err == nil {
		l.column--
		// Note: handling line number on unread is tricky if we unread newline, 
		// but for this simple lexer we might not need complex unread logic.
		// A simple 1-char lookahead/backup is usually enough.
	}
}

func (l *Lexer) NextToken() (Token, error) {
	var r rune
	// Skip whitespace
	for {
		r = l.read()
		if r == 0 {
			return Token{Type: TokenEOF}, nil
		}
		if !unicode.IsSpace(r) {
			break
		}
	}

	// Skip comments
	if r == '#' {
		for {
			r = l.read()
			if r == '\n' || r == 0 {
				break
			}
		}
		return l.NextToken()
	}

	startLine := l.line
	startCol := l.column

	switch r {
	case '{':
		return Token{Type: TokenLBrace, Literal: "{", Line: startLine, Column: startCol}, nil
	case '}':
		return Token{Type: TokenRBrace, Literal: "}", Line: startLine, Column: startCol}, nil
	case '[':
		return Token{Type: TokenLBracket, Literal: "[", Line: startLine, Column: startCol}, nil
	case ']':
		return Token{Type: TokenRBracket, Literal: "]", Line: startLine, Column: startCol}, nil
	case '=':
		return Token{Type: TokenAssign, Literal: "=", Line: startLine, Column: startCol}, nil
	case ',':
		return Token{Type: TokenComma, Literal: ",", Line: startLine, Column: startCol}, nil
	case '"':
		return l.readString(startLine, startCol)
	case '`':
		return l.readBacktickString(startLine, startCol)
	}

	if unicode.IsDigit(r) {
		l.unread()
		return l.readNumber(startLine, startCol)
	}

	if unicode.IsLetter(r) || r == '_' {
		l.unread()
		return l.readIdentifier(startLine, startCol)
	}

	return Token{}, fmt.Errorf("unexpected character: %c at line %d:%d", r, l.line, l.column)
}

func (l *Lexer) readString(line, col int) (Token, error) {
	var buf bytes.Buffer
	for {
		r := l.read()
		if r == 0 {
			return Token{}, fmt.Errorf("unterminated string at line %d:%d", line, col)
		}
		if r == '"' {
			// Check for escape? The prompt says "avoid escape characters" by using backticks, 
			// but normal strings might still have them. For simplicity, let's assume standard escaping if needed,
			// or just simple string for now.
			break
		}
		if r == '\\' {
			next := l.read()
			if next == '"' || next == '\\' {
				buf.WriteRune(next)
			} else {
				buf.WriteRune(r)
				buf.WriteRune(next)
			}
		} else {
			buf.WriteRune(r)
		}
	}
	return Token{Type: TokenString, Literal: buf.String(), Line: line, Column: col}, nil
}

func (l *Lexer) readBacktickString(line, col int) (Token, error) {
	var buf bytes.Buffer
	for {
		r := l.read()
		if r == 0 {
			return Token{}, fmt.Errorf("unterminated raw string at line %d:%d", line, col)
		}
		if r == '`' {
			break
		}
		buf.WriteRune(r)
	}
	return Token{Type: TokenString, Literal: buf.String(), Line: line, Column: col}, nil
}

func (l *Lexer) readIdentifier(line, col int) (Token, error) {
	var buf bytes.Buffer
	for {
		r := l.read()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			buf.WriteRune(r)
		} else {
			if r != 0 {
				l.unread()
			}
			break
		}
	}
	lit := buf.String()
	if lit == "true" || lit == "false" {
		return Token{Type: TokenBool, Literal: lit, Line: line, Column: col}, nil
	}
	return Token{Type: TokenIdent, Literal: lit, Line: line, Column: col}, nil
}

func (l *Lexer) readNumber(line, col int) (Token, error) {
	var buf bytes.Buffer
	for {
		r := l.read()
		if unicode.IsDigit(r) || r == '.' {
			buf.WriteRune(r)
		} else {
			if r != 0 {
				l.unread()
			}
			break
		}
	}
	return Token{Type: TokenNumber, Literal: buf.String(), Line: line, Column: col}, nil
}
