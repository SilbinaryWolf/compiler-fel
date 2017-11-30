package parser

import (
	"fmt"
	"io/ioutil"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/util"
)

type Parser struct {
	scanner.Scanner
	typeinfo                      TypeInfoManager
	errors                        map[string][]error
	typecheckHtmlNodeDependencies map[string]*ast.HTMLNode
}

func New() *Parser {
	p := new(Parser)
	p.errors = make(map[string][]error)
	p.typeinfo.Init()
	//p.typecheckHtmlDefinitionDependencies = make(map[string]*ast.HTMLComponentDefinition)
	//p.typecheckHtmlDefinitionStack = make([]*ast.HTMLComponentDefinition, 0, 20)
	return p
}

func (p *Parser) ParseFile(filepath string) (*ast.File, error) {
	filecontentsAsBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	result := p.Parse(filecontentsAsBytes, filepath)
	return result, nil
}

func (p *Parser) Parse(filecontentsAsBytes []byte, filepath string) *ast.File {
	p.Scanner = *scanner.New(filecontentsAsBytes, filepath)
	resultNode := &ast.File{
		Filepath: filepath,
	}
	resultNode.ChildNodes = p.parseStatements()
	return resultNode
}

func (p *Parser) checkHTMLNode(node *ast.HTMLNode) {
	name := node.Name.String()
	if len(node.ChildNodes) > 0 && util.IsSelfClosingTagName(name) {
		p.addErrorToken(fmt.Errorf("%s is a self-closing tag and cannot have child elements.", name), node.Name)
	}

	//
	// todo(Jake): Extend this to allow user configured/whitelisted tag names
	//
	//isValidHTML5TagName := util.IsValidHTML5TagName(name)
	//if !isValidHTML5TagName {
	//p.htmlComponentNodes = append(p.htmlComponentNodes, node)
	//p.addErrorLine(fmt.Errorf("\"%s\" is not a valid HTML5 element.", name), node.Name.Line)
	//}
}

func (p *Parser) NewDeclareStatement(name token.Token, typeIdent ast.Type, expressionNodes []ast.Node) *ast.DeclareStatement {
	node := new(ast.DeclareStatement)
	node.Name = name
	node.TypeIdentifier = typeIdent
	node.ChildNodes = expressionNodes

	nameString := name.String()
	if len(nameString) > 0 && nameString[len(nameString)-1] == '-' {
		p.addErrorToken(fmt.Errorf("Declaring variable name ending with - is illegal."), name)
	}

	return node
}

func (p *Parser) eatNewlines() {
	t := p.PeekNextToken()
	for t.Kind == token.Newline {
		p.GetNextToken()
		t = p.PeekNextToken()
	}
}

