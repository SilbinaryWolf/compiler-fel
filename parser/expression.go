package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseExpression() *ast.Expression {
	node := new(ast.Expression)
	node.ChildNodes = p.parseExpressionNodes()
	return node
}

func (p *Parser) parseExpressionNodes() []ast.Node {
	parenOpenCount := 0
	parenCloseCount := 0

	expectOperator := false

	//childNodes := make([]ast.Node, 0, 10)
	infixNodes := make([]ast.Node, 0, 10)
	operatorNodes := make([]*ast.Token, 0, 10)

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier, token.KeywordTrue, token.KeywordFalse:
			p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not identifier.")
			}
			if p.PeekNextToken().Kind == token.ParenOpen {
				panic("parseExpression(): todo: Handle component/function in expression")
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.String:
			p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not string")
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.Semicolon:
			p.GetNextToken()
			break Loop
		case token.Newline:
			p.GetNextToken()
			// Allow expression to go over newline if operator is next
			//if t := p.PeekNextToken(); t.IsOperator() {
			//	continue
			//}
			break Loop
		case token.BraceOpen, token.BraceClose, token.Comma,
			token.EOF, token.Illegal:
			// NOTE(Jake): We specifically don't call p.GetNextToken()
			//			   so the calleee function can consume and use
			//			   the token.
			break Loop
		case token.Number:
			p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not number")
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.ParenOpen:
			parenOpenCount++
		case token.ParenClose:
			// If hit end of parameter list
			if parenCloseCount == 0 && parenOpenCount == 0 {
				break Loop
			}

			parenCloseCount++

			topOperatorNode := operatorNodes[len(operatorNodes)-1]
			if topOperatorNode.Kind == token.ParenOpen {
				infixNodes = append(infixNodes, topOperatorNode)
				operatorNodes = operatorNodes[:len(operatorNodes)-1]
			}
		// ie. :: css, :: html
		case token.DoubleColon:
			p.GetNextToken()
			node := p.parseDefinition(token.Token{})
			if node == nil {
				panic("parseExpressionNodes: parseDefinition returned nil")
			}
			infixNodes = append(infixNodes, node)
		// ie. []string{"item1", "item2", "item3"}
		case token.BracketOpen:
			typeIdent := p.parseType()
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.addErrorToken(p.expect(t, token.BraceOpen), t)
				return nil
			}

			node := new(ast.ArrayLiteral)
			node.TypeIdentifier = typeIdent

		ArrayLiteralLoop:
			for i := 0; true; i++ {
				expr := p.parseExpression()
				sep := p.GetNextToken()
				switch sep.Kind {
				case token.Comma:
					node.ChildNodes = append(node.ChildNodes, expr)
					continue
				case token.BraceClose:
					break ArrayLiteralLoop
				case token.EOF:
					p.addErrorToken(p.unexpected(sep), sep)
					return nil
				}
				p.addErrorToken(fmt.Errorf("Expected , or } after array item #%d, not %s.", i, sep.Kind.String()), sep)
				return nil
			}
			infixNodes = append(infixNodes, node)
			continue Loop
		default:
			if t.IsOperator() {
				if !expectOperator {
					p.addErrorToken(fmt.Errorf("Expected identifiers or string, instead got operator \"%s\".", t.String()), t)
					return nil
				}
				p.GetNextToken()
				expectOperator = false

				// https://github.com/SilbinaryWolf/fel/blob/master/c_compiler/parser.h
				for len(operatorNodes) > 0 {
					topOperatorNode := operatorNodes[len(operatorNodes)-1]
					if topOperatorNode.Precedence() < t.Precedence() {
						break
					}
					operatorNodes = operatorNodes[:len(operatorNodes)-1]
					infixNodes = append(infixNodes, topOperatorNode)
				}
				operatorNodes = append(operatorNodes, &ast.Token{Token: t})
				continue
			}
			p.fatalErrorToken(fmt.Errorf("Unhandled token type: \"%s\" (value: %s)", t.Kind.String(), t.String()), t)
			return nil
		}
	}

	for len(operatorNodes) > 0 {
		topOperatorNode := operatorNodes[len(operatorNodes)-1]
		operatorNodes = operatorNodes[:len(operatorNodes)-1]
		infixNodes = append(infixNodes, topOperatorNode)
	}

	if parenOpenCount != parenCloseCount {
		// todo(Jake): better error message
		panic("Mismatching paren open and close count")
	}

	// DEBUG
	//json, _ := json.MarshalIndent(infixNodes, "", "   ")
	//fmt.Printf("%s", string(json))
	//panic("todo: Finish parseExpression() func")

	return infixNodes
}
