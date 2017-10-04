package parser

import (
	//"encoding/json"

	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

func GetNewTokenList() []ast.Node {
	return make([]ast.Node, 0, 30)
}

func removeTrailingWhitespaceTokens(tokenList []ast.Node) []ast.Node {
	for i := len(tokenList) - 1; i >= 0; i-- {
		itNode := tokenList[i]
		node, ok := itNode.(*ast.Token)
		if !ok || node.Kind != token.Whitespace {
			// If not whitespace, consider the trimming complete
			break
		}
		// Cut off last element
		tokenList = tokenList[:i]
	}
	return tokenList
}

func (p *Parser) eatWhitespace() {
	for {
		t := p.PeekNextToken()
		if t.Kind != token.Whitespace {
			return
		}
		p.GetNextToken()
	}
}

func (p *Parser) parseCSS(name token.Token) *ast.CSSDefinition {
	isNamedCSSDefinition := name.Kind != token.Unknown

	node := new(ast.CSSDefinition)
	if isNamedCSSDefinition {
		node.Name = name
	}
	p.SetScanMode(scanner.ModeCSS)
	node.ChildNodes = p.parseCSSStatements()
	p.SetScanMode(scanner.ModeDefault)

	//{
	//	json, _ := json.MarshalIndent(resultNodes, "", "   ")
	//	fmt.Printf("%s", string(json))
	//}
	//panic("Finish parseCSS()")
	return node
}

func (p *Parser) parseCSSProperty(tokenList []ast.Node) *ast.CSSProperty {
	i := 0

	// Get property name
	var name token.Token
	for i < len(tokenList) {
		itNode := tokenList[i]
		i++

		tokenNode, ok := itNode.(*ast.Token)
		if !ok {
			panic(fmt.Sprintf("parseCSSStatements(): Expected property to be identifier type, not %T", itNode))
			return nil
		}
		if tokenNode.Kind == token.Whitespace {
			continue
		}
		name = tokenNode.Token
		break
	}

	// Check for declare op
	for i < len(tokenList) {
		itNode := tokenList[i]
		i++

		tokenNode, ok := itNode.(*ast.Token)
		if !ok {
			panic(fmt.Sprintf("parseCSSStatements(): Expected *ast.Token not %T.", itNode))
			return nil
		}
		if tokenNode.Kind == token.Whitespace {
			continue
		}
		if tokenNode.Kind != token.Colon {
			panic(fmt.Sprintf("parseCSSStatements(): Expected : after property name, not \"%s\" (Data: %s) on Line %d.", tokenNode.Kind.String(), tokenNode.Data, tokenNode.Line))
			return nil
		}
		// Found it!
		break
	}

	// Get remaining value nodes
	valueNodes := tokenList[i:len(tokenList)]

	cssPropertyNode := new(ast.CSSProperty)
	cssPropertyNode.Name = name
	cssPropertyNode.ChildNodes = valueNodes
	return cssPropertyNode
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
			tok, ok := tokenList[0].(*ast.Token)
			if !ok {
				panic(fmt.Sprintf("parseCSSStatements(): Invalid use of := on Line %d", t.Line))
			}
			var name token.Token = tok.Token
			//switch tok := tokenList[0].(type) {
			//case *ast.Token:
			//	name = tok.Token
			//default:
			//	panic(fmt.Sprintf("parseCSSStatements(): Invalid use of := on Line %d", t.Line))
			//}

			node := p.NewDeclareStatement(name, token.Token{}, p.parseExpressionNodes())
			resultNodes = append(resultNodes, node)

			// Clear
			tokenList = GetNewTokenList()
		case token.AtKeyword, token.Identifier, token.Number:
			// NOTE: We do -NOT- want to eat whitespace surrounding `token.Identifier`
			//       as that is used to detect / determine descendent selectors. (ie. ".top-class .descendent")
			tokenList = append(tokenList, &ast.Token{Token: t})
		case token.GreaterThan, token.Colon, token.DoubleColon:
			tokenList = removeTrailingWhitespaceTokens(tokenList)
			tokenList = append(tokenList, &ast.Token{Token: t})
			p.eatWhitespace()
		case token.Semicolon, token.Newline:
			if len(tokenList) == 0 {
				continue Loop
			}

			cssPropertyNode := p.parseCSSProperty(tokenList)
			if cssPropertyNode == nil {
				break Loop
			}
			resultNodes = append(resultNodes, cssPropertyNode)

			// Clear
			tokenList = GetNewTokenList()
		case token.Whitespace:
			if len(tokenList) == 0 {
				continue Loop
			}
			tokenList = append(tokenList, &ast.Token{Token: t})
		case token.Comma:
			tokenList = append(tokenList, &ast.Token{Token: t})

			// Consume any newlines to avoid end of statement if
			// getting a list of selectors.
			for {
				t := p.PeekNextToken()
				if t.Kind != token.Newline && t.Kind != token.Whitespace {
					break
				}
				p.GetNextToken()
			}
		case token.BraceOpen:
			if len(tokenList) == 0 {
				panic(fmt.Sprintf("parseCSSStatements(): Got {, expected identifiers preceding for CSS rule on Line %d", t.Line))
			}

			// Remove trailing whitespace tokens
			tokenList = removeTrailingWhitespaceTokens(tokenList)

			// Put selectors into a single array
			selectorList := make([]ast.CSSSelector, 0, 10)
			selector := ast.CSSSelector{}
			for _, itNode := range tokenList {
				switch node := itNode.(type) {
				case *ast.Token:
					if node.Kind == token.Comma {
						// Split into new selector
						// ie. .selector-1,
						//	   .selector-2 {
						//
						//	   }
						//
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

			// Determine and validate rule type
			kind := ast.CSSKindUnknown
			for _, selector := range selectorList {
				firstNode := selector.ChildNodes[0]
				switch firstToken := firstNode.(type) {
				case *ast.Token:
					switch firstToken.Kind {
					case token.AtKeyword:
						if kind != ast.CSSKindUnknown && kind != ast.CSSKindAtKeyword {
							panic("parseCSSStatements(): Cannot mix CSS rule then media query.")
						}
						kind = ast.CSSKindAtKeyword
						continue
					}
				}
				if kind == ast.CSSKindAtKeyword {
					panic("parseCSSStatements(): Cannot mix media query then CSS rule.")
				}
				kind = ast.CSSKindRule
			}
			if kind == ast.CSSKindUnknown {
				panic("parseCSSStatements: Unable to determine or validate CSS rule type")
			}

			// Add node
			rule := new(ast.CSSRule)
			rule.Kind = kind
			rule.Selectors = selectorList
			rule.ChildNodes = p.parseCSSStatements()
			resultNodes = append(resultNodes, rule)

			// Clear
			tokenList = GetNewTokenList()
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
