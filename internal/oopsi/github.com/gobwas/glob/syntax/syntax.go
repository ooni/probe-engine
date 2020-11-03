package syntax

import (
	"github.com/ooni/probe-engine/internal/oopsi/github.com/gobwas/glob/syntax/ast"
	"github.com/ooni/probe-engine/internal/oopsi/github.com/gobwas/glob/syntax/lexer"
)

func Parse(s string) (*ast.Node, error) {
	return ast.Parse(lexer.NewLexer(s))
}

func Special(b byte) bool {
	return lexer.Special(b)
}
