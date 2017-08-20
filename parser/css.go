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

Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.DeclareSet:
			if len(tokenList) != 1 {
				panic(fmt.Sprintf("parseCSSStatements(): Invalid use of := on Line %d", t.Line))
			}
			var name token.Token
			switch tok := tokenList[0].(type) {
			case *ast.Token:
				name = tok.Token
			default:
				panic(fmt.Sprintf("parseCSSStatements(): Invalid use of := on Line %d", t.Line))
			}

			node := &ast.DeclareStatement{}
			node.Name = name
			node.ChildNodes = p.parseExpression()
			resultNodes = append(resultNodes, node)

			// Clear
			tokenList = make([]ast.Node, 0, 30)
		case token.AtKeyword, token.Identifier, token.Colon, token.DoubleColon, token.Number:
			tokenList = append(tokenList, &ast.Token{Token: t})
		case token.Semicolon, token.Newline:
			if len(tokenList) == 0 {
				continue Loop
			}

			// Get property name
			var name token.Token
			nameNode := tokenList[0]
			switch tokenNode := nameNode.(type) {
			case *ast.Token:
				if tokenNode.Kind != token.Identifier {
					panic(fmt.Sprintf("parseCSSStatements(): Expected property name to be identifier, not %s on Line %d", tokenNode.Kind.String(), tokenNode.Line))
					break Loop
				}
				name = tokenNode.Token
			default:
				panic(fmt.Sprintf("parseCSSStatements(): Expected property to be identifier type, not %T"))
			}

			if len(tokenList) < 3 {
				panic(fmt.Sprintf("parseCSSStatements(): Unexpected token count on property statement on Line %d", name.Line))
				break Loop
			}

			// Get declare op
			declareOpNode := tokenList[1]
			switch tokenNode := declareOpNode.(type) {
			case *ast.Token:
				if tokenNode.Kind != token.Colon {
					panic(fmt.Sprintf("parseCSSStatements(): Expected : after property name, not %s", tokenNode.Kind.String()))
					break Loop
				}
			default:
				panic(fmt.Sprintf("parseCSSStatements(): Expected property to begin with identifier type, not %T"))
			}
			if len(tokenList) < 3 {
				panic(fmt.Sprintf("parseCSSStatements(): Expected property statement on Line %d"))
				break Loop
			}

			valueNodes := tokenList[2:len(tokenList)]
			/*for _, val := range valueNodes {
				// something have to do with val
			}*/
			cssPropertyNode := new(ast.CSSProperty)
			cssPropertyNode.Name = name
			cssPropertyNode.ChildNodes = valueNodes
			resultNodes = append(resultNodes, cssPropertyNode)

			// Clear
			tokenList = make([]ast.Node, 0, 30)
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
			if len(tokenList) == 0 {
				panic(fmt.Sprintf("parseCSSStatements(): Got {, expected identifiers preceding for CSS rule on Line %d", t.Line))
			}

			// Put selectors into a single array
			selectorList := make([]ast.CSSSelector, 0, 10)
			selector := ast.CSSSelector{}
			for _, itNode := range tokenList {
				switch node := itNode.(type) {
				case *ast.Token:
					if node.Kind == token.Comma {
						if len(selector.ChildNodes) > 0 {
							selectorList = append(selectorList, selector)
							selector = ast.CSSSelector{}
						}
						continue
					}
				}
				selector.ChildNodes = append(selector.ChildNodes, itNode)
			}
			if len(selector.ChildNodes) > 0 {
				selectorList = append(selectorList, selector)
			}

			// Add node
			rule := new(ast.CSSRule)
			rule.Selectors = selectorList
			rule.ChildNodes = p.parseCSSStatements()
			resultNodes = append(resultNodes, rule)

			// Clear
			tokenList = make([]ast.Node, 0, 30)
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
		case token.ParenOpen:
			nodes := p.parseCSSStatements()
			if len(nodes) == 0 {
				panic(fmt.Sprintf("parseCSSStatements(): Expected a node inside () on Line %d", t.Line))
			}
			if len(nodes) > 1 {
				panic(fmt.Sprintf("parseCSSStatements(): Too many nodes inside () on Line %d", t.Line))
			}
			tokenList = append(tokenList, nodes[0])
		case token.BraceClose, token.ParenClose:
			// Finish statement
			break Loop
		case token.EOF:
			panic("parseCSSStatements(): Reached end of file, Should be closed with }")
		default:
			panic(fmt.Sprintf("parseCSSStatements(): Unhandled token type(%d): \"%s\" (value: %s) on Line %d", t.Kind, t.Kind.String(), t.String(), t.Line))
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
