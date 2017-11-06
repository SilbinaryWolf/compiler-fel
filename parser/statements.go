package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) NewDeclareStatement(name token.Token, typeIdent ast.Type, expressionNodes []ast.Node) *ast.DeclareStatement {
	node := new(ast.DeclareStatement)
	node.Name = name
	node.TypeIdentifier = typeIdent
	node.ChildNodes = expressionNodes
	return node
}

func (p *Parser) parseType() ast.Type {
	result := ast.Type{}

	t := p.GetNextToken()
	if t.Kind == token.BracketOpen {
		// Parse array / array-of-array / etc
		// ie. []string, [][]string, [][][]string, etc
		result.ArrayDepth = 1
		for {
			t = p.GetNextToken()
			if t.Kind != token.BracketClose {
				p.addErrorToken(p.expect(t, token.BracketClose), t)
				return result
			}
			t = p.GetNextToken()
			if t.Kind == token.BracketOpen {
				result.ArrayDepth++
				continue
			}
			break
		}
	}
	if t.Kind != token.Identifier {
		p.addErrorToken(p.expect(t, token.Identifier), t)
		return result
	}
	result.Name = t
	return result
}

func (p *Parser) parseStatements() []ast.Node {
	resultNodes := make([]ast.Node, 0, 10)

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier:
			storeScannerState := p.ScannerState()
			name := p.GetNextToken()
			switch t := p.GetNextToken(); t.Kind {
			// myVar := {Expression} \n
			//
			case token.DeclareSet:
				node := p.NewDeclareStatement(name, ast.Type{}, p.parseExpressionNodes())
				resultNodes = append(resultNodes, node)
			// myVar : string \n
			case token.Colon:
				typeName := p.parseType()
				// myVar : string = {Expression} \n
				var expressionNodes []ast.Node
				if p.PeekNextToken().Kind == token.Equal {
					p.GetNextToken()
					expressionNodes = p.parseExpressionNodes()
				}
				node := p.NewDeclareStatement(name, typeName, expressionNodes)
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
				p.SetScannerState(storeScannerState)
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
					p.SetScannerState(storeScannerState)
					node := p.parseExpression()
					resultNodes = append(resultNodes, node)
					continue
				}
				p.addErrorToken(fmt.Errorf("Unexpected %s after identifier.", t.Kind.String()), t)
				return nil
			}
		// :: css {
		// ^
		// (anonymous definiton)
		case token.DoubleColon:
			// NOTE(Jake): Passing :: token for unnamed definition
			//			   so the line/column can be reasoned about for errors.
			blankName := p.GetNextToken()
			blankName.Kind = token.Unknown
			blankName.Data = ""

			node := p.parseDefinition(blankName)
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
		case token.Newline, token.Semicolon:
			// no-op
			p.GetNextToken()
		case token.EOF, token.Illegal:
			break Loop
		default:
			p.GetNextToken()
			p.addErrorToken(p.unexpected(t), t)
			return nil
		}
	}
	return resultNodes
}
