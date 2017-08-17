package parser

import (
	//"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseCSS() []ast.Node {
	p.SetScanMode(scanner.ModeCSS)
	resultNodes := p.parseCSSStatements()
	p.SetScanMode(scanner.ModeDefault)
	//{
	//	json, _ := json.MarshalIndent(resultNodes, "", "   ")
	//	fmt.Printf("%s", string(json))
	//}
	//panic("Finish parseCSS()")
	return resultNodes
}

func (p *Parser) parseCSSStatements() []ast.Node {
	resultNodes := make([]ast.Node, 0, 10)
	tokenList := make([]ast.Node, 0, 30)
	selectorList := make([]ast.CSSSelector, 0, 10)

Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			tokenList = append(tokenList, &ast.Token{Token: t})
		case token.Declare, token.Define: // : or ::
			tokenList = append(tokenList, &ast.Token{Token: t})
		case token.Semicolon, token.Newline:
			if len(tokenList) == 0 {
				continue Loop
			}
			panic(fmt.Sprintf("parseCSSStatements(): Unexpected newline or ; at Line %d", t.Line))
		case token.Comma:
			tokenList = append(tokenList, &ast.Token{Token: t})

			// Consume any newlines
			for {
				t := p.PeekNextToken()
				if t.Kind != token.Newline {
					break
				}
				p.GetNextToken()
			}
		case token.BraceOpen:
			if len(tokenList) == 0 && len(selectorList) == 0 {
				panic(fmt.Sprintf("parseCSSStatements(): Got {, expected identifiers preceding for CSS rule on Line %d", t.Line))
			}
			panic("todo(Jake): Handle brace open")
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
			//selectorList = make([]ast.CSSSelector, 0, 3)
		case token.BracketOpen:
			node := new(ast.CSSAttributeSelector)
			node.Name = p.GetNextToken()
			tokenList = append(tokenList, node)
			if p.PeekNextToken().Kind == token.BracketClose {
				p.GetNextToken() // ]
				continue
			}
			switch operator := p.GetNextToken(); operator.Kind {
			case token.Equal:
				node.Operator = operator
			default:
				panic(fmt.Sprintf("parseCSSStatements(): Expected = on Line %s", operator.Line))
			}
			value := p.GetNextToken()
			switch value.Kind {
			case token.String:
				node.Value = value
			case token.Identifier:
				node.Value = value
			default:
				panic(fmt.Sprintf("parseCSSStatements(): Unexpected token in attribute after operator on Line %d", value.Line))
			}
			if t := p.GetNextToken(); t.Kind != token.BracketClose {
				panic("parseCSSStatements: Expected ]")
				p.addError(p.expect(t, token.BracketClose))
				break Loop
			}
		case token.BraceClose, token.ParenClose:
			// Finish statement
			break Loop
		case token.EOF:
			panic("parseCSSStatements(): Reached end of file, Should be closed with }")
		default:
			panic(fmt.Sprintf("parseCSSStatements(): Unhandled token type: \"%s\" (value: %s) on Line %d", t.Kind.String(), t.String(), t.Line))
		}
	}

	if len(tokenList) > 0 {
		selectorNode := new(ast.CSSSelector)
		selectorNode.ChildNodes = tokenList
		resultNodes = append(resultNodes, selectorNode)
		//node := new(ast.CSSTokens)
		//node.ChildNodes = tokenList
		//resultNodes = append(resultNodes, node)
	}

	return resultNodes
}
