package parser

import (
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseCSS() []ast.Node {
	p.SetScanMode(scanner.ModeCSS)
	resultNodes := p.parseCSSStatements()
	p.SetScanMode(scanner.ModeDefault)
	{
		json, _ := json.MarshalIndent(resultNodes, "", "   ")
		fmt.Printf("%s", string(json))
	}
	panic("Finish parseCSS()")
	return resultNodes
}

func (p *Parser) parseCSSStatements() []ast.Node {
	resultNodes := make([]ast.Node, 0, 10)
	tokenList := make([]ast.Node, 0, 5)
	selectorList := make([]ast.CSSSelector, 0, 3)

Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			if len(tokenList) > 0 {
				tokenList = append(tokenList, &ast.Token{Token: t})
				continue
			}
			switch nextT := p.PeekNextToken(); nextT.Kind {
			case token.Declare:
				p.GetNextToken() // :
				node := new(ast.CSSProperty)
				node.Name = t
			PropertyLoop:
				for {
					t := p.GetNextToken()
					switch t.Kind {
					case token.Identifier:
						tokenList = append(tokenList, &ast.Token{Token: t})
					case token.Newline, token.Semicolon:
						break PropertyLoop
					default:
						panic(fmt.Sprintf("Unhandled token type: %s in CSS property statement", t.Kind))
					}
				}
				node.ChildNodes = tokenList
				resultNodes = append(resultNodes, node)
			case token.Identifier, token.BraceOpen:
				tokenList = append(tokenList, &ast.Token{Token: t})
			default:
				panic(fmt.Sprintf("Unhandled token type: %s after CSS identifier", t.Kind))
			}
		case token.Declare: // :
			panic("parseCSSStatements(): Invalid :")
		case token.Comma:
			if len(tokenList) == 0 {
				// Ignore comma if no tokens
				continue
			}
			selectorNode := ast.CSSSelector{}
			selectorNode.ChildNodes = tokenList
			selectorList = append(selectorList, selectorNode)

			// Clear/create new slices
			tokenList = make([]ast.Node, 0, 5)
			/*resultNodes = append(resultNodes, selectorNode)

			// Clear/create new slices
			tokenList = make([]ast.Node, 0, 5)
			selectorList = make([]ast.CSSSelector, 0, 3)*/
		case token.BraceOpen:
			if len(tokenList) == 0 {
				panic("Got {, expected identifiers preceding")
			}
			selectorNode := ast.CSSSelector{}
			selectorNode.ChildNodes = tokenList
			selectorList = append(selectorList, selectorNode)
			rule := new(ast.CSSRule)
			rule.Selectors = selectorList
			rule.ChildNodes = p.parseCSSStatements()
			resultNodes = append(resultNodes, rule)

			// Clear/create new slices
			tokenList = make([]ast.Node, 0, 5)
			selectorList = make([]ast.CSSSelector, 0, 3)
		case token.BraceClose:
			// Finish statement
			break Loop
		case token.Newline:
			if len(tokenList) == 0 {
				continue
			}
			fallthrough
		case token.Semicolon:
			if len(tokenList) == 0 {
				panic("parseCSS(): Expected statement before ;")
			}
			panic("handle end of statement")
		default:
			panic(fmt.Sprintf("parseCSS(): Unhandled token type: \"%s\" (value: %s) on Line %d", t.Kind.String(), t.String(), t.Line))
		}
	}
	return resultNodes
}
