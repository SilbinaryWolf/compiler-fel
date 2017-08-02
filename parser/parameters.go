package parser

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseParameters() []ast.Parameter {
	resultNodes := make([]ast.Parameter, 0, 5)

Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			node := ast.Parameter{
				Name: t,
			}
			t := p.GetNextToken()
			if t.Kind != token.Equal {
				p.addError(p.expect(t, token.Equal))
			}
			node.ChildNodes = p.parseExpression()
			t = p.GetNextToken()
			if t.Kind != token.Comma && t.Kind != token.ParenClose {
				p.addError(p.expect(t, token.Comma, token.ParenClose))
				return nil
			}
			resultNodes = append(resultNodes, node)
			if t.Kind == token.ParenClose {
				break Loop
			}
		case token.ParenClose:
			break Loop
		default:
			p.addError(p.expect(t, token.Identifier))
			return nil
		}
	}
	return resultNodes
}
