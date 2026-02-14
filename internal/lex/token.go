package lex

import "fmt"

// Type defines the kind of token.
type Type int

const (
	TypeEOF Type = iota
	TypeError

	// Keywords
	TypeRLMDSL
	TypeCELL
	TypeINTO
	TypeSET_FINAL
	TypeASSERT
	TypePRINT

	// Literals
	TypeIdent
	TypeString
	TypeInt
	TypeBool
	TypeJSON

	// Symbols
	TypeColon
	TypeNewline
)

// Loc represents a location in the source code.
type Loc struct {
	File string
	Line int
	Col  int
}

func (l Loc) String() string {
	return fmt.Sprintf("%s:%d:%d", l.File, l.Line, l.Col)
}

// Token represents a single lexical unit.
type Token struct {
	Type  Type
	Value string
	Loc   Loc
}
