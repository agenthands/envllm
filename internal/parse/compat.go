package parse

import (
	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/lex"
)

type compatParser struct {
	baseParser
}

func NewCompatParser(l *lex.Lexer) Parser {
	p := &compatParser{
		baseParser: baseParser{
			l:    l,
			mode: ModeCompat,
		},
	}
	p.init()
	return p
}

func (p *compatParser) Parse() (*ast.Program, error) {
	return p.baseParser.Parse()
}
