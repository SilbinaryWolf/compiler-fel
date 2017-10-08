package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) NewDeclareStatement(name token.Token, typeToken token.Token, expressionNodes []ast.Node) *ast.DeclareStatement {
	node := new(ast.DeclareStatement)
	node.Name = name
	node.TypeToken = typeToken
	node.ChildNodes = expressionNodes
	return node
}

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
				node := p.NewDeclareStatement(name, token.Token{}, p.parseExpressionNodes())
				resultNodes = append(resultNodes, node)
			// myVar : string \n
			case token.Colon:
				tType := p.GetNextToken()
				if tType.Kind != token.Identifier {
					p.addError(p.expect(tType, token.Identifier))
					return nil
				}
				var expressionNodes []ast.Node
				// myVar : string = {Expression} \n
				if p.PeekNextToken().Kind == token.Equal {
					p.GetNextToken()
					expressionNodes = p.parseExpressionNodes()
				}
				node := p.NewDeclareStatement(name, tType, expressionNodes)
				resultNodes = append(resultNodes, node)
			// div {
			//     ^
			case token.BraceOpen:
				node := &ast.HTMLNode{
					Name: name,
				}
				node.ChildNodes = p.parseStatements()
				p.checkHTMLNode(node)
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
				}
				p.checkHTMLNode(node)
				resultNodes = append(resultNodes, node)
			// PrintThisVariable \n
			// ^
			case token.Newline:
				p.ScannerState = storeScannerState
				node := p.parseExpression()
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
					node := p.parseExpression()
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
		case token.BraceOpen:
			panic("todo(Jake): Handle map data structure { \"thing\": [] }")
		case token.String:
			node := p.parseExpression()
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
			panic(fmt.Errorf("parseStatements(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}
	return resultNodes
}
