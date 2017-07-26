package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseBlock() *ast.Block {
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			// p.parseStatement()
			resultNode := &ast.DeclareStatement{}
			resultNode.Name = t
			t := p.PeekNextToken()
			if t.Kind == token.DeclareSet {
				p.GetNextToken()
				node, err := p.parseExpression()
				if err != nil {
					// todo
				}
				resultNode.Expression = node
				panic("parseBlock(): Handle after parse expression")
			} else {
				panic(fmt.Sprintf("parseBlock(): Handle other ident case"))
			}
			panic("parseBlock(): Finish expression")
		default:
			panic(fmt.Sprintf("parseBlock(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}
	panic("todo: Finish parseBlock() func")
}
