package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/errors"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/util"
)

type Parser struct {
	scanner.Scanner
	errors.ErrorHandler

	// Used and reset per-file parsed
	dependencies map[string]bool
}

func New() *Parser {
	p := new(Parser)

	p.ErrorHandler.Init()
	p.ErrorHandler.SetDeveloperMode(true)
	return p
}

func (p *Parser) HasErrors() bool {
	return p.Scanner.HasErrors() || p.ErrorHandler.HasErrors()
}

func (p *Parser) ParseFile(filepath string) (*ast.File, error) {
	filecontentsAsBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	result := p.Parse(filecontentsAsBytes, filepath)
	if result == nil {
		return nil, fmt.Errorf("Parse error.")
	}
	return result, nil
}

func (p *Parser) Parse(filecontentsAsBytes []byte, filepath string) *ast.File {
	p.Scanner.Init(filecontentsAsBytes, filepath)
	t := p.PeekNextToken()
	if t.Kind == token.EOF {
		p.AddError(t, fmt.Errorf("Empty source file: %s", filepath))
		return nil
	}

	//
	p.dependencies = make(map[string]bool)
	astFile := &ast.File{
		Filepath: filepath,
	}
	astFile.ChildNodes = p.parseStatements()
	astFile.Dependencies = p.dependencies
	json, _ := json.MarshalIndent(astFile.Dependencies, "", "   ")
	fmt.Printf("%s\nJSON AST\n---------------\n", string(json))
	p.dependencies = nil

	return astFile
}

func (p *Parser) validateHTMLNode(node *ast.Call) {
	name := node.Name.String()
	if len(node.ChildNodes) > 0 && util.IsSelfClosingTagName(name) {
		p.AddError(node.Name, fmt.Errorf("%s is a self-closing tag and cannot have child elements.", name))
	}
	p.dependencies[name] = true
	// todo(Jake): Extend this to allow user configured/whitelisted tag names
	//
	//isValidHTML5TagName := util.IsValidHTML5TagName(name)
	//if !isValidHTML5TagName {
	//	p.AddError(node.Name, fmt.Errorf("\"%s\" is not a valid HTML5 element.", name))
	//}
}