func (p *Parser) parseStatements() []ast.Node {
	resultNodes := make([]ast.Node, 0, 10)

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier:
			storeScannerState := p.ScannerState()
			name := p.GetNextToken()
			switch t := p.GetNextToken(); t.Kind {
			// myVar := {Expression} \n
			//
			case token.DeclareSet:
				node := p.NewDeclareStatement(name, ast.Type{}, p.parseExpressionNodes(false))
				resultNodes = append(resultNodes, node)
			// myVar = {Expression} \n
			//
			case token.Equal, token.AddEqual:
				node := new(ast.OpStatement)
				node.Name = name
				node.Operator = t
				node.Expression.ChildNodes = p.parseExpressionNodes(false)
				resultNodes = append(resultNodes, node)
			// myVar : string \n
			case token.Colon:
				typeName := p.parseType()
				// myVar : string = {Expression} \n
				var expressionNodes []ast.Node
				if p.PeekNextToken().Kind == token.Equal {
					p.GetNextToken()
					expressionNodes = p.parseExpressionNodes(false)
				}
				node := p.NewDeclareStatement(name, typeName, expressionNodes)
				resultNodes = append(resultNodes, node)
			// div if {expr} {
			//     ^
			case token.KeywordIf:
				node := &ast.HTMLNode{
					Name: name,
				}
				node.IfExpression.ChildNodes = p.parseExpressionNodes(true)
				if t := p.GetNextToken(); t.Kind != token.BraceOpen {
					p.addErrorToken(p.expect(t, token.BraceOpen), t)
					return nil
				}
				if t := p.GetNextToken(); t.Kind == token.BraceOpen {
					p.expect(t, token.BraceOpen)
					return nil
				}
				p.GetNextToken()
				node.ChildNodes = p.parseStatements()
			// div {
			//     ^
			case token.BraceOpen:
				node := &ast.HTMLNode{
					Name: name,
				}
				node.ChildNodes = p.parseStatements()
				p.checkHTMLNode(node)
				resultNodes = append(resultNodes, node)
			// div(class="hey")  -or-  div(class="hey") if expr {
			//    ^						  ^
			case token.ParenOpen:
				node := &ast.HTMLNode{
					Name: name,
				}
				node.Parameters = p.parseParameters()
				t := p.GetNextToken()
				switch t.Kind {
				case token.Newline:
					// no-op
				case token.BraceOpen:
					node.ChildNodes = p.parseStatements()
				case token.KeywordIf:
					node.IfExpression.ChildNodes = p.parseExpressionNodes(true)
					if t := p.GetNextToken(); t.Kind != token.BraceOpen {
						p.addErrorToken(p.expect(t, token.BraceOpen), t)
						return nil
					}
					node.ChildNodes = p.parseStatements()
				default:
					p.addErrorToken(p.expect(t, token.Newline, token.BraceOpen), t)
				}
				p.checkHTMLNode(node)
				resultNodes = append(resultNodes, node)
			// PrintThisVariable \n
			// ^
			case token.Newline:
				p.SetScannerState(storeScannerState)
				node := p.parseExpression(false)
				resultNodes = append(resultNodes, node)
			// Normalize :: css {
			//			 ^
			case token.DoubleColon:
				node := p.parseDefinition(name)
				if node == nil {
					break Loop
				}
				resultNodes = append(resultNodes, node)
			default:
				if t.IsOperator() {
					p.SetScannerState(storeScannerState)
					node := p.parseExpression(false)
					resultNodes = append(resultNodes, node)
					continue
				}
				p.addErrorToken(fmt.Errorf("Unexpected %s after identifier.", t.Kind.String()), t)
				return nil
			}
		// :: css {
		// ^
		// (anonymous definiton)
		case token.DoubleColon:
			// NOTE(Jake): Passing :: token for unnamed definition
			//			   so the line/column can be reasoned about for errors.
			blankName := p.GetNextToken()
			blankName.Kind = token.Unknown
			blankName.Data = ""

			node := p.parseDefinition(blankName)
			if node == nil {
				break Loop
			}
			resultNodes = append(resultNodes, node)
		case token.String:
			node := p.parseExpression(false)
			resultNodes = append(resultNodes, node)
		case token.BraceClose:
			p.GetNextToken()
			break Loop
		case token.Newline, token.Semicolon:
			// no-op
			p.GetNextToken()
		case token.KeywordIf:
			p.GetNextToken()
			// NOTE(Jake): Disable struct literal in if-statement as
			//			   the parser needs to understand '{' is the
			//			   start of the if-block when testing boolean exxpressions
			//
			//			   ie. "if myBool {" vs "if myStructLit{val:3} {"
			//
			exprNodes := p.parseExpressionNodes(true)
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.addErrorToken(p.expect(t, token.BraceOpen), t)
				return nil
			}
			node := new(ast.If)
			node.Condition.ChildNodes = exprNodes
			node.ChildNodes = p.parseStatements()
			resultNodes = append(resultNodes, node)
			// Eat newlines
			for {
				t := p.PeekNextToken()
				if t.Kind == token.Newline {
					p.GetNextToken()
					continue
				}
				break
			}
			t = p.PeekNextToken()
			if t.Kind == token.KeywordElse {
				p.GetNextToken()
				t := p.PeekNextToken()
				switch t.Kind {
				case token.KeywordIf:
					// no-op
				case token.BraceOpen:
					p.GetNextToken()
				default:
					p.addErrorToken(p.expect(t, token.BraceOpen, token.KeywordIf), t)
					return nil
				}
				node.ElseNodes = p.parseStatements()
			}
		case token.KeywordFor:
			p.GetNextToken()
			varName := p.GetNextToken()
			if varName.Kind != token.Identifier {
				p.addErrorToken(p.expect(varName, token.Identifier), varName)
				return nil
			}
			t := p.GetNextToken()

			var node *ast.For
			switch t.Kind {
			case token.DeclareSet:
				node = new(ast.For)
				node.IsDeclareSet = true
				node.RecordName = varName
				node.Array.ChildNodes = p.parseExpressionNodes(false)
			case token.Comma:
				secondVarName := p.GetNextToken()
				if secondVarName.Kind != token.Identifier {
					p.addErrorToken(p.expect(secondVarName, token.Identifier), secondVarName)
					return nil
				}
				t = p.GetNextToken()
				if t.Kind != token.DeclareSet {
					p.addErrorToken(p.expect(t, token.DeclareSet), t)
					return nil
				}
				node = new(ast.For)
				node.IsDeclareSet = true
				node.IndexName = varName
				node.RecordName = secondVarName
				node.Array.ChildNodes = p.parseExpressionNodes(false)
			default:
				p.addErrorToken(p.expect(t, token.DeclareSet, token.Comma), t)
				return nil
			}
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.addErrorToken(p.expect(t, token.BraceOpen), t)
				return nil
			}
			node.ChildNodes = p.parseStatements()
			resultNodes = append(resultNodes, node)
		case token.EOF, token.Illegal:
			break Loop
		default:
			p.GetNextToken()
			p.addErrorToken(p.unexpected(t), t)
			return nil
		}
	}
	return resultNodes
}

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
			node.ChildNodes = p.parseExpressionNodes(false)
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

