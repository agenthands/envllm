package parse

import (
	"fmt"
	"strconv"

	"github.com/agenthands/rlm-go/internal/ast"
	"github.com/agenthands/rlm-go/internal/lex"
)

type Parser struct {
	l         *lex.Lexer
	curToken  lex.Token
	peekToken lex.Token
}

func NewParser(l *lex.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}

	if p.curToken.Type == lex.TypeRLMDSL {
		p.nextToken()
		if p.curToken.Type != lex.TypeIdent {
			return nil, fmt.Errorf("%s: expected version after RLMDSL", p.curToken.Loc)
		}
		prog.Version = p.curToken.Value
		p.nextToken()
		if err := p.expectNewline(); err != nil {
			return nil, err
		}
	}

	for p.curToken.Type != lex.TypeEOF {
		if p.curToken.Type == lex.TypeNewline {
			p.nextToken()
			continue
		}
		if p.curToken.Type == lex.TypeCELL {
			cell, err := p.parseCell()
			if err != nil {
				return nil, err
			}
			prog.Cells = append(prog.Cells, cell)
		} else {
			return nil, fmt.Errorf("%s: expected CELL, got %v (%q)", p.curToken.Loc, p.curToken.Type, p.curToken.Value)
		}
	}

	return prog, nil
}

func (p *Parser) parseCell() (*ast.Cell, error) {
	cell := &ast.Cell{Loc: p.curToken.Loc}
	p.nextToken()

	if p.curToken.Type != lex.TypeIdent {
		return nil, fmt.Errorf("%s: expected cell name", p.curToken.Loc)
	}
	cell.Name = p.curToken.Value
	p.nextToken()

	if p.curToken.Type != lex.TypeColon {
		return nil, fmt.Errorf("%s: expected ':' after cell name", p.curToken.Loc)
	}
	p.nextToken()

	if err := p.expectNewline(); err != nil {
		return nil, err
	}

	for p.curToken.Type != lex.TypeCELL && p.curToken.Type != lex.TypeEOF {
		if p.curToken.Type == lex.TypeNewline {
			p.nextToken()
			continue
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		cell.Stmts = append(cell.Stmts, stmt)
	}

	return cell, nil
}

func (p *Parser) parseStatement() (ast.Stmt, error) {
	switch p.curToken.Type {
	case lex.TypeSET_FINAL:
		return p.parseSetFinal()
	case lex.TypeASSERT:
		return p.parseAssert()
	case lex.TypePRINT:
		return p.parsePrint()
	case lex.TypeIdent:
		return p.parseOpStatement()
	default:
		return nil, fmt.Errorf("%s: unexpected token in statement: %v (%q)", p.curToken.Loc, p.curToken.Type, p.curToken.Value)
	}
}

func (p *Parser) parseOpStatement() (*ast.OpStmt, error) {
	stmt := &ast.OpStmt{Loc: p.curToken.Loc, OpName: p.curToken.Value}
	p.nextToken()

	for p.curToken.Type == lex.TypeIdent && p.peekToken.Type != lex.TypeEOF && p.curToken.Value != "INTO" {
		kw := p.curToken.Value
		p.nextToken()
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Args = append(stmt.Args, ast.KwArg{Keyword: kw, Value: val})
	}

	if p.curToken.Type != lex.TypeINTO {
		return nil, fmt.Errorf("%s: expected INTO, got %v (%q)", p.curToken.Loc, p.curToken.Type, p.curToken.Value)
	}
	p.nextToken()

	if p.curToken.Type != lex.TypeIdent {
		return nil, fmt.Errorf("%s: expected identifier after INTO", p.curToken.Loc)
	}
	stmt.Into = p.curToken.Value
	p.nextToken()

	if err := p.expectNewline(); err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	switch p.curToken.Type {
	case lex.TypeIdent:
		e := &ast.IdentExpr{Loc: p.curToken.Loc, Name: p.curToken.Value}
		p.nextToken()
		return e, nil
	case lex.TypeString:
		e := &ast.StringExpr{Loc: p.curToken.Loc, Value: p.curToken.Value}
		p.nextToken()
		return e, nil
	case lex.TypeInt:
		val, _ := strconv.Atoi(p.curToken.Value)
		e := &ast.IntExpr{Loc: p.curToken.Loc, Value: val}
		p.nextToken()
		return e, nil
	case lex.TypeBool:
		val := p.curToken.Value == "true"
		e := &ast.BoolExpr{Loc: p.curToken.Loc, Value: val}
		p.nextToken()
		return e, nil
	default:
		return nil, fmt.Errorf("%s: expected expression, got %v", p.curToken.Loc, p.curToken.Type)
	}
}

func (p *Parser) parseSetFinal() (*ast.SetFinalStmt, error) {
	stmt := &ast.SetFinalStmt{Loc: p.curToken.Loc}
	p.nextToken()
	if p.curToken.Value != "SOURCE" {
		return nil, fmt.Errorf("%s: expected SOURCE after SET_FINAL", p.curToken.Loc)
	}
	p.nextToken()
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	stmt.Source = val
	if err := p.expectNewline(); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) parseAssert() (*ast.AssertStmt, error) {
	stmt := &ast.AssertStmt{Loc: p.curToken.Loc}
	p.nextToken()
	if p.curToken.Value != "COND" {
		return nil, fmt.Errorf("%s: expected COND after ASSERT", p.curToken.Loc)
	}
	p.nextToken()
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	stmt.Cond = cond
	if p.curToken.Value != "MESSAGE" {
		return nil, fmt.Errorf("%s: expected MESSAGE after ASSERT COND", p.curToken.Loc)
	}
	p.nextToken()
	if p.curToken.Type != lex.TypeString {
		return nil, fmt.Errorf("%s: expected string message for ASSERT", p.curToken.Loc)
	}
	stmt.Message = p.curToken.Value
	p.nextToken()
	if err := p.expectNewline(); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) parsePrint() (*ast.PrintStmt, error) {
	stmt := &ast.PrintStmt{Loc: p.curToken.Loc}
	p.nextToken()
	if p.curToken.Value != "SOURCE" {
		return nil, fmt.Errorf("%s: expected SOURCE after PRINT", p.curToken.Loc)
	}
	p.nextToken()
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	stmt.Source = val
	if err := p.expectNewline(); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) expectNewline() error {
	if p.curToken.Type != lex.TypeNewline && p.curToken.Type != lex.TypeEOF {
		return fmt.Errorf("%s: expected newline, got %v (%q)", p.curToken.Loc, p.curToken.Type, p.curToken.Value)
	}
	if p.curToken.Type == lex.TypeNewline {
		p.nextToken()
	}
	return nil
}
