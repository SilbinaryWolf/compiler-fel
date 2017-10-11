package parser

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseCSSConfigRuleDefinition(name token.Token) *ast.CSSConfigDefinition {
	isNamedCSSDefinition := name.Kind != token.Unknown

	node := new(ast.CSSConfigDefinition)
	if isNamedCSSDefinition {
		node.Name = name
	}
	p.SetScanMode(scanner.ModeCSS)
	node.ChildNodes = p.parseCSSStatements()
	p.SetScanMode(scanner.ModeDefault)
	return nil
}

/*

func getNewSelectorPartList() []ast.Node {
	return make([]ast.Node, 0, 10)
}

func (p *Parser) parseCSSConfigSelector(firstToken token.Token) ast.CSSSelectorWildcard {
	selectorWildcardParts := make([]ast.Node, 0, 10)
	selectorWildcardParts = append(selectorWildcardParts, &ast.Token{Token: firstToken})

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier, token.Multiply:
			p.GetNextToken()
			selectorWildcardParts = append(selectorWildcardParts, &ast.Token{Token: t})
		case token.Comma:
			p.GetNextToken()
			break Loop
		case token.BraceOpen:
			break Loop
		case token.Whitespace, token.Newline:
			// no-op
			p.GetNextToken()
		default:
			panic(fmt.Sprintf("parseCSSConfigSelector(): Unhandled token type(%d): \"%s\" (value: %s) on Line %d", t.Kind, t.Kind.String(), t.String(), t.Line))
		}
	}

	result := ast.CSSSelectorWildcard{}
	result.ChildNodes = selectorWildcardParts
	return result
}

func (p *Parser) parseCSSConfigRules() []ast.Node {
	selectorWildcardList := make([]ast.CSSSelectorWildcard, 0, 3)

	for {
		p.eatWhitespace()
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			selector := p.parseCSSConfigSelector(t)
			selectorWildcardList = append(selectorWildcardList, selector)
		case token.BraceOpen:

		case token.Whitespace, token.Newline:
			// no-op
		default:
			panic(fmt.Sprintf("parseCSSConfigRules(): Unhandled token type(%d): \"%s\" (value: %s) on Line %d", t.Kind, t.Kind.String(), t.String(), t.Line))
		}
	}
	return nil
}*/
