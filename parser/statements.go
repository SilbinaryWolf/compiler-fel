package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/util"
)

// todo(Jake): Change to parseStatements, make it return []ast.Node slice
func (p *Parser) parseStatements() []ast.Node {
	resultNodes := make([]ast.Node, 0, 10)

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier:
			storeScannerState := p.ScannerState
			name := p.GetNextToken()
			switch t := p.GetNextToken(); t.Kind {
			// myVar := {Expression} \n
			//
			case token.DeclareSet:
				node := &ast.DeclareStatement{}
				node.Name = name
				node.ChildNodes = p.parseExpression()
				resultNodes = append(resultNodes, node)
			// div {
			//     ^
			case token.BraceOpen:
				node := &ast.HTMLNode{
					Name: name,
				}
				node.ChildNodes = p.parseStatements()
				if len(node.ChildNodes) > 0 && util.IsSelfClosingTagName(name.String()) {
					p.addError(fmt.Errorf("%s is a self-closing tag and cannot have child elements.", name.String()))
				}
				resultNodes = append(resultNodes, node)
			// div(class="hey")
			//    ^
			case token.ParenOpen:
				node := &ast.HTMLNode{
					Name: name,
				}
				node.Parameters = p.parseParameters()
				if p.PeekNextToken().Kind == token.BraceOpen {
					p.GetNextToken()
					node.ChildNodes = p.parseStatements()
					if len(node.ChildNodes) > 0 && util.IsSelfClosingTagName(name.String()) {
						p.addError(fmt.Errorf("%s is a self-closing tag and cannot have child elements.", name.String()))
					}
				}
				resultNodes = append(resultNodes, node)
			// PrintThisVariable \n
			// ^
			case token.Newline:
				p.ScannerState = storeScannerState
				node := new(ast.Expression)
				node.ChildNodes = p.parseExpression()
				resultNodes = append(resultNodes, node)
			// Normalize :: css {
			//			 ^
			case token.DoubleColon:
				node := p.parseDefinition(name)
				if node == nil {
					break Loop
				}
				resultNodes = append(resultNodes, node)
			default:
				if t.IsOperator() {
					p.ScannerState = storeScannerState
					node := new(ast.Expression)
					node.ChildNodes = p.parseExpression()
					resultNodes = append(resultNodes, node)
					continue
				}
				panic(fmt.Sprintf("parseStatements(): Handle other ident case kind: %s", t.Kind.String()))
			}
		// :: css {
		// ^
		// (anonymous definiton)
		case token.DoubleColon:
			p.GetNextToken()
			node := p.parseDefinition(token.Token{})
			if node == nil {
				break Loop
			}
			resultNodes = append(resultNodes, node)
		case token.String:
			node := new(ast.Expression)
			node.ChildNodes = p.parseExpression()
			resultNodes = append(resultNodes, node)
		case token.BraceClose:
			p.GetNextToken()
			break Loop
		case token.Newline:
			// no-op
			p.GetNextToken()
		case token.EOF, token.Illegal:
			break Loop
		default:
			p.GetNextToken()
			p.PrintErrors()
			//json, _ := json.MarshalIndent(resultNodes, "", "   ")
			//fmt.Printf("%s", string(json))
			panic(fmt.Sprintf("parseStatements(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}
	return resultNodes
}
