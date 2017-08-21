package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseDefinition(name token.Token) ast.Node {
	keywordToken := p.GetNextToken()
	switch keyword := keywordToken.String(); keyword {
	case "css":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		node := new(ast.CSSDefinition)
		if name.Kind != token.Unknown {
			node.Name = name
		}
		node.ChildNodes = p.parseCSS()
		return node
	default:
		p.addError(fmt.Errorf("Unexpected keyword '%s' for definition (::) type. Expected 'css' on Line %d", keyword, keywordToken.Line))
		return nil
	}
	return nil
}
