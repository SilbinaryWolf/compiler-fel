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
			name := t
			switch t := p.PeekNextToken(); t.Kind {
			case token.DeclareSet:
				p.GetNextToken()
				node := &ast.DeclareStatement{}
				node.Name = name
				node.ChildNodes = p.parseExpression()
				resultNodes = append(resultNodes, node)
			case token.Declare:
				p.GetNextToken() // :
				node := new(ast.CSSProperty)
				node.Name = name
			PropertyLoop:
				for {
					t := p.GetNextToken()
					switch t.Kind {
					case token.Identifier, token.Number, token.Comma:
						tokenList = append(tokenList, &ast.Token{Token: t})
					case token.Newline, token.Semicolon:
						break PropertyLoop
					default:
						panic(fmt.Sprintf("parseCSSStatements(): Unhandled token type: %s in CSS property statement on Line %d", t.Kind, t.Line))
					}
				}
				node.ChildNodes = tokenList
				resultNodes = append(resultNodes, node)

				// Clear tokens
				tokenList = make([]ast.Node, 0, 5)
			case token.Identifier, token.BraceOpen, token.Comma:
				tokenList = append(tokenList, &ast.Token{Token: name})
			default:
				panic(fmt.Sprintf("parseCSSStatements(): Unhandled token type: %s after CSS identifier on Line %d", t.Kind, t.Line))
			}
		case token.Declare: // :
			{
				json, _ := json.MarshalIndent(tokenList, "", "   ")
				fmt.Printf("%s", string(json))
				panic(fmt.Sprintf("parseCSSStatements(): Invalid : on Line %d", t.Line))
			}
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
			if len(tokenList) == 0 && len(selectorList) == 0 {
				panic("parseCSSStatements(): Got {, expected identifiers preceding for CSS rule")
			}
			if len(tokenList) > 0 {
				selectorNode := ast.CSSSelector{}
				selectorNode.ChildNodes = tokenList
				selectorList = append(selectorList, selectorNode)
			}
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
			// no-op
		case token.Semicolon:
			panic("parseCSSStatements(): Unexpected ;")
		case token.EOF:
			panic("parseCSSStatements(): Reached end of file, Should be closed with }")
		default:
			panic(fmt.Sprintf("parseCSSStatements(): Unhandled token type: \"%s\" (value: %s) on Line %d", t.Kind.String(), t.String(), t.Line))
		}
	}
	return resultNodes
}
