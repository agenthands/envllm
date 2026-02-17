package parse

import (
	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/lex"
)

type strictParser struct {
	baseParser
}

func NewStrictParser(l *lex.Lexer) Parser {
	p := &strictParser{
		baseParser: baseParser{
			l:    l,
			mode: ModeStrict,
		},
	}
	p.init()
	return p
}

func (p *strictParser) Parse() (*ast.Program, error) {
	return p.baseParser.Parse()
}