func (p *Parser) NewDeclareStatement(name token.Token, typeIdent ast.TypeIdent, expressionNodes []ast.Node) *ast.DeclareStatement {
	node := new(ast.DeclareStatement)
	node.Name = name
	node.TypeIdentifier = typeIdent
	node.ChildNodes = expressionNodes

	nameString := name.String()
	if len(nameString) > 0 && nameString[len(nameString)-1] == '-' {
		p.AddError(name, fmt.Errorf("Declaring variable name ending with - is illegal."))
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
				node := p.NewDeclareStatement(name, ast.TypeIdent{}, p.parseExpressionNodes(false))
				resultNodes = append(resultNodes, node)
			// myVar : string \n
			case token.Colon:
				typeName := p.parseTypeIdent()
				if typeName.Name.Kind == token.Unknown {
					return nil
				}
				// myVar : string = {Expression} \n
				var expressionNodes []ast.Node
				if p.PeekNextToken().Kind == token.Equal {
					p.GetNextToken()
					expressionNodes = p.parseExpressionNodes(false)
				}
				node := p.NewDeclareStatement(name, typeName, expressionNodes)
				resultNodes = append(resultNodes, node)
			// myVar []= "append array item"
			case token.BracketOpen:
				if t := p.GetNextToken(); t.Kind != token.BracketClose {
					p.AddExpectError(t, token.BracketClose)
					continue
				}
				if t := p.GetNextToken(); t.Kind != token.Equal {
					p.AddExpectError(t, token.Equal)
					continue
				}

				leftHandSide := make([]token.Token, 0, 1)
				leftHandSide = append(leftHandSide, name)

				node := new(ast.ArrayAppendStatement)
				node.LeftHandSide = leftHandSide
				node.Expression.ChildNodes = p.parseExpressionNodes(false)
				resultNodes = append(resultNodes, node)
			// myVar.Property.SubProperty = {Expression}
			//
			case token.Dot:
				leftHandSide := make([]token.Token, 0, 5)
				leftHandSide = append(leftHandSide, name)
				for {
					t := p.GetNextToken()
					if t.Kind != token.Identifier {
						p.AddExpectError(t, token.Identifier)
						return nil
					}
					leftHandSide = append(leftHandSide, t)
					if dotToken := p.PeekNextToken(); dotToken.Kind == token.Dot {
						p.GetNextToken()
						continue
					}
					break
				}

				operatorToken := p.GetNextToken()
				if operatorToken.Kind == token.BracketOpen {
					if t := p.GetNextToken(); t.Kind != token.BracketClose {
						p.AddExpectError(t, token.BracketClose)
						continue
					}
					if t := p.GetNextToken(); t.Kind != token.Equal {
						p.AddExpectError(t, token.Equal)
						continue
					}

					node := new(ast.ArrayAppendStatement)
					node.LeftHandSide = leftHandSide
					node.Expression.ChildNodes = p.parseExpressionNodes(false)
					resultNodes = append(resultNodes, node)
					continue
				}
				if operatorToken.Kind == token.DeclareSet {
					p.AddError(operatorToken, fmt.Errorf("Cannot use := on a property. (%s)", ast.LeftHandSide(leftHandSide)))
					continue
				}

				if operatorToken.Kind != token.Equal &&
					operatorToken.Kind != token.AddEqual {
					p.AddExpectError(operatorToken, operatorToken, token.Equal, token.AddEqual)
					continue
				}

				node := new(ast.OpStatement)
				node.LeftHandSide = leftHandSide
				node.Operator = operatorToken
				node.Expression.ChildNodes = p.parseExpressionNodes(false)
				resultNodes = append(resultNodes, node)
			// myVar = {Expression} \n
			//
			case token.Equal, token.AddEqual:
				leftHandSide := make([]token.Token, 1)
				leftHandSide[0] = name
				node := new(ast.OpStatement)
				node.LeftHandSide = leftHandSide
				node.Operator = t
				node.Expression.ChildNodes = p.parseExpressionNodes(false)
				resultNodes = append(resultNodes, node)
			// div if {expr} {
			//     ^
			case token.KeywordIf:
				node := ast.NewHTMLNode()
				node.Name = name
				node.IfExpression.ChildNodes = p.parseExpressionNodes(true)
				if t := p.GetNextToken(); t.Kind != token.BraceOpen {
					p.AddExpectError(t, token.BraceOpen)
					return nil
				}
				// NOTE(Jake): 2018-01-14
				//
				// This code seems like a mistake, but perhaps its solving
				// a quirk I've forgotten about. To be deleted if its not
				// needed.
				//
				//if t := p.GetNextToken(); t.Kind == token.BraceOpen {
				//	p.AddExpectError(t, token.BraceOpen)
				//	return nil
				//}
				p.GetNextToken()
				node.ChildNodes = p.parseStatements()
				p.validateHTMLNode(node)
				resultNodes = append(resultNodes, node)
			// div {
			//     ^
			case token.BraceOpen:
				node := ast.NewHTMLNode()
				node.Name = name
				node.ChildNodes = p.parseStatements()
				p.validateHTMLNode(node)
				resultNodes = append(resultNodes, node)
			// div(class="hey")  -or-  div(class="hey") if expr {
			//    ^						  ^
			case token.ParenOpen:
				node := p.parseProcedureOrHTMLNode(name)
				if node == nil {
					return nil
				}
				resultNodes = append(resultNodes, node)
			// PrintThisVariable \n
			// ^
			case token.Newline:
				if name.String() == "return" {
					p.SetScannerState(storeScannerState)
					node := new(ast.Return)
					node.TypeIdentifier.Name = p.GetNextToken() // consume `return`
					// NOTE(Jake): 2017-12-30, Hack to store Line/Column/File data from token on ast.Return
					node.TypeIdentifier.Name.Kind = token.Unknown
					node.TypeIdentifier.Name.Data = ""
					//node.Expression.ChildNodes = nil
					resultNodes = append(resultNodes, node)
					continue
				}
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
				// return {expr}
				if name.Kind == token.Identifier &&
					name.String() == "return" {
					p.SetScannerState(storeScannerState)
					node := new(ast.Return)
					node.TypeIdentifier.Name = p.GetNextToken() // consume `return`
					// NOTE(Jake): 2017-12-30, Hack to store Line/Column/File data from token on ast.Return
					node.TypeIdentifier.Name.Kind = token.Unknown
					node.TypeIdentifier.Name.Data = ""
					node.Expression.ChildNodes = p.parseExpressionNodes(false)
					resultNodes = append(resultNodes, node)
					continue
				}
				p.AddError(t, fmt.Errorf("Unexpected %s (%s) after identifier (%s).", t.Kind.String(), t.String(), name.String()))
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
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			node := new(ast.If)
			node.Condition.ChildNodes = exprNodes
			node.ChildNodes = p.parseStatements()
			resultNodes = append(resultNodes, node)
			p.eatNewlines()
			// Eat newlines
			/*for {
				t := p.PeekNextToken()
				if t.Kind == token.Newline {
					p.GetNextToken()
					continue
				}
				break
			}*/
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
					p.AddExpectError(t, token.BraceOpen, token.KeywordIf)
					return nil
				}
				node.ElseNodes = p.parseStatements()
			}
		case token.KeywordFor:
			p.GetNextToken()
			varName := p.GetNextToken()
			if varName.Kind != token.Identifier {
				p.AddExpectError(varName, token.Identifier)
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
					p.AddExpectError(secondVarName, token.Identifier)
					return nil
				}
				t = p.GetNextToken()
				if t.Kind != token.DeclareSet {
					p.AddExpectError(t, token.DeclareSet)
					return nil
				}
				node = new(ast.For)
				node.IsDeclareSet = true
				node.IndexName = varName
				node.RecordName = secondVarName
				node.Array.ChildNodes = p.parseExpressionNodes(false)
			default:
				p.AddExpectError(t, token.DeclareSet, token.Comma)
				return nil
			}
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			node.ChildNodes = p.parseStatements()
			resultNodes = append(resultNodes, node)
		case token.BraceOpen:
			p.GetNextToken()
			node := new(ast.Block)
			node.ChildNodes = p.parseStatements()
			resultNodes = append(resultNodes, node)
		case token.EOF, token.Illegal:
			break Loop
		default:
			p.GetNextToken()
			p.AddUnexpectedErrorWithContext(t, "statements")
			return nil
		}
	}
	return resultNodes
}

