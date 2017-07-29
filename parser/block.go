package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseBlock() *ast.Block {
	resultNodes := make([]ast.Node, 0, 10)

Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			// p.parseStatement()
			node := &ast.DeclareStatement{}
			node.Name = t
			t := p.PeekNextToken()
			if t.Kind == token.DeclareSet {
				p.GetNextToken()
				node.Expression = p.parseExpression()
				resultNodes = append(resultNodes, node)
			} else {
				panic(fmt.Sprintf("parseBlock(): Handle other ident case"))
			}
		case token.BraceClose:
			break Loop
		case token.Newline:
			// no-op
		default:
			panic(fmt.Sprintf("parseBlock(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}

	resultNode := &ast.Block{}
	resultNode.ChildNodes = resultNodes
	return resultNode
}