func (p *Parser) parseExpression(disableStructLiteral bool) *ast.Expression {
	node := new(ast.Expression)
	node.ChildNodes = p.parseExpressionNodes(disableStructLiteral)
	return node
}

func (p *Parser) parseExpressionNodes(disableStructLiteral bool) []ast.Node {
	parenOpenCount := 0
	parenCloseCount := 0

	expectOperator := false

	//childNodes := make([]ast.Node, 0, 10)
	infixNodes := make([]ast.Node, 0, 10)
	operatorNodes := make([]*ast.Token, 0, 10)

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier:
			name := p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not identifier.")
			}
			switch t := p.PeekNextToken(); t.Kind {
			case token.ParenOpen:
				p.fatalError(fmt.Errorf("todo: Handle component/function in expression"))
				return nil
			case token.BraceOpen:
				if disableStructLiteral {
					// Dont parse struct literal and use identifier as-is
					// (ie. "if isBool {")
					break
				}
				p.GetNextToken()

				if t := p.PeekNextToken(); t.Kind == token.BraceClose {
					node := new(ast.StructLiteral)
					node.Name = name
					expectOperator = true
					infixNodes = append(infixNodes, node)
					continue Loop
				}

				var errorMsgLastToken token.Token
				fields := make([]ast.StructLiteralField, 0, 10)
			StructLiteralLoop:
				for i := 0; true; i++ {
					propertyName := p.GetNextToken()
					for propertyName.Kind == token.Newline {
						propertyName = p.GetNextToken()
					}
					if propertyName.Kind != token.Identifier {
						if i == 0 {
							p.addErrorToken(fmt.Errorf("Expected identifier after %s{ not %s", name, t.Kind.String()), t)
							return nil
						}
						p.addErrorToken(fmt.Errorf("Expected identifier after \"%s\" not %s", errorMsgLastToken, t.Kind.String()), t)
						return nil
					}
					if t := p.GetNextToken(); t.Kind != token.Colon {
						if i == 0 {
							p.addErrorToken(fmt.Errorf("Expected : after \"%s{%s\"", name.String(), propertyName.String()), t)
							return nil
						}
						p.addErrorToken(fmt.Errorf("Expected : after property \"%s\"", propertyName.String()), t)
						return nil
					}
					exprNodes := p.parseExpressionNodes(false)
					node := ast.StructLiteralField{}
					node.Name = propertyName
					node.ChildNodes = exprNodes
					fields = append(fields, node)
					switch t := p.GetNextToken(); t.Kind {
					case token.BraceClose:
						break StructLiteralLoop
					case token.Comma:
						// OK
					default:
						p.addErrorToken(p.expect(t, token.BraceClose, token.Comma), t)
						return nil
					}
					p.eatNewlines()
					if t := p.PeekNextToken(); t.Kind == token.BraceClose {
						// Allow for trailing comma
						p.GetNextToken()
						break StructLiteralLoop
					}
					errorMsgLastToken = propertyName
				}
				node := new(ast.StructLiteral)
				node.Name = name
				node.Fields = fields
				expectOperator = true
				infixNodes = append(infixNodes, node)
				continue Loop
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.KeywordTrue, token.KeywordFalse:
			p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not identifier.")
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.String:
			p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not string")
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.Semicolon:
			p.GetNextToken()
			break Loop
		case token.Newline:
			p.GetNextToken()
			// Allow expression to go over newline if operator is next
			//if t := p.PeekNextToken(); t.IsOperator() {
			//	continue
			//}
			break Loop
		case token.BraceOpen, token.BraceClose, token.Comma,
			token.EOF, token.Illegal:
			// NOTE(Jake): We specifically don't call p.GetNextToken()
			//			   so the calleee function can consume and use
			//			   the token.
			break Loop
		case token.Number:
			p.GetNextToken()
			if expectOperator {
				panic("Expected operator, not number")
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.ParenOpen:
			parenOpenCount++
		case token.ParenClose:
			// If hit end of parameter list
			if parenCloseCount == 0 && parenOpenCount == 0 {
				break Loop
			}

			parenCloseCount++

			topOperatorNode := operatorNodes[len(operatorNodes)-1]
			if topOperatorNode.Kind == token.ParenOpen {
				infixNodes = append(infixNodes, topOperatorNode)
				operatorNodes = operatorNodes[:len(operatorNodes)-1]
			}
		// ie. :: css, :: html
		case token.DoubleColon:
			p.GetNextToken()
			node := p.parseDefinition(token.Token{})
			if node == nil {
				panic("parseExpressionNodes: parseDefinition returned nil")
			}
			infixNodes = append(infixNodes, node)
		// ie. []string{"item1", "item2", "item3"}
		case token.BracketOpen:
			typeIdent := p.parseType()
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.addErrorToken(p.expect(t, token.BraceOpen), t)
				return nil
			}

			childNodes := make([]ast.Node, 10)

		ArrayLiteralLoop:
			for i := 0; true; i++ {
				expr := p.parseExpression(false)
				sep := p.GetNextToken()
				switch sep.Kind {
				case token.Comma:
					childNodes = append(childNodes, expr)
					continue
				case token.BraceClose:
					childNodes = append(childNodes, expr)
					break ArrayLiteralLoop
				case token.EOF:
					p.addErrorToken(p.unexpected(sep), sep)
					return nil
				}
				p.addErrorToken(fmt.Errorf("Expected , or } after array item #%d, not %s.", i, sep.Kind.String()), sep)
				return nil
			}

			node := new(ast.ArrayLiteral)
			node.TypeIdentifier = typeIdent
			node.ChildNodes = childNodes

			if len(node.ChildNodes) == 0 {
				p.addErrorToken(fmt.Errorf("Cannot have array literal with zero elements."), node.TypeIdentifier.Name)
			}

			infixNodes = append(infixNodes, node)
			continue Loop
		default:
			if t.IsOperator() {
				if !expectOperator {
					p.addErrorToken(fmt.Errorf("Expected identifiers or string, instead got operator \"%s\".", t.String()), t)
					return nil
				}
				p.GetNextToken()
				expectOperator = false

				// https://github.com/SilbinaryWolf/fel/blob/master/c_compiler/parser.h
				for len(operatorNodes) > 0 {
					topOperatorNode := operatorNodes[len(operatorNodes)-1]
					if topOperatorNode.Precedence() < t.Precedence() {
						break
					}
					operatorNodes = operatorNodes[:len(operatorNodes)-1]
					infixNodes = append(infixNodes, topOperatorNode)
				}
				operatorNodes = append(operatorNodes, &ast.Token{Token: t})
				continue
			}
			p.fatalErrorToken(fmt.Errorf("Unhandled token type: \"%s\" (value: %s)", t.Kind.String(), t.String()), t)
			return nil
		}
	}

	for len(operatorNodes) > 0 {
		topOperatorNode := operatorNodes[len(operatorNodes)-1]
		operatorNodes = operatorNodes[:len(operatorNodes)-1]
		infixNodes = append(infixNodes, topOperatorNode)
	}

	if parenOpenCount != parenCloseCount {
		// todo(Jake): better error message
		panic("Mismatching paren open and close count")
	}

	// DEBUG
	//json, _ := json.MarshalIndent(infixNodes, "", "   ")
	//fmt.Printf("%s", string(json))
	//panic("todo: Finish parseExpression() func")

	return infixNodes
}

