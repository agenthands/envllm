package lex

import (
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	file    string
	input   string
	pos     int
	readPos int
	ch      rune
	line    int
	col     int
}

func NewLexer(file, input string) *Lexer {
	l := &Lexer{
		file:  file,
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPos:])
		l.ch = r
		l.pos = l.readPos
		l.readPos += size
	}
	
	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Loc = Loc{File: l.file, Line: l.line, Col: l.col}

	switch l.ch {
	case '"':
		tok.Type = TypeString
		tok.Value = l.readString()
		return tok
	case ':':
		tok.Type = TypeColon
		tok.Value = ":"
	case '\n':
		tok.Type = TypeNewline
		tok.Value = "\n"
	case 0:
		tok.Type = TypeEOF
		tok.Value = ""
	default:
		if isLetter(l.ch) || unicode.IsDigit(l.ch) {
			tok.Value = l.readIdentifier()
			tok.Type = lookupIdent(tok.Value)
			return tok
		} else {
			tok.Type = TypeError
			tok.Value = string(l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readString() string {
	l.readChar() // skip "
	pos := l.pos
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	s := l.input[pos:l.pos]
	l.readChar() // skip "
	return s
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '.' || l.ch == '_' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func lookupIdent(ident string) Type {
	switch ident {
	case "RLMDSL":
		return TypeRLMDSL
	case "CELL":
		return TypeCELL
	case "INTO":
		return TypeINTO
	case "SET_FINAL":
		return TypeSET_FINAL
	case "ASSERT":
		return TypeASSERT
	case "PRINT":
		return TypePRINT
	case "true", "false":
		return TypeBool
	default:
		allDigits := true
		for _, r := range ident {
			if !unicode.IsDigit(r) {
				allDigits = false
				break
			}
		}
		if allDigits && len(ident) > 0 {
			return TypeInt
		}
		return TypeIdent
	}
}
