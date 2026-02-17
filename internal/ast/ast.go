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
	Version    string            `json:"version,omitempty"`
	Dialect    string            `json:"dialect,omitempty"`
	Extensions map[string]string `json:"extensions,omitempty"`
	Task       *Task             `json:"task,omitempty"`
}

func (p *Program) Pos() lex.Loc {
	if p.Task != nil {
		return p.Task.Loc
	}
	return lex.Loc{}
}

// Task represents the top-level container for a program.
type Task struct {
	Loc    lex.Loc       `json:"-"`
	Name   string        `json:"name"`
	Inputs []*InputDecl  `json:"inputs,omitempty"`
	Body   []BodyItem    `json:"body"`
	Output string        `json:"output"`
}

func (t *Task) Pos() lex.Loc { return t.Loc }

// InputDecl represents an INPUT variable declaration.
type InputDecl struct {
	Loc  lex.Loc `json:"-"`
	Name string  `json:"name"`
	Type string  `json:"type"`
}

func (i *InputDecl) Pos() lex.Loc { return i.Loc }

// BodyItem is the interface for items allowed in a task or if body.
type BodyItem interface {
	Node
	bodyItemNode()
}

func (r *Requirement) bodyItemNode() {}
func (c *Cell) bodyItemNode()        {}

// IfStmt represents an IF/ELSE conditional block.
type IfStmt struct {
	Loc      lex.Loc    `json:"-"`
	Type     string     `json:"type"` // "if"
	Cond     Expr       `json:"cond"`
	ThenBody []BodyItem `json:"then_body"`
	ElseBody []BodyItem `json:"else_body,omitempty"`
}

func (s *IfStmt) Pos() lex.Loc    { return s.Loc }
func (s *IfStmt) bodyItemNode()   {}
func (s *IfStmt) stmtNode()       {} // If can also be used as a statement? EBNF says body_item.

// Requirement represents a REQUIRES capability statement.
type Requirement struct {
	Loc        lex.Loc `json:"-"`
	Capability string  `json:"capability"`
}

func (r *Requirement) Pos() lex.Loc { return r.Loc }

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
	BodyItem
}

// OpStmt represents a standard operation call: OP KW VAL... INTO ident.
type OpStmt struct {
	Loc      lex.Loc `json:"-"`
	Type     string  `json:"type"` // "op"
	OpName   string  `json:"op_name"`
	Args     []KwArg `json:"args"`
	Into     string  `json:"into"`
	IntoType string  `json:"into_type,omitempty"`
}

func (s *OpStmt) Pos() lex.Loc { return s.Loc }
func (s *OpStmt) stmtNode()   {}
func (s *OpStmt) bodyItemNode() {}

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
func (e *IdentExpr) exprNode()   {}

// StringExpr represents a string literal.
type StringExpr struct {
	Loc   lex.Loc `json:"-"`
	Kind  string  `json:"kind"` // "STRING"
	Value string  `json:"value"`
}

func (e *StringExpr) Pos() lex.Loc { return e.Loc }
func (e *StringExpr) exprNode()   {}

// IntExpr represents an integer literal.
type IntExpr struct {
	Loc   lex.Loc `json:"-"`
	Kind  string  `json:"kind"` // "INT"
	Value int     `json:"value"`
}

func (e *IntExpr) Pos() lex.Loc { return e.Loc }
func (e *IntExpr) exprNode()   {}

// BoolExpr represents a boolean literal.
type BoolExpr struct {
	Loc   lex.Loc `json:"-"`
	Kind  string  `json:"kind"` // "BOOL"
	Value bool    `json:"value"`
}

func (e *BoolExpr) Pos() lex.Loc { return e.Loc }
func (e *BoolExpr) exprNode()   {}

// NullExpr represents a null literal.
type NullExpr struct {
	Loc  lex.Loc `json:"-"`
	Kind string  `json:"kind"` // "NULL"
}

func (e *NullExpr) Pos() lex.Loc { return e.Loc }
func (e *NullExpr) exprNode()   {}

// SetFinalStmt represents the SET_FINAL command.
type SetFinalStmt struct {
	Loc    lex.Loc `json:"-"`
	Type   string  `json:"type"` // "set_final"
	Source Expr    `json:"source"`
}

func (s *SetFinalStmt) Pos() lex.Loc { return s.Loc }
func (s *SetFinalStmt) stmtNode()   {}
func (s *SetFinalStmt) bodyItemNode() {}

// AssertStmt represents the ASSERT command.
type AssertStmt struct {
	Loc     lex.Loc `json:"-"`
	Type    string  `json:"type"` // "assert"
	Cond    Expr    `json:"cond"`
	Message string  `json:"message"`
}

func (s *AssertStmt) Pos() lex.Loc { return s.Loc }
func (s *AssertStmt) stmtNode()   {}
func (s *AssertStmt) bodyItemNode() {}

// PrintStmt represents the PRINT command.
type PrintStmt struct {
	Loc    lex.Loc `json:"-"`
	Type   string  `json:"type"` // "print"
	Source Expr    `json:"source"`
}

func (s *PrintStmt) Pos() lex.Loc { return s.Loc }
func (s *PrintStmt) stmtNode()   {}
func (s *PrintStmt) bodyItemNode() {}

// ForEachStmt represents the FOR_EACH loop.
type ForEachStmt struct {
	Loc        lex.Loc `json:"-"`
	Type       string  `json:"type"` // "for_each"
	Iterator   string  `json:"iterator"`
	Collection string  `json:"collection"`
	Limit      int     `json:"limit"`
	Body       []Stmt  `json:"body"`
}

func (s *ForEachStmt) Pos() lex.Loc { return s.Loc }
func (s *ForEachStmt) stmtNode()   {}
func (s *ForEachStmt) bodyItemNode() {}

// Visitor interface for AST traversal.
type Visitor interface {
	Visit(Node) (w Visitor)
}

// Walk traverses an AST in depth-first order.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *Program:
		if n.Task != nil {
			Walk(v, n.Task)
		}
	case *Task:
		for _, in := range n.Inputs {
			Walk(v, in)
		}
		for _, item := range n.Body {
			Walk(v, item)
		}
	case *InputDecl:
		// Leaf
	case *IfStmt:
		Walk(v, n.Cond)
		for _, item := range n.ThenBody {
			Walk(v, item)
		}
		for _, item := range n.ElseBody {
			Walk(v, item)
		}
	case *Requirement:
		// Leaf
	case *Cell:
		for _, stmt := range n.Stmts {
			Walk(v, stmt)
		}
	case *OpStmt:
		for _, arg := range n.Args {
			Walk(v, arg.Value)
		}
	case *SetFinalStmt:
		Walk(v, n.Source)
	case *AssertStmt:
		Walk(v, n.Cond)
	case *PrintStmt:
		Walk(v, n.Source)
	case *ForEachStmt:
		for _, stmt := range n.Body {
			Walk(v, stmt)
		}
	case *IdentExpr, *StringExpr, *IntExpr, *BoolExpr, *NullExpr:
		// Leaf
	}

	v.Visit(nil)
}
