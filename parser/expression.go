package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
)

func (p *Parser) parseExpression() (*ast.Expression, error) {
	resultNode := &ast.Expression{}

	for {
		t := p.GetNextToken()
		switch t.Kind {
		//case token.Identifier:

		default:
			panic(fmt.Sprintf("parseExpression(): Unhandled token type: \"%s\" (value: %s) on Line %d", t.Kind.String(), t.String(), t.Line))
		}
	}
	panic("todo: Finish parseExpression() func")
	return resultNode, nil
}
