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
		case token.Identifier:
			p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not identifier")
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
		case token.BraceOpen, token.BraceClose, token.Comma:
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
		case token.EOF:
			break Loop
		default:
			if t.IsOperator() {
				if !expectOperator {
					panic(fmt.Sprintf("Expected identifiers or string, instead got operator \"%s\" on Line %d.", t.String(), t.Line))
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
			panic(fmt.Sprintf("parseExpression(): Unhandled token type: \"%s\" (value: %s) on Line %d", t.Kind.String(), t.String(), t.Line))
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
