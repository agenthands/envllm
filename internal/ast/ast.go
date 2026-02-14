package ast

import (
	"github.com/agenthands/envllm/internal/lex"
)

// Node is the interface for all AST nodes.
type Node interface {
	Pos() lex.Loc
}

// Program represents the root of the RLMDSL script.
type Program struct {
	Version string  `json:"version,omitempty"`
	Cells   []*Cell `json:"cells,omitempty"`
}

func (p *Program) Pos() lex.Loc {
	if len(p.Cells) > 0 {
		return p.Cells[0].Pos()
	}
	return lex.Loc{}
}

// Cell represents a block of execution.
type Cell struct {
	Loc   lex.Loc `json:"-"`
	Name  string  `json:"name"`
	Stmts []Stmt  `json:"stmts"`
}

func (c *Cell) Pos() lex.Loc { return c.Loc }

// Stmt represents a single statement in a cell.
type Stmt interface {
	Node
	stmtNode()
}

// OpStmt represents a standard operation call: OP KW VAL... INTO ident.
type OpStmt struct {
	Loc    lex.Loc `json:"-"`
	Type   string  `json:"type"` // "op"
	OpName string  `json:"op_name"`
	Args   []KwArg `json:"args"`
	Into   string  `json:"into"`
}

func (s *OpStmt) Pos() lex.Loc { return s.Loc }
func (s *OpStmt) stmtNode()    {}

// KwArg represents a keyword-argument pair.
type KwArg struct {
	Keyword string `json:"kw"`
	Value   Expr   `json:"value"`
}

// Expr represents an expression (literal or identifier).
type Expr interface {
	Node
	exprNode()
}

// IdentExpr represents an identifier.
type IdentExpr struct {
	Loc  lex.Loc `json:"-"`
	Kind string  `json:"kind"` // "IDENT"
	Name string  `json:"name"`
}

func (e *IdentExpr) Pos() lex.Loc { return e.Loc }
func (e *IdentExpr) exprNode()    {}

// StringExpr represents a string literal.
type StringExpr struct {
	Loc   lex.Loc `json:"-"`
	Kind  string  `json:"kind"` // "STRING"
	Value string  `json:"value"`
}

func (e *StringExpr) Pos() lex.Loc { return e.Loc }
func (e *StringExpr) exprNode()    {}

// IntExpr represents an integer literal.
type IntExpr struct {
	Loc   lex.Loc `json:"-"`
	Kind  string  `json:"kind"` // "INT"
	Value int     `json:"value"`
}

func (e *IntExpr) Pos() lex.Loc { return e.Loc }
func (e *IntExpr) exprNode()    {}

// BoolExpr represents a boolean literal.
type BoolExpr struct {
	Loc   lex.Loc `json:"-"`
	Kind  string  `json:"kind"` // "BOOL"
	Value bool    `json:"value"`
}

func (e *BoolExpr) Pos() lex.Loc { return e.Loc }
func (e *BoolExpr) exprNode()    {}

// SetFinalStmt represents the SET_FINAL command.
type SetFinalStmt struct {
	Loc    lex.Loc `json:"-"`
	Type   string  `json:"type"` // "set_final"
	Source Expr    `json:"source"`
}

func (s *SetFinalStmt) Pos() lex.Loc { return s.Loc }
func (s *SetFinalStmt) stmtNode()    {}

// AssertStmt represents the ASSERT command.
type AssertStmt struct {
	Loc     lex.Loc `json:"-"`
	Type    string  `json:"type"` // "assert"
	Cond    Expr    `json:"cond"`
	Message string  `json:"message"`
}

func (s *AssertStmt) Pos() lex.Loc { return s.Loc }
func (s *AssertStmt) stmtNode()    {}

// PrintStmt represents the PRINT command.
type PrintStmt struct {
	Loc    lex.Loc `json:"-"`
	Type   string  `json:"type"` // "print"
	Source Expr    `json:"source"`
}

func (s *PrintStmt) Pos() lex.Loc { return s.Loc }
func (s *PrintStmt) stmtNode()    {}