func (p *Parser) parseDefinition(name token.Token) ast.Node {
	keywordToken := p.GetNextToken()
	keyword := keywordToken.String()
	switch keyword {
	case "css":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		node := p.parseCSS(name)
		return node
	case "css_config":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		node := p.parseCSSConfigRuleDefinition(name)
		return node
	case "struct":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		//
		//
		childNodes := p.parseStatements()
		fields := make([]ast.StructField, 0, len(childNodes))
		// NOTE(Jake): A bit of a hack, we should have a 'parseStruct' function
		for _, itNode := range childNodes {
			switch node := itNode.(type) {
			case *ast.DeclareStatement:
				field := ast.StructField{}
				field.Name = node.Name
				//field.Expression.TypeIdentifier = node.Expression.TypeIdentifier
				field.Expression = node.Expression
				fields = append(fields, field)
			default:
				p.addErrorToken(fmt.Errorf("Expected statement, instead got %T.", itNode), name)
				return nil
			}
		}
		node := new(ast.StructDefinition)
		node.Name = name
		node.Fields = fields
		return node
	case "html":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		childNodes := p.parseStatements()

		// Check HTML nodes
		htmlNodeCount := 0
		for _, itNode := range childNodes {
			_, ok := itNode.(*ast.HTMLNode)
			if !ok {
				continue
			}
			htmlNodeCount++
		}
		if htmlNodeCount == 0 || htmlNodeCount > 1 {
			var nameString string
			if name.Kind != token.Unknown {
				nameString = name.String() + " "
			}
			if htmlNodeCount == 0 {
				p.addErrorToken(fmt.Errorf("\"%s:: html\" must contain one HTML node at the top-level.", nameString), name)
			}
			// NOTE: No longer applicable.
			//if htmlNodeCount > 1 {
			//	p.addErrorToken(fmt.Errorf("\"%s:: html\" cannot have multiple HTML nodes at the top-level.", nameString), name)
			//}
		}

		if name.Kind != token.Unknown {
			// Retrieve properties block
			var cssDef *ast.CSSDefinition
			var structDef *ast.StructDefinition
		RetrievePropertyDefinitionLoop:
			for _, itNode := range childNodes {
				switch node := itNode.(type) {
				case *ast.StructDefinition:
					if structDef != nil {
						p.addError(fmt.Errorf("Cannot declare \":: struct\" twice in the same HTML component."))
						break RetrievePropertyDefinitionLoop
					}
					structDef = node
				case *ast.CSSDefinition:
					if cssDef != nil {
						p.addError(fmt.Errorf("Cannot declare \":: css\" twice in the same HTML component."))
						break RetrievePropertyDefinitionLoop
					}
					cssDef = node
				}
			}

			// Component
			node := new(ast.HTMLComponentDefinition)
			node.Name = name
			node.Struct = structDef
			node.CSSDefinition = cssDef
			node.ChildNodes = childNodes

			return node
		}

		// TODO(Jake): Disallow ":: properties" block in un-named HTML
		node := new(ast.HTMLBlock)
		node.HTMLKeyword = keywordToken
		node.ChildNodes = childNodes
		return node
	}
	p.addError(fmt.Errorf("Unexpected keyword '%s' for definition (::) type. Expected 'css', 'html' or 'properties' on Line %d", keyword, keywordToken.Line))
	return nil
}

