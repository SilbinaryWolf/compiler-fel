package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
)

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

func (p *Parser) DetermineType(node *ast.Type) types.TypeInfo {
	var resultType types.TypeInfo

	str := node.Name.String()
	resultType = types.GetRegisteredType(str)
	if resultType == nil {
		p.addErrorToken(fmt.Errorf("Undeclared type \"%s\".", str), node.Name)
	}
	if node.ArrayDepth > 0 {
		for i := 0; i < node.ArrayDepth; i++ {
			arrayItemType := resultType
			resultType = types.Array(arrayItemType)
		}
	}
	return resultType
}