func (p *Parser) isParseTypeAhead() bool {
	if t := p.PeekNextToken(); t.Kind != token.BracketOpen && t.Kind != token.Identifier {
		return false
	}
	return true
}

func (p *Parser) parseTypeIdent() ast.TypeIdent {
	result := ast.TypeIdent{}

	t := p.GetNextToken()
	if t.Kind == token.BracketOpen {
		// Parse array / array-of-array / etc
		// ie. []string, [][]string, [][][]string, etc
		result.ArrayDepth = 1
		for {
			t = p.GetNextToken()
			if t.Kind != token.BracketClose {
				p.AddExpectError(t, token.BracketClose)
				return ast.TypeIdent{}
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
		p.AddExpectError(t, "type identifier")
		return ast.TypeIdent{}
	}
	result.Name = t
	return result
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
				p.AddError(name, fmt.Errorf("Expected operator instead got identifier (%s).", name.String()))
				return nil
			}
			switch t := p.PeekNextToken(); t.Kind {
			case token.Dot:
				p.GetNextToken()
				tokens := make([]token.Token, 0, 5)
				tokens = append(tokens, name)
				for {
					identToken := p.GetNextToken()
					tokens = append(tokens, identToken)
					if identToken.Kind != token.Identifier {
						p.AddExpectError(identToken, token.Identifier)
						return nil
					}
					t := p.PeekNextToken()
					if t.Kind == token.Dot {
						p.GetNextToken()
						continue
					}
					if t.IsOperator() ||
						t.Kind == token.Comma ||
						t.Kind == token.Newline ||
						t.Kind == token.ParenClose {
						break
					}
					p.AddExpectError(t, token.Operator, token.Newline, token.ParenClose)
					return nil
				}
				expectOperator = true
				infixNodes = append(infixNodes, ast.NewTokenList(tokens))
				continue Loop
			case token.ParenOpen:
				p.GetNextToken()
				node := p.parseProcedureOrHTMLNode(name)
				if node == nil {
					return nil
				}
				expectOperator = true
				infixNodes = append(infixNodes, node)
				continue Loop
			case token.BraceOpen:
				if disableStructLiteral {
					// Dont parse struct literal and use identifier as-is
					// (ie. "if isBool {")
					break
				}
				p.GetNextToken()

				{
					p.eatNewlines()
					if t := p.PeekNextToken(); t.Kind == token.BraceClose {
						p.GetNextToken()
						expectOperator = true
						infixNodes = append(infixNodes, &ast.StructLiteral{
							Name: name,
						})
						continue Loop
					}
				}

				var errorMsgLastToken token.Token
				fields := make([]ast.Parameter, 0, 10)
			StructLiteralLoop:
				for i := 0; true; i++ {
					propertyName := p.GetNextToken()
					for propertyName.Kind == token.Newline {
						propertyName = p.GetNextToken()
					}
					if propertyName.Kind != token.Identifier {
						if i == 0 {
							p.AddError(propertyName, fmt.Errorf("Expected identifier after %s{ not %s", name, propertyName.Kind.String()))
							return nil
						}
						p.AddError(propertyName, fmt.Errorf("Expected identifier after \"%s\" not %s", errorMsgLastToken, propertyName.Kind.String()))
						return nil
					}
					if t := p.GetNextToken(); t.Kind != token.Colon {
						if i == 0 {
							p.AddError(t, fmt.Errorf("Expected : after \"%s{%s\"", name.String(), propertyName.String()))
							return nil
						}
						p.AddError(t, fmt.Errorf("Expected : after property \"%s\"", propertyName.String()))
						return nil
					}
					exprNodes := p.parseExpressionNodes(false)
					node := ast.Parameter{}
					node.Name = propertyName
					node.ChildNodes = exprNodes
					fields = append(fields, node)
					switch t := p.GetNextToken(); t.Kind {
					case token.BraceClose:
						break StructLiteralLoop
					case token.Comma:
						// OK
					default:
						p.AddExpectError(t, token.BraceClose, token.Comma)
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
				expectOperator = true
				infixNodes = append(infixNodes, &ast.StructLiteral{
					Name:   name,
					Fields: fields,
				})
				continue Loop
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.KeywordTrue, token.KeywordFalse:
			p.GetNextToken()
			if expectOperator {
				p.AddError(t, fmt.Errorf("Expected operator, instead got true/false keyword (\"%s\").", t.String()))
				return nil
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.String:
			p.GetNextToken()
			if expectOperator {
				p.AddError(t, fmt.Errorf("Expected operator, instead got string (\"%s\").", t.String()))
				return nil
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
				p.AddError(t, fmt.Errorf("Expected operator, instead got number (\"%s\").", t.String()))
				return nil
			}
			expectOperator = true
			infixNodes = append(infixNodes, &ast.Token{Token: t})
		case token.ParenOpen:
			parenOpenCount++
		case token.ParenClose:
			// If hit end of parameter list
			// ie. `div(prop=param1, prop2=param2)` or `functionCall(param1, param2)`
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
				p.PanicError(t, fmt.Errorf("parseExpressionNodes: parseDefinition returned nil."))
			}
			infixNodes = append(infixNodes, node)
		// ie. []string{"item1", "item2", "item3"}
		case token.BracketOpen:
			typeIdent := p.parseTypeIdent()
			if typeIdent.Name.Kind == token.Unknown {
				return nil
			}
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			p.eatNewlines()

			childNodes := make([]ast.Node, 0, 10)

		ArrayLiteralLoop:
			for i := 0; true; i++ {
				expr := p.parseExpression(false)
				sep := p.GetNextToken()
				switch sep.Kind {
				case token.Comma:
					childNodes = append(childNodes, expr)
					// Handle trailing comma
					p.eatNewlines()
					if p.PeekNextToken().Kind == token.BraceClose {
						p.GetNextToken()
						break ArrayLiteralLoop
					}
					continue
				case token.BraceClose:
					childNodes = append(childNodes, expr)
					break ArrayLiteralLoop
				case token.EOF:
					p.AddUnexpectedErrorWithContext(sep, "array literal")
					return nil
				}
				p.AddError(sep, fmt.Errorf("Expected , or } after array item #%d, not %s.", i, sep.Kind.String()))
				return nil
			}

			if len(childNodes) == 0 {
				p.AddError(typeIdent.Name, fmt.Errorf("Cannot have array literal with zero elements."))
			}

			node := new(ast.ArrayLiteral)
			node.TypeIdentifier = typeIdent
			node.ChildNodes = childNodes

			infixNodes = append(infixNodes, node)
			continue Loop
		default:
			if t.IsOperator() {
				if !expectOperator {
					p.AddError(t, fmt.Errorf("Expected identifiers or string, instead got operator \"%s\".", t.String()))
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
			p.PanicError(t, fmt.Errorf("Unhandled token type: \"%s\" (value: %s)", t.Kind.String(), t.String()))
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

func (p *Parser) parseProcedureOrHTMLNode(name token.Token) *ast.Call {
	hasDeterminedMode := false
	isHTMLNode := false
	parameters := make([]*ast.Parameter, 0, 10)

	// Eat all newlines after (
	p.eatNewlines()

	if p.PeekNextToken().Kind == token.ParenClose {
		p.GetNextToken() // no parameters
	} else {
	CallLoop:
		for {
			// NOTE(Jake): 2018-01-03
			//
			// Eating the surrounding newlines so we can do the following:
			//
			// `blah(
			//	  "param1"
			//	  ,
			//    "param2"
			// )`
			//
			p.eatNewlines()

			storeScannerState := p.ScannerState()
			name := p.GetNextToken()
			equalOp := p.GetNextToken()
			if name.Kind == token.Identifier &&
				equalOp.Kind == token.Equal {
				if hasDeterminedMode && !isHTMLNode {
					p.AddError(name, fmt.Errorf("Cannot use named parameter after unnamed parameter, parameter #%d.", len(parameters)))
				}
				isHTMLNode = true
				hasDeterminedMode = true
			} else {
				if hasDeterminedMode && isHTMLNode {
					p.AddError(name, fmt.Errorf("Cannot use unnamed parameter after named parameter, parameter #%d.", len(parameters)))
				}
				hasDeterminedMode = true
			}
			if !isHTMLNode {
				p.SetScannerState(storeScannerState)
			}

			exprNodes := p.parseExpressionNodes(false)
			p.eatNewlines()

			if exprNodes == nil {
				p.AddError(name, fmt.Errorf("Missing value for parameter #%d", len(parameters)))
				return nil
			}
			parameter := new(ast.Parameter)
			if isHTMLNode {
				parameter.Name = name
			}
			parameter.ChildNodes = exprNodes
			parameters = append(parameters, parameter)

			switch t := p.PeekNextToken(); t.Kind {
			case token.Newline:
				continue CallLoop
			case token.Comma:
				p.GetNextToken()

				// NOTE(Jake): 2018-01-03
				//
				// Needed to allow for trailing commas
				//
				p.eatNewlines()
				switch t := p.PeekNextToken(); t.Kind {
				case token.ParenClose:
					p.GetNextToken()
					break CallLoop
				case token.Comma:
					p.AddError(t, fmt.Errorf("Cannot have more than 1 trailing comma for procedure calls."))
					return nil
				}
			case token.ParenClose:
				p.GetNextToken()
				break CallLoop
			default:
				p.AddExpectError(t, token.Comma, token.ParenClose)
				return nil
			}
		}
	}

	childStatements := make([]ast.Node, 0, 10)
	ifExprNodes := make([]ast.Node, 0, 10)

	{
		storeScannerState := p.ScannerState()
		switch t := p.GetNextToken(); t.Kind {
		case token.Newline:
			// no-op
		case token.BraceOpen:
			childStatements = p.parseStatements()
			isHTMLNode = true
		case token.KeywordIf:
			ifExprNodes = p.parseExpressionNodes(true)
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			childStatements = p.parseStatements()
			isHTMLNode = true
		default:
			if t.IsOperator() {
				p.SetScannerState(storeScannerState)
				break
			}
			p.AddExpectError(t, token.BraceOpen, token.KeywordIf, token.Newline)
		}
	}

	// todo(Jake): Extend this to allow user configured/whitelisted tag names
	//isValidHTML5TagName := util.IsValidHTML5TagName(name.String())
	//if isValidHTML5TagName {
	//	isHTMLNode = true
	//}

	if !isHTMLNode {
		node := ast.NewCall()
		node.Name = name
		node.Parameters = parameters
		p.dependencies[name.String()] = true
		return node
	}
	node := ast.NewHTMLNode()
	node.Name = name
	node.Parameters = parameters
	node.ChildNodes = childStatements
	node.IfExpression.ChildNodes = ifExprNodes
	p.validateHTMLNode(node)
	return node
}

func (p *Parser) parseProcedureDefinition(name token.Token) *ast.ProcedureDefinition {
	var parameters []ast.Parameter
	if hasNoParameters := p.PeekNextToken().Kind == token.ParenClose; hasNoParameters {
		p.GetNextToken()
	} else {
		parameters = make([]ast.Parameter, 0, 4)
		for {
			name := p.GetNextToken()
			if name.Kind != token.Identifier {
				p.AddExpectError(name, token.ParenClose, token.Identifier)
				return nil
			}
			typeIdent := p.parseTypeIdent()
			if typeIdent.Name.Kind == token.Unknown {
				return nil
			}

			parameter := ast.Parameter{
				Name: name,
			}
			parameter.TypeIdentifier = typeIdent
			parameters = append(parameters, parameter)

			t := p.GetNextToken()
			if t.Kind == token.Comma {
				continue
			}
			if t.Kind == token.ParenClose {
				break
			}
			p.AddExpectError(t, token.Comma, token.ParenClose)
			return nil
		}
	}
	var returnTypeIdent ast.TypeIdent
	if p.isParseTypeAhead() {
		returnTypeIdent = p.parseTypeIdent()
		if returnTypeIdent.Name.Kind == token.Unknown {
			return nil
		}
	}
	if t := p.GetNextToken(); t.Kind != token.BraceOpen {
		p.AddExpectError(t, token.BraceOpen)
		return nil
	}
	node := new(ast.ProcedureDefinition)
	node.Name = name
	node.Parameters = parameters
	node.TypeIdentifier = returnTypeIdent
	node.ChildNodes = p.parseStatements()
	return node
}

func (p *Parser) parseDefinition(name token.Token) ast.Node {
	keywordToken := p.GetNextToken()
	keyword := keywordToken.String()
	switch keywordToken.Kind {
	case token.ParenOpen:
		node := p.parseProcedureDefinition(name)
		if node == nil {
			return nil
		}
		return node
	case token.Identifier:
		switch keyword {
		case "workspace":
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			childNodes := p.parseStatements()
			node := new(ast.WorkspaceDefinition)
			node.Name = name
			node.ChildNodes = childNodes
			return node
		case "css":
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			node := p.parseCSS(name)
			return node
		case "css_config":
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			node := p.parseCSSConfigRuleDefinition(name)
			return node
		case "struct":
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			//
			//
			childNodes := p.parseStatements()
			fields := make([]ast.StructField, 0, len(childNodes))
			fieldIndex := 0
			// NOTE(Jake): A bit of a hack, we should have a 'parseStruct' function
			for _, itNode := range childNodes {
				switch node := itNode.(type) {
				case *ast.DeclareStatement:
					field := ast.StructField{}
					field.Name = node.Name
					field.Index = fieldIndex
					fieldIndex++
					//field.Expression.TypeIdentifier = node.Expression.TypeIdentifier
					field.Expression = node.Expression
					fields = append(fields, field)
				default:
					p.AddError(name, fmt.Errorf("Expected statement, instead got %T.", itNode))
					return nil
				}
			}
			node := new(ast.StructDefinition)
			node.Name = name
			node.Fields = fields
			return node
		case "html":
			if t := p.GetNextToken(); t.Kind != token.BraceOpen {
				p.AddExpectError(t, token.BraceOpen)
				return nil
			}
			childNodes := p.parseStatements()

			// Check HTML nodes
			htmlNodeCount := 0
			for _, itNode := range childNodes {
				node, ok := itNode.(*ast.Call)
				if !ok || node.Kind() != ast.CallHTMLNode {
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
					p.AddError(name, fmt.Errorf("\"%s:: html\" must contain one HTML node at the top-level.", nameString))
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
							p.AddError(node.Name, fmt.Errorf("Cannot declare \":: struct\" twice in the same HTML component."))
							p.AddError(structDef.Name, fmt.Errorf("Cannot declare \":: struct\" twice in the same HTML component."))
							break RetrievePropertyDefinitionLoop
						}
						structDef = node
					case *ast.CSSDefinition:
						if cssDef != nil {
							p.AddError(node.Name, fmt.Errorf("Cannot declare \":: css\" twice in the same HTML component."))
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
	}
	p.AddError(keywordToken, fmt.Errorf("Unexpected keyword '%s' for definition (::) type. Expected 'css', 'html', 'struct', 'workspace' or () on Line %d", keyword, keywordToken.Line))
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
						p.AddError(node.Name, fmt.Errorf("Invalid config key \"%s\". Expected \"modify\".", name))
						return nil
					}
				case *ast.DeclareStatement:
					p.AddError(node.Name, fmt.Errorf("Cannot declare variables in a css_config block. Did you mean to use : instead of :="))
					return nil
				default:
					panic(fmt.Sprintf("parseCSSConfigRuleDefinition:propertyLoop: Unknown type %T", node))
				}
			}

			// Get matching parts
			for _, selector := range node.Selectors() {
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
								p.AddError(selectorPartNode.Token, fmt.Errorf("Only supports * wildcard, not %s", operator))
								return nil
							}
							continue
						}
						if selectorPartNode.Kind != token.Identifier {
							p.AddError(selectorPartNode.Token, fmt.Errorf("Expected identifier, instead got %s", selectorPartNode.Kind.String()))
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
			p.AddError(node.Name, fmt.Errorf("Cannot declare variables in a css_config block."))
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
		p.AddError(node.Name, fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", node.Name.String()))
		return false, false
	}
	itNode := node.ChildNodes[0]
	switch node := itNode.(type) {
	case *ast.Token:
		t := node.Token
		if t.Kind != token.KeywordTrue && t.Kind != token.KeywordFalse {
			p.AddError(t, fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", t.String()))
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
			p.AddError(t, fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", t.String()))
			return false, false
		}
		return value, ok
	}
	p.PanicError(node.Name, fmt.Errorf("Unknown type %T", node))
	return false, false
}
