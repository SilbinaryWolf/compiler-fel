package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

// todo(Jake): Change to parseStatements, make it return []ast.Node slice
func (p *Parser) parseBlock() *ast.Block {
	resultNodes := make([]ast.Node, 0, 10)

Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			name := t
			t := p.PeekNextToken()
			switch t.Kind {
			case token.DeclareSet:
				p.GetNextToken()
				node := &ast.DeclareStatement{}
				node.Name = name
				node.Expression = p.parseExpression()
				resultNodes = append(resultNodes, node)
			case token.BraceOpen:
				p.GetNextToken()
				subNode := p.parseBlock()
				node := &ast.HTMLNode{}
				node.Name = name
				node.ChildNodes = subNode.ChildNodes
				resultNodes = append(resultNodes, node)
			//case token.ParenOpen:
			default:
				panic(fmt.Sprintf("parseBlock(): Handle other ident case kind: %s", t.Kind.String()))
			}
		case token.BraceClose:
			break Loop
		case token.Newline:
			// no-op
		case token.EOF:
			break Loop
		default:
			panic(fmt.Sprintf("parseBlock(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}

	resultNode := &ast.Block{}
	resultNode.ChildNodes = resultNodes
	return resultNode
}
