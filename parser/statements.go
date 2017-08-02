package parser

import (
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

// todo(Jake): Change to parseStatements, make it return []ast.Node slice
func (p *Parser) parseStatements() []ast.Node {
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
				node.ChildNodes = p.parseExpression()
				resultNodes = append(resultNodes, node)
			// div {
			//     ^
			case token.BraceOpen:
				p.GetNextToken()
				node := &ast.HTMLNode{
					Name: name,
				}
				node.ChildNodes = p.parseStatements()
				resultNodes = append(resultNodes, node)
			// div(class="hey")
			//    ^
			case token.ParenOpen:
				p.GetNextToken()
				node := &ast.HTMLNode{
					Name: name,
				}
				node.Parameters = p.parseParameters()
				if node.Parameters == nil {
					panic("parseStatements(): No parameters")
				}
				t = p.PeekNextToken()
				if t.Kind == token.BraceOpen {
					p.GetNextToken()
					node.ChildNodes = p.parseStatements()
				}
				resultNodes = append(resultNodes, node)
				//panic("Finish parseStatements() parseAttributes")
			default:
				panic(fmt.Sprintf("parseStatements(): Handle other ident case kind: %s", t.Kind.String()))
			}
		case token.BraceClose:
			break Loop
		case token.Newline:
			// no-op
		case token.EOF:
			break Loop
		default:
			p.PrintErrors()
			json, _ := json.MarshalIndent(resultNodes, "", "   ")
			fmt.Printf("%s", string(json))
			panic(fmt.Sprintf("parseStatements(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}
	return resultNodes
}