func (p *Parser) parseCSSConfigRuleDefinition(name token.Token) *ast.CSSConfigDefinition {
	p.SetScanMode(scanner.ModeCSS)
	nodes := p.parseCSSStatements()
	p.SetScanMode(scanner.ModeDefault)

	// Check / read data from ast
	cssConfigDefinition := new(ast.CSSConfigDefinition)
	cssConfigDefinition.Name = name
	//cssConfigDefinition.Rules = make([]ast.CSSConfigMatchPart, 0, len(nodes))
	for _, itNode := range nodes {
		switch node := itNode.(type) {
		case *ast.CSSRule:
			configRule := ast.NewCSSConfigRule()

			// Get rules
			for _, itNode := range node.ChildNodes {
				switch node := itNode.(type) {
				case *ast.CSSProperty:
					name := node.Name.String()
					switch name {
					case "modify":
						value, ok := p.getBoolFromCSSConfigProperty(node)
						if !ok {
							return nil
						}
						configRule.Modify = value
					default:
						p.addErrorToken(fmt.Errorf("Invalid config key \"%s\". Expected \"modify\".", name), node.Name)
						return nil
					}
				case *ast.DeclareStatement:
					p.addErrorToken(fmt.Errorf("Cannot declare variables in a css_config block. Did you mean to use : instead of :="), node.Name)
					return nil
				default:
					panic(fmt.Sprintf("parseCSSConfigRuleDefinition:propertyLoop: Unknown type %T", node))
				}
			}

			// Get matching parts
			for _, selector := range node.Selectors {
				rulePartList := make(ast.CSSConfigMatchPart, 0, len(selector.ChildNodes))
				for _, itSelectorPart := range selector.ChildNodes {
					switch selectorPartNode := itSelectorPart.(type) {
					case *ast.Token:
						if selectorPartNode.IsOperator() {
							operator := selectorPartNode.Kind.String()
							switch selectorPartNode.Kind {
							case token.Multiply:
								rulePartList = append(rulePartList, operator)
							default:
								p.addErrorToken(fmt.Errorf("Only supports * wildcard, not %s", operator), selectorPartNode.Token)
								return nil
							}
							continue
						}
						if selectorPartNode.Kind != token.Identifier {
							p.addErrorToken(fmt.Errorf("Expected identifier, instead got %s", selectorPartNode.Kind.String()), selectorPartNode.Token)
							return nil
						}
						name := selectorPartNode.String()
						rulePartList = append(rulePartList, name)
					default:
						panic(fmt.Sprintf("parseCSSConfigRuleDefinition:selectorPartLoop: Unknown type %T", selectorPartNode))
					}
				}
				configRule.Selectors = append(configRule.Selectors, rulePartList)
			}

			// Generate string. (For easy feeding into path.Match() function)
			for _, selector := range configRule.Selectors {
				pattern := ""
				for _, part := range selector {
					pattern += part
				}
				configRule.SelectorsAsPattern = append(configRule.SelectorsAsPattern, pattern)
			}

			cssConfigDefinition.Rules = append(cssConfigDefinition.Rules, configRule)
		case *ast.DeclareStatement:
			p.addErrorToken(fmt.Errorf("Cannot declare variables in a css_config block."), node.Name)
			return nil
		default:
			panic(fmt.Sprintf("parseCSSConfigRuleDefinition: Unknown type %T", node))
		}
	}

	// Test
	//config := cssConfigDefinition.GetRule(".js-")
	//fmt.Printf("\n\n %v \n\n", config)
	//panic("parser/css_config.go test")

	return cssConfigDefinition
}

func (p *Parser) getBoolFromCSSConfigProperty(node *ast.CSSProperty) (bool, bool) {
	if len(node.ChildNodes) == 0 && len(node.ChildNodes) > 1 {
		p.addErrorToken(fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", node.Name.String()), node.Name)
		return false, false
	}
	itNode := node.ChildNodes[0]
	switch node := itNode.(type) {
	case *ast.Token:
		t := node.Token
		if t.Kind != token.KeywordTrue && t.Kind != token.KeywordFalse {
			p.addErrorToken(fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", t.String()), t)
			return false, false
		}
		valueString := node.String()
		var value bool
		var ok bool
		if valueString == "true" {
			value = true
			ok = true
		}
		if !ok && valueString == "false" {
			value = false
			ok = true
		}
		if !ok {
			p.addErrorToken(fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", t.String()), t)
			return false, false
		}
		return value, ok
	}
	p.fatalErrorToken(fmt.Errorf("Unknown type %T", node), node.Name)
	return false, false
}
