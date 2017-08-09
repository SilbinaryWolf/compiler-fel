package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseCSS() []ast.Node {
	p.SetScanMode(scanner.ModeCSS)
	result := p.parseCSSBlock()
	p.SetScanMode(scanner.ModeDefault)
	return result
}

func (p *Parser) parseCSSBlock() []ast.Node {
	tokenList := make([]token.Token, 0, 40)

	//Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			tokenList = append(tokenList, t)
		case token.BraceOpen:
			if len(tokenList) == 0 {
				panic("Got {, expected identifiers preceding")
			}
			p.parseCSSBlock()
			panic("Parse CSS block")
		case token.Newline:
			if len(tokenList) == 0 {
				continue
			}
			fallthrough
		case token.Semicolon:
			panic("handle end of statement")
		default:
			panic(fmt.Sprintf("parseCSS(): Unhandled token type: \"%s\" (value: %s) on Line %d", t.Kind.String(), t.String(), t.Line))
		}
	}
	panic("Finish parseCSS()")
	return nil
}
