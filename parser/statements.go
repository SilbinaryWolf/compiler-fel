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

	nameString := name.String()
	if len(nameString) > 0 && nameString[len(nameString)-1] == '-' {
		p.addErrorToken(fmt.Errorf("Declaring variable name ending with - is illegal."), name)
	}

	return node
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
			// myVar = {Expression} \n
			//
			case token.Equal, token.AddEqual:
				node := new(ast.OpStatement)
				node.Name = name
				node.Operator = t
				node.Expression.ChildNodes = p.parseExpressionNodes()
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
				t := p.GetNextToken()
				switch t.Kind {
				case token.Newline:
					// no-op
				case token.BraceOpen:
					node.ChildNodes = p.parseStatements()
				case token.KeywordIf:
					node.IfExpression.ChildNodes = p.parseExpressionNodes()
					if t := p.GetNextToken(); t.Kind != token.BraceOpen {
						p.addErrorToken(p.expect(t, token.BraceOpen), t)
						return nil
					}
					node.ChildNodes = p.parseStatements()
				default:
					p.addErrorToken(p.expect(t, token.Newline, token.BraceOpen), t)
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
		case token.String:
			node := p.parseExpression()
			resultNodes = append(resultNodes, node)
		case token.BraceClose:
			p.GetNextToken()
			break Loop
		case token.Newline, token.Semicolon:
			// no-op
			p.GetNextToken()
		case token.KeywordIf:
			p.GetNextToken()
			exprNodes := p.parseExpressionNodes()
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.addErrorToken(p.expect(t, token.BraceOpen), t)
				return nil
			}
			node := new(ast.If)
			node.Condition.ChildNodes = exprNodes
			node.ChildNodes = p.parseStatements()
			resultNodes = append(resultNodes, node)
			// Eat newlines
			for {
				t := p.PeekNextToken()
				if t.Kind == token.Newline {
					p.GetNextToken()
					continue
				}
				break
			}
			t = p.PeekNextToken()
			if t.Kind == token.KeywordElse {
				p.GetNextToken()
				t := p.PeekNextToken()
				switch t.Kind {
				case token.KeywordIf:
					// no-op
				case token.BraceOpen:
					p.GetNextToken()
				default:
					p.addErrorToken(p.expect(t, token.BraceOpen, token.KeywordIf), t)
					return nil
				}
				node.ElseNodes = p.parseStatements()
			}
		case token.KeywordFor:
			p.GetNextToken()
			varName := p.GetNextToken()
			if varName.Kind != token.Identifier {
				p.addErrorToken(p.expect(varName, token.Identifier), varName)
				return nil
			}
			t := p.GetNextToken()

			var node *ast.For
			switch t.Kind {
			case token.DeclareSet:
				node = new(ast.For)
				node.IsDeclareSet = true
				node.RecordName = varName
				node.Array.ChildNodes = p.parseExpressionNodes()
			case token.Comma:
				secondVarName := p.GetNextToken()
				if secondVarName.Kind != token.Identifier {
					p.addErrorToken(p.expect(secondVarName, token.Identifier), secondVarName)
					return nil
				}
				t = p.GetNextToken()
				if t.Kind != token.DeclareSet {
					p.addErrorToken(p.expect(t, token.DeclareSet), t)
					return nil
				}
				node = new(ast.For)
				node.IsDeclareSet = true
				node.IndexName = varName
				node.RecordName = secondVarName
				node.Array.ChildNodes = p.parseExpressionNodes()
			default:
				p.addErrorToken(p.expect(t, token.DeclareSet, token.Comma), t)
				return nil
			}
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.addErrorToken(p.expect(t, token.BraceOpen), t)
				return nil
			}
			node.ChildNodes = p.parseStatements()
			resultNodes = append(resultNodes, node)
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
