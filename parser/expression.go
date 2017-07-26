package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseExpression() (*ast.Expression, error) {
	resultNode := &ast.Expression{}

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier, token.String:
			p.GetNextToken()
			resultNode.Nodes = append(resultNode.Nodes, ast.Token{Token: t})
		// todo(Jake): Update scanner to get newline tokens and handle
		case token.BraceOpen, token.BraceClose:
			break Loop
		default:
			if t.IsOperator() {
				p.GetNextToken()
				resultNode.Nodes = append(resultNode.Nodes, ast.Token{Token: t})
				continue
			}
			panic(fmt.Sprintf("parseExpression(): Unhandled token type: \"%s\" (value: %s) on Line %d", t.Kind.String(), t.String(), t.Line))
		}
	}
	panic("todo: Finish parseExpression() func")
	return resultNode, nil
}
