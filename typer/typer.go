package typer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/errors"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
	"github.com/silbinarywolf/compiler-fel/util"
)

type Typer struct {
	errors.ErrorHandler
	typeinfo                      TypeInfoManager
	typecheckHtmlNodeDependencies map[string]*ast.Call
	htmlComponentsUsed            []*ast.HTMLComponentDefinition // track used components for emitting css
}

func New() *Typer {
	p := new(Typer)

	p.ErrorHandler.Init()
	p.ErrorHandler.SetDeveloperMode(true)

	p.typeinfo.Init()

	p.htmlComponentsUsed = make([]*ast.HTMLComponentDefinition, 0, 10)

	// NOTE(Jake): 2018-04-09
	//
	// Keeping this value unset is intentional.
	// ... I think.
	//
	//p.typecheckHtmlNodeDependencies = nil

	return p
}

func (p *Typer) HTMLComponentsUsed() []*ast.HTMLComponentDefinition { return p.htmlComponentsUsed }

func (p *Typer) typerStructLiteral(scope *Scope, literal *ast.StructLiteral) {
	name := literal.Name.String()
	symbol := scope.GetSymbol(name)
	if symbol == nil {
		p.AddError(literal.Name, fmt.Errorf("Undeclared \"%s :: struct\"", name))
		return
	}
	def := symbol.structDefinition
	if def == nil {
		p.AddError(literal.Name, fmt.Errorf("Expected \"%s\" to be \"%s :: struct\", not \"%s\".", name, name, symbol.GetType()))
		return
	}
	if len(def.Fields) == 0 && len(literal.Fields) > 0 {
		p.AddError(literal.Name, fmt.Errorf("Struct %s does not have any fields.", name))
		return
	}
	literal.TypeInfo = p.typeinfo.getByName(name)
	if literal.TypeInfo == nil {
		p.PanicError(literal.Name, fmt.Errorf("Missing type info for \"%s :: struct\".", name))
		return
	}

	// Check literal against definition
	for _, defField := range def.Fields {
		defTypeInfo := defField.Expression.TypeInfo
		if defTypeInfo == nil {
			p.PanicError(defField.Name, fmt.Errorf("Missing type info from field \"%s\" on \"%s :: struct\".", defField.Name.String(), name))
			return
		}
		defName := defField.Name.String()
		for i := 0; i < len(literal.Fields); i++ {
			property := &literal.Fields[i]
			if defName != property.Name.String() {
				continue
			}
			p.typerExpression(scope, &property.Expression)
			litTypeInfo := property.Expression.TypeInfo
			if litTypeInfo != defTypeInfo {
				p.AddError(property.Name, fmt.Errorf("Mismatching type, expected \"%s\" but got \"%s\"", defField.TypeIdentifier.Name.String(), property.Expression.TypeInfo.String()))
			}
		}
	}

	for _, property := range literal.Fields {
		propertyName := property.Name.String()
		hasFieldNameOnDef := false
		for _, defField := range def.Fields {
			hasFieldNameOnDef = hasFieldNameOnDef || defField.Name.String() == propertyName
		}
		if !hasFieldNameOnDef {
			p.AddError(property.Name, fmt.Errorf("Field \"%s\" does not exist on \"%s :: struct\"", propertyName, name))
		}
	}
}

func (p *Typer) typerArrayLiteral(scope *Scope, literal *ast.ArrayLiteral) {
	//test := [][]string{
	//	[]string{"test"}
	//}
	//if len(test) > 0 {
	//
	//}

	typeIdentName := literal.TypeIdentifier.Name
	typeIdentString := typeIdentName.String()
	typeInfo := p.DetermineType(&literal.TypeIdentifier)
	if typeInfo == nil {
		p.AddError(typeIdentName, fmt.Errorf("Undeclared type \"%s\" used for array literal", typeIdentString))
		return
	}
	literal.TypeInfo = typeInfo

	//
	resultTypeInfo, ok := typeInfo.(*types.Array)
	if !ok {
		p.PanicError(typeIdentName, fmt.Errorf("Expected array type but got \"%s\".", typeIdentString))
		return
	}
	underlyingTypeInfo := resultTypeInfo.Underlying()

	// Run type checking on each array element
	nodes := literal.Nodes()
	for i, node := range nodes {
		node := node.(*ast.Expression)
		// NOTE(Jake): Set to 'string' type info so
		//			   type checking will catch things immediately
		//			   when we call `typerExpression`
		//			   ie. Won't infer, will mark as invalid.
		if node.TypeInfo == nil {
			node.TypeInfo = underlyingTypeInfo
		}
		p.typerExpression(scope, node)

		if node.TypeInfo == nil {
			panic(fmt.Sprintf("typerArrayLiteral: Missing type on array literal item #%d.", i))
		}
	}
}

func (p *Typer) typerCall(scope *Scope, node *ast.Call) {
	switch node.Kind() {
	case ast.CallProcedure:
		p.typerProcedureCall(scope, node)
	case ast.CallHTMLNode:
		p.typerHTMLNode(scope, node)
	default:
		p.PanicError(node.Name, fmt.Errorf("Unhandled ast.Call kind: %s", node.Name))
	}
}

func (p *Typer) typerProcedureCall(scope *Scope, node *ast.Call) {
	typeInfo := p.typeinfo.getByName(node.Name.String())
	callTypeInfo, ok := typeInfo.(*types.Procedure)
	if !ok {
		// todo(Jake): 2018-01-14
		//
		//
		//
		if typeInfo == nil {
			p.AddError(node.Name, fmt.Errorf("Procedure \"%s()\" is not defined.", node.Name.String()))
			return
		}
		p.PanicError(node.Name, fmt.Errorf("Expected %s to be a procedure, instead got %T.", node.Name.String(), typeInfo))
		return
	}
	procDefinition := callTypeInfo.Definition()
	node.Definition = procDefinition

	parameters := node.Parameters
	definitionParameters := procDefinition.Parameters
	hasMismatchingTypes := len(definitionParameters) != len(parameters)
	for i := 0; i < len(parameters); i++ {
		parameter := parameters[i]
		p.typerExpression(scope, &parameter.Expression)
		if hasMismatchingTypes == false && i < len(definitionParameters) {
			definitionParameter := definitionParameters[i]
			if !TypeEquals(parameter.TypeInfo, definitionParameter.TypeInfo) {
				hasMismatchingTypes = true
			}
		}
	}
	if hasMismatchingTypes {
		haveStr := "("
		for i := 0; i < len(parameters); i++ {
			parameter := parameters[i]
			if i != 0 {
				haveStr += ", "
			}
			if parameter.TypeInfo == nil {
				// NOTE(Jake): 2018-01-03
				//
				// This should be applied in p.typerExpression(scope, &parameter.Expression)
				//
				haveStr += "missing"
				continue
			}
			haveStr += parameter.TypeInfo.String()
		}
		haveStr += ")"
		wantStr := "("
		for i := 0; i < len(procDefinition.Parameters); i++ {
			parameter := procDefinition.Parameters[i]
			if i != 0 {
				wantStr += ", "
			}
			if parameter.TypeInfo == nil {
				// NOTE(Jake): 2018-01-03
				//
				// This should be applied in p.typerExpression(scope, &parameter.Expression)
				//
				haveStr += "missing"
				continue
			}
			wantStr += parameter.TypeInfo.String()
		}
		wantStr += ")"
		callStr := node.Name.String()
		if len(definitionParameters) != len(parameters) {
			p.AddError(node.Name, fmt.Errorf("Expected %d parameters, instead got %d parameters on call \"%s\".\nhave %s\nwant %s", len(definitionParameters), len(parameters), callStr, haveStr, wantStr))
		} else {
			p.AddError(node.Name, fmt.Errorf("Mismatching types on call \"%s\".\nhave %s\nwant %s", callStr, haveStr, wantStr))
		}
	}
}

func (p *Typer) typerExpression(scope *Scope, expression *ast.Expression) {
	resultTypeInfo := expression.TypeInfo

	// Get type info from text (ie. "string", "int", etc)
	if typeIdent := expression.TypeIdentifier.Name; resultTypeInfo == nil && typeIdent.Kind != token.Unknown {
		typeIdentString := typeIdent.String()
		resultTypeInfo = p.DetermineType(&expression.TypeIdentifier)
		if resultTypeInfo == nil {
			p.AddError(typeIdent, fmt.Errorf("Undeclared type %s", typeIdentString))
			return
		}
	}

	var leftToken token.Token
	nodes := expression.Nodes()
	for i, itNode := range nodes {
		switch node := itNode.(type) {
		case *ast.StructLiteral:
			p.typerStructLiteral(scope, node)
			expectedTypeInfo := node.TypeInfo
			if expectedTypeInfo == nil {
				p.PanicError(node.Name, fmt.Errorf("Missing type info for \"%s :: struct\".", node.Name.String()))
				continue
			}
			if resultTypeInfo == nil {
				resultTypeInfo = expectedTypeInfo
			}
			if !TypeEquals(resultTypeInfo, expectedTypeInfo) {
				p.AddError(node.Name, fmt.Errorf("Cannot mix struct literal %s with %s", expectedTypeInfo.String(), resultTypeInfo.String()))
			}
			continue
		case *ast.ArrayLiteral:
			p.typerArrayLiteral(scope, node)
			expectedTypeInfo := node.TypeInfo
			if resultTypeInfo == nil {
				resultTypeInfo = expectedTypeInfo
			}
			if !TypeEquals(resultTypeInfo, expectedTypeInfo) {
				p.AddError(node.TypeIdentifier.Name, fmt.Errorf("Cannot mix array literal %s with %s", expectedTypeInfo.String(), resultTypeInfo.String()))
			}
			continue
		case *ast.Call:
			p.typerCall(scope, node)
			switch node.Kind() {
			case ast.CallProcedure:
				if node.Definition != nil {
					expectedTypeInfo := node.Definition.TypeInfo
					if resultTypeInfo == nil {
						resultTypeInfo = expectedTypeInfo
					}
					if !TypeEquals(resultTypeInfo, expectedTypeInfo) {
						p.AddError(node.Name, fmt.Errorf("Cannot mix call type %s with %s", expectedTypeInfo.String(), resultTypeInfo.String()))
					}
				}
			case ast.CallHTMLNode:
				p.AddError(node.Name, fmt.Errorf("Cannot use HTML node in expression."))
			default:
				panic(fmt.Sprintf("typerExpression:Call: Unhandled call kind: %s", node.Kind()))
			}

			continue
		case *ast.HTMLBlock:
			panic("typerExpression: todo(Jake): Fix HTMLBlock")
			/*variableType := data.KindHTMLNode
			if exprType == data.KindUnknown {
				exprType = variableType
			}
			if exprType != variableType {
				p.addErrorToken(fmt.Errorf("\":: html\" must be a %s not %s.", exprType.String(), variableType.String()), node.HTMLKeyword)
			}
			p.typerHTMLBlock(node, scope)*/
		case *ast.TokenList:
			tokens := node.Tokens()
			expectedTypeInfo := p.getTypeFromLeftHandSide(tokens, scope)
			if expectedTypeInfo == nil {
				continue
			}
			if resultTypeInfo == nil {
				resultTypeInfo = expectedTypeInfo
			}
			if !TypeEquals(resultTypeInfo, expectedTypeInfo) {
				p.AddError(tokens[0], fmt.Errorf("Cannot mix variable \"%s\" type %s with %s", node.String(), expectedTypeInfo.String(), resultTypeInfo.String()))
			}
			continue
		case *ast.Token:
			if node.IsOperator() {
				continue
			}

			switch node.Kind {
			case token.Identifier:
				name := node.String()
				symbol := scope.GetSymbol(name)
				if symbol == nil {
					p.AddError(node.Token, fmt.Errorf("Undeclared identifier \"%s\".", name))
					continue
				}
				variableTypeInfo := symbol.variable
				if variableTypeInfo == nil {
					if htmlComponentDefinition := symbol.htmlDefinition; htmlComponentDefinition != nil {
						p.AddError(node.Token, fmt.Errorf("Undeclared identifier \"%s\". Did you mean \"%s()\" or \"%s{ }\" to reference the \"%s :: html\" component?", name, name, name, name))
						continue
					}
					p.AddError(node.Token, fmt.Errorf("Identifier \"%s\" is not a variable", name))
					continue
				}
				if resultTypeInfo == nil {
					resultTypeInfo = variableTypeInfo
				}
				if !TypeEquals(resultTypeInfo, variableTypeInfo) {
					if variableTypeInfo == nil {
						p.PanicError(node.Token, fmt.Errorf("Unable to determine type, typerer must have failed."))
						return
					}
					p.AddError(node.Token, fmt.Errorf("Identifier \"%s\" must be a %s not %s.", name, resultTypeInfo.String(), variableTypeInfo.String()))
				}
			case token.String:
				expectedTypeInfo := p.typeinfo.NewTypeInfoString()
				if resultTypeInfo == nil {
					resultTypeInfo = expectedTypeInfo
				}
				if !TypeEquals(resultTypeInfo, expectedTypeInfo) {
					// NOTE(Jake): 2018-01-04
					//
					// Need to make mixing types error messages consistent. Either by removing
					// the look-ahead for an operator token or by improving the check.
					//
					var opToken token.Token
					if i+1 < len(nodes) {
						switch node := nodes[i+1].(type) {
						case *ast.Token:
							opToken = node.Token
						case *ast.TokenList:
							opToken = node.Tokens()[0]
						default:
							panic(fmt.Sprintf("Expected *ast.Token or *ast.TokenList, not %T", node))
						}
					}
					p.AddError(node.Token, fmt.Errorf("Cannot %s (%s) %s %s (\"%s\"), mismatching types.", resultTypeInfo.String(), leftToken.String(), opToken.String(), expectedTypeInfo.String(), node.String()))
				}
			case token.Number:
				IntTypeInfo := p.typeinfo.NewTypeInfoInt()
				FloatTypeInfo := p.typeinfo.NewTypeInfoFloat()

				if resultTypeInfo == nil {
					resultTypeInfo = IntTypeInfo
					if strings.ContainsRune(node.Data, '.') {
						resultTypeInfo = FloatTypeInfo
					}
				}
				if !TypeEquals(resultTypeInfo, IntTypeInfo) && !TypeEquals(resultTypeInfo, FloatTypeInfo) {
					p.AddError(node.Token, fmt.Errorf("Cannot use %s with number \"%s\"", resultTypeInfo.String(), node.String()))
				}
			case token.KeywordTrue, token.KeywordFalse:
				expectedTypeInfo := p.typeinfo.NewTypeInfoBool()
				if resultTypeInfo == nil {
					resultTypeInfo = expectedTypeInfo
				}
				if !TypeEquals(resultTypeInfo, expectedTypeInfo) {
					p.AddError(node.Token, fmt.Errorf("Cannot use %s with %s \"%s\"", resultTypeInfo.String(), expectedTypeInfo.String(), node.String()))
				}
			default:
				panic(fmt.Sprintf("typerExpression: Unhandled token kind: \"%s\" with value: %s", node.Kind.String(), node.String()))
			}
			leftToken = node.Token
			continue
		}
		panic(fmt.Sprintf("typerExpression: Unhandled type %T", itNode))
	}

	expression.TypeInfo = resultTypeInfo
}

//func (p *Typer) typerHTMLBlock(htmlBlock *ast.HTMLBlock, scope *Scope) {
//	scope = NewScope(scope)
//	p.typerStatements(htmlBlock, scope)
//}

func (p *Typer) getTypeFromLeftHandSide(leftHandSideTokens []token.Token, scope *Scope) types.TypeInfo {
	nameToken := leftHandSideTokens[0]
	if nameToken.Kind != token.Identifier {
		p.PanicError(nameToken, fmt.Errorf("Expected identifier on left hand side, instead got %s.", nameToken.Kind.String()))
		return nil
	}
	name := nameToken.String()
	symbol := scope.GetSymbol(name)
	if symbol == nil {
		p.AddError(nameToken, fmt.Errorf("Undeclared variable \"%s\".", name))
		return nil
	}
	variableTypeInfo := symbol.variable
	if variableTypeInfo == nil {
		p.AddError(nameToken, fmt.Errorf("Identifier \"%s\" is not a variable", name))
		return nil
	}
	if variableTypeInfo == nil {
		p.PanicError(nameToken, fmt.Errorf("Variable declared \"%s\" but it has \"nil\" type information, this is a bug in typering scope.", name))
		return nil
	}
	var concatPropertyName bytes.Buffer
	concatPropertyName.WriteString(name)
	for i := 1; i < len(leftHandSideTokens); i++ {
		propertyName := leftHandSideTokens[i].String()
		concatPropertyName.WriteString(".")
		concatPropertyName.WriteString(propertyName)
		structInfo, ok := variableTypeInfo.(*types.Struct)
		if !ok {
			p.AddError(nameToken, fmt.Errorf("Property \"%s\" does not exist on type \"%s\".", concatPropertyName.String(), variableTypeInfo.String()))
			return nil
		}
		structField := structInfo.GetFieldByName(propertyName)
		if structField == nil {
			p.AddError(nameToken, fmt.Errorf("Property \"%s\" does not exist on \"%s :: struct\".", concatPropertyName.String(), structInfo.Name()))
			return nil
		}
		variableTypeInfo = structField.TypeInfo
	}
	return variableTypeInfo
}

func (p *Typer) typerHTMLNode(scope *Scope, node *ast.Call) {
	p.typerExpression(scope, &node.IfExpression)

	for i, _ := range node.Parameters {
		p.typerExpression(scope, &node.Parameters[i].Expression)
	}

	name := node.Name.String()
	isValidHTML5TagName := util.IsValidHTML5TagName(name)
	if isValidHTML5TagName {
		return
	}
	symbol := scope.GetSymbol(name)
	if symbol == nil {
		if name != strings.ToLower(name) {
			p.AddError(node.Name, fmt.Errorf("\"%s\" is an undefined component. If you want to use a standard HTML5 element, the name must all be in lowercase.", name))
			return
		}
		p.AddError(node.Name, fmt.Errorf("\"%s\" is not a valid HTML5 element or defined component.", name))
		return
	}
	htmlDefinition := symbol.htmlDefinition
	if htmlDefinition == nil {
		p.AddError(node.Name, fmt.Errorf("Expected \"%s\" to be \"%s :: html\", not \"%s\".", name, name, symbol.GetType()))
		return
	}
	//fmt.Printf("%s -- %d\n", htmlComponentDefinition.Name.String(), len(p.typerHtmlDefinitionStack))
	//for _, itHtmlDefinition := range p.typerHtmlDefinitionStack {
	//	if htmlComponentDefinition == itHtmlDefinition {
	//		p.addErrorLine(fmt.Errorf("Cannot reference self in \"%s :: html\".", htmlComponentDefinition.Name.String()), node.Name.Line)
	//		//continue Loop
	//		return
	//	}
	//}

	if p.typecheckHtmlNodeDependencies != nil {
		p.typecheckHtmlNodeDependencies[name] = node
	}
	node.HTMLDefinition = htmlDefinition

	// Mark this component as used
	{
		isFound := false
		for _, htmlDefinitionUsed := range p.htmlComponentsUsed {
			isFound = isFound || htmlDefinition == htmlDefinitionUsed
		}
		if !isFound {
			p.htmlComponentsUsed = append(p.htmlComponentsUsed, node.HTMLDefinition)
		}
	}

	structDef := node.HTMLDefinition.Struct
	if structDef != nil && len(structDef.Fields) > 0 {
		// Check if parameters exist
	ParameterCheckLoop:
		for i, _ := range node.Parameters {
			parameterNode := node.Parameters[i]
			paramName := parameterNode.Name.String()
			for _, field := range structDef.Fields {
				if paramName == field.Name.String() {
					parameterType := parameterNode.TypeInfo
					componentStructType := field.TypeInfo
					if parameterType != componentStructType {
						if field.TypeInfo == nil {
							p.PanicMessage(fmt.Errorf("Struct field \"%s\" is missing type info.", paramName))
							return
						}
						p.AddError(parameterNode.Name, fmt.Errorf("\"%s\" must be of type %s, not %s", paramName, componentStructType.String(), parameterType.String()))
					}
					continue ParameterCheckLoop
				}
			}
			p.AddError(parameterNode.Name, fmt.Errorf("\"%s\" is not a property on \"%s :: html\"", paramName, name))
			continue
		}
	}
}

func (p *Typer) typerCSSDefinition(cssDef *ast.CSSDefinition) {
	scope := NewScope(nil)
	p.typerStatements(cssDef, scope)
}

func (p *Typer) typerHTMLDefinition(htmlDefinition *ast.HTMLComponentDefinition, parentScope *Scope) {
	name := htmlDefinition.Name.String()
	symbol := parentScope.GetSymbol(name)
	if symbol == nil {
		panic(fmt.Sprintf("Cannot find symbol for \"%s :: html\", this should not be possible.", name))
	}

	// Attach CSSDefinition if found
	if symbol.cssDefinition != nil {
		htmlDefinition.CSSDefinition = symbol.cssDefinition
	}

	// Attach CSSConfigDefinition if found
	if symbol.cssConfigDefinition != nil {
		htmlDefinition.CSSConfigDefinition = symbol.cssConfigDefinition
	}

	// Attach StructDefinition if found
	if structDef := symbol.structDefinition; structDef != nil {
		if htmlDefinition.Struct != nil {
			anonymousStructDef := htmlDefinition.Struct
			p.AddError(anonymousStructDef.Name, fmt.Errorf("Cannot have \"%s :: struct\" and embedded \":: struct\" inside \"%s :: html\"", structDef.Name.String(), htmlDefinition.Name.String()))
		} else {
			htmlDefinition.Struct = structDef
		}
	}

	//
	// NOTE(Jake): 2018-04-15
	//
	// Used to do the following before the combined symbol
	// refactor.
	//
	// var globalScopeNoVariables Scope = *parentScope
	// globalScopeNoVariables.identifiers = nil
	// scope := NewScope(globalScopeNoVariables)
	//
	scope := NewScope(parentScope)
	scope.SetVariable("children", p.typeinfo.NewHTMLNode())

	if structDef := htmlDefinition.Struct; structDef != nil {
		for i, _ := range structDef.Fields {
			var propertyNode *ast.StructField = &structDef.Fields[i]
			p.typerExpression(scope, &propertyNode.Expression)
			name := propertyNode.Name.String()
			if symbol := scope.GetSymbol(name); symbol != nil {
				if name == "children" {
					p.AddError(propertyNode.Name, fmt.Errorf("Cannot use \"children\" as it's a reserved property."))
					continue
				}
				p.AddError(propertyNode.Name, fmt.Errorf("Property \"%s\" declared twice.", name))
				continue
			}
			/*_, ok := scope.GetVariable(name)
			if ok {
				if name == "children" {
					p.AddError(propertyNode.Name, fmt.Errorf("Cannot use \"children\" as it's a reserved property."))
					continue
				}
				p.AddError(propertyNode.Name, fmt.Errorf("Property \"%s\" declared twice.", name))
				continue
			}*/
			scope.SetVariable(name, propertyNode.TypeInfo)
		}
	}

	if p.typecheckHtmlNodeDependencies != nil {
		panic("typecheckHtmlNodeDependencies must be nil before being re-assigned")
	}
	p.typecheckHtmlNodeDependencies = make(map[string]*ast.Call)
	p.typerStatements(htmlDefinition, scope)
	htmlDefinition.Dependencies = p.typecheckHtmlNodeDependencies
	p.typecheckHtmlNodeDependencies = nil
}

func (p *Typer) typerCSSProperty(property *ast.CSSProperty, scope *Scope) {
	for _, node := range property.Nodes() {
		switch node := node.(type) {
		case *ast.TokenList:
			panic("todo: Handle typechecking of property vars, ie. myval.property")
			//opcodes, _ = emit.emitVariableIdentWithProperty(opcodes, node.Tokens())
		case *ast.Call:
			p.typerCall(scope, node)
		case *ast.Token:
			t := node.Token
			switch t.Kind {
			case token.Identifier,
				token.Number,
				token.String:
				// no-op, valid token kind
			default: // ie. number, string
				panic(fmt.Sprintf("emitCSSProperty: Unhandled token kind: %s", node.Kind.String()))
			}
		default:
			panic(fmt.Sprintf("emitCSSProperty: Unhandled type: %T", node))
		}
	}
}

func (p *Typer) typerStatements(topNode ast.Node, scope *Scope) {
	nodeStack := make([]ast.Node, 0, 50)
	nodes := topNode.Nodes()
	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		nodeStack = append(nodeStack, node)
	}

	//Loop:
	for len(nodeStack) > 0 {
		itNode := nodeStack[len(nodeStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]
		avoidNestingScopeThisIteration := false

		if itNode == nil {
			scope = scope.parent
			continue
		}

		switch node := itNode.(type) {
		case *ast.CSSDefinition,
			*ast.CSSConfigDefinition,
			*ast.HTMLComponentDefinition,
			*ast.StructDefinition,
			*ast.ProcedureDefinition:
			// Skip nodes and child nodes
			continue
		case *ast.WorkspaceDefinition:
			// NOTE(Jake): 2018-04-15
			//
			// Typecheck this at statement level so that all procedures / structs / etc
			// in the scope of the workspace file are accessible.
			//
			if node == nil {
				p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
				continue
			}
			p.typerWorkspaceDefinition(scope, node)
		case *ast.Call:
			p.typerCall(scope, node)
		case *ast.HTMLBlock:
			panic("todo(Jake): Remove below commented out line if this is unused.")
			//p.typerHTMLBlock(node, scope)
		case *ast.Return:
			p.typerExpression(scope, &node.Expression)
			continue
		case *ast.ArrayAppendStatement:
			nameToken := node.LeftHandSide[0]
			name := nameToken.String()
			arrayTypeInfo := p.getTypeFromLeftHandSide(node.LeftHandSide, scope)
			if arrayTypeInfo == nil {
				continue
			}
			variableTypeInfo, isArray := arrayTypeInfo.(*types.Array)
			if !isArray {
				// todo(Jake): 2017-12-25
				//
				// Retrieve the whole string for the token list and use instead of "name"
				//
				p.AddError(nameToken, fmt.Errorf("Cannot do \"%s []=\" as it's not an array type.", name))
				continue
			}
			p.typerExpression(scope, &node.Expression)
			resultTypeInfo := node.Expression.TypeInfo
			if !TypeEquals(variableTypeInfo.Underlying(), resultTypeInfo) {
				// todo(Jake): 2017-12-25
				//
				// test these... as the error messages are untested
				//
				p.AddError(nameToken, fmt.Errorf("Cannot change \"%s\" from %s to %s", name, variableTypeInfo, resultTypeInfo.String()))
				p.AddError(nameToken, fmt.Errorf("Cannot change \"%s\" from %s to %s", name, variableTypeInfo, resultTypeInfo.String()))
			}
			continue
		case *ast.OpStatement:
			variableTypeInfo := p.getTypeFromLeftHandSide(node.LeftHandSide, scope)
			if variableTypeInfo == nil {
				continue
			}
			p.typerExpression(scope, &node.Expression)
			resultTypeInfo := node.Expression.TypeInfo
			if !TypeEquals(variableTypeInfo, resultTypeInfo) {
				nameToken := node.LeftHandSide[0]
				name := nameToken.String()
				for i := 1; i < len(node.LeftHandSide); i++ {
					name += "." + node.LeftHandSide[i].String()
				}
				if variableTypeInfo == nil {
					p.PanicError(nameToken, fmt.Errorf("\"variableTypeInfo\" is nil, \"%s\" should have type info.", name))
					continue
				}
				if resultTypeInfo == nil {
					p.PanicError(nameToken, fmt.Errorf("\"resultTypeInfo\" is nil, right-side of \"%s\" should have type info.", name))
					continue
				}
				p.AddError(nameToken, fmt.Errorf("Cannot change \"%s\" from %s to %s", name, variableTypeInfo, resultTypeInfo.String()))
			}
			continue
		case *ast.DeclareStatement:
			expr := &node.Expression
			p.typerExpression(scope, expr)
			name := node.Name.String()
			if symbol := scope.GetSymbolFromThisScope(name); symbol != nil {
				p.AddError(node.Name, fmt.Errorf("Cannot redeclare \"%s\".", name))
				continue
			}
			scope.SetVariable(name, expr.TypeInfo)
			continue
		case *ast.Expression:
			p.typerExpression(scope, node)
			continue
		case *ast.If:
			expr := &node.Condition
			//expr.TypeInfo = types.Bool()
			p.typerExpression(scope, expr)

			scope = NewScope(scope)
			nodeStack = append(nodeStack, nil)
			avoidNestingScopeThisIteration = true

			// Add if true children
			{
				nodes := node.Nodes()
				for i := len(nodes) - 1; i >= 0; i-- {
					nodeStack = append(nodeStack, nodes[i])
				}
			}
			{
				// Add else children
				nodes := node.ElseNodes
				for i := len(nodes) - 1; i >= 0; i-- {
					nodeStack = append(nodeStack, nodes[i])
				}
			}
			continue
		case *ast.CSSProperty:
			p.typerCSSProperty(node, scope)
			continue
		case *ast.Block,
			*ast.CSSRule:
			// no-op, will jump to adding child nodes / new scope below
		case *ast.For:
			if !node.IsDeclareSet {
				panic("todo(Jake): handle array without declare set")
			}
			p.typerExpression(scope, &node.Array)
			iTypeInfo := node.Array.TypeInfo
			typeInfo, ok := iTypeInfo.(*types.Array)
			if !ok {
				p.AddError(node.RecordName, fmt.Errorf("Cannot use type %s as array.", iTypeInfo.String()))
				continue
			}
			if node.IsDeclareSet {
				// Nest scope
				// - Earlier nesting so we declare variables in the `for` line rather
				//	 then only after the {
				//
				// WARNING: Ensure nothing else appends to `nodeStack` after this.
				//
				scope = NewScope(scope)
				nodeStack = append(nodeStack, nil)
				avoidNestingScopeThisIteration = true
			}

			// Add "i" to scope if used
			if node.IndexName.Kind != token.Unknown {
				indexName := node.IndexName.String()
				if scope := scope.GetSymbolFromThisScope(indexName); scope != nil {
					p.AddError(node.IndexName, fmt.Errorf("Cannot redeclare \"%s\" in for-loop.", indexName))
					continue
				}
				scope.SetVariable(indexName, p.typeinfo.NewTypeInfoInt())
			}

			// Set left-hand value type
			name := node.RecordName.String()
			if scope := scope.GetSymbolFromThisScope(name); scope != nil {
				p.AddError(node.RecordName, fmt.Errorf("Cannot redeclare \"%s\" in for-loop.", name))
				continue
			}
			scope.SetVariable(name, typeInfo.Underlying())
		default:
			panic(fmt.Sprintf("TypecheckStatements: Unknown type %T", node))
		}

		// Nest scope
		if !avoidNestingScopeThisIteration {
			scope = NewScope(scope)
			nodeStack = append(nodeStack, nil)
		}

		// Add children
		nodes := itNode.Nodes()
		for i := len(nodes) - 1; i >= 0; i-- {
			nodeStack = append(nodeStack, nodes[i])
		}
	}
}

func (p *Typer) typerStruct(node *ast.StructDefinition, scope *Scope) {
	// Add typeinfo to each struct field
	for i := 0; i < len(node.Fields); i++ {
		structField := &node.Fields[i]
		typeIdent := structField.TypeIdentifier.Name
		if typeIdent.Kind == token.Unknown {
			p.AddError(structField.Name, fmt.Errorf("Missing type identifier on \"%s :: struct\" field \"%s\"", structField.Name, node.Name.String()))
			continue
		}
		typeIdentString := typeIdent.String()
		resultTypeInfo := p.DetermineType(&structField.TypeIdentifier)
		if resultTypeInfo == nil {
			p.AddError(typeIdent, fmt.Errorf("Undeclared type %s", typeIdentString))
			return
		}
		structField.TypeInfo = resultTypeInfo
		if len(structField.Expression.Nodes()) > 0 {
			// NOTE(Jake): 2018-04-13
			//
			// Before combined symbols refactor, this simply passed in
			// NewScope(nil)
			//
			p.typerExpression(scope, &structField.Expression)
		}
	}
}

func (p *Typer) typerProcedureDefinition(node *ast.ProcedureDefinition, scope *Scope) {
	name := node.Name.String()
	symbol := scope.getOrCreateSymbol(name)
	if typeInfo := symbol.variable; typeInfo != nil {
		errorMessage := fmt.Errorf("Cannot redeclare \"%s :: ()\" more than once in global scope.", name)
		//p.AddError(symbol..Name, errorMessage)
		p.AddError(node.Name, errorMessage)
		return
	}

	scope = NewScope(scope)
	for i := 0; i < len(node.Parameters); i++ {
		parameter := &node.Parameters[i]
		typeinfo := p.DetermineType(&parameter.TypeIdentifier)
		if typeinfo == nil {
			p.AddError(parameter.TypeIdentifier.Name, fmt.Errorf("Unknown type %s on parameter %s", parameter.TypeIdentifier.String(), parameter.Name))
			continue
		}
		parameter.TypeInfo = typeinfo
		scope.SetVariable(parameter.Name.String(), typeinfo)
	}
	if p.HasErrors() {
		return
	}
	p.typerStatements(node, scope)

	var returnType types.TypeInfo
	if node.TypeIdentifier.Name.Kind != token.Unknown {
		returnType = p.DetermineType(&node.TypeIdentifier)
		node.TypeInfo = returnType
	}
	// Check return statements
	// NOTE(Jake): 2017-12-30
	//
	// The code below is probably super slow compared
	// to if we just added a flag/callback to `typerStatements`
	// this would probably be faster.
	//
	nodes := node.ChildNodes
	nodeStack := make([]ast.Node, 0, len(nodes))
	for i := len(nodes) - 1; i >= 0; i-- {
		nodeStack = append(nodeStack, nodes[i])
	}
	for len(nodeStack) > 0 {
		node := nodeStack[len(nodeStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]

		returnNode, ok := node.(*ast.Return)
		if ok {
			if TypeEquals(returnType, returnNode.TypeInfo) {
				continue
			}
			t := returnNode.TypeIdentifier.Name
			if returnType == nil {
				p.AddError(t, fmt.Errorf("Return statement %s doesn't match procedure type void", returnNode.TypeInfo.String()))
				continue
			}
			if returnNode.TypeInfo == nil {
				p.AddError(t, fmt.Errorf("Return statement void doesn't match procedure type %s", returnType.String()))
				continue
			}
			p.AddError(t, fmt.Errorf("Return statement %s doesn't match procedure type %s", returnNode.TypeInfo.String(), returnType.String()))
			continue
		}

		nodes := node.Nodes()
		for i := len(nodes) - 1; i >= 0; i-- {
			nodeStack = append(nodeStack, nodes[i])
		}
	}

	functionType := p.typeinfo.NewProcedureInfo(node)
	symbol.variable = functionType
	p.typeinfo.register(node.Name.String(), functionType)
}

func (p *Typer) typerWorkspaceDefinition(scope *Scope, node *ast.WorkspaceDefinition) {
	node.WorkspaceTypeInfo = p.typeinfo.InternalWorkspaceStruct()
	scope.SetVariable("workspace", node.WorkspaceTypeInfo)
	p.typerStatements(node, scope)
}

func (p *Typer) typecheckFile(file *ast.File, globalScope *Scope) {
	scope := NewScope(globalScope)
	p.typerStatements(file, scope)
}

func (p *Typer) ApplyTypeInfoAndTypecheck(files []*ast.File) {
	globalScope := NewScope(nil)

	//
	globalScopeHtmlDefinitions := make([]*ast.HTMLComponentDefinition, 0, 10)
	globalScopeCssConfigDefinitions := make([]*ast.CSSConfigDefinition, 0, 10)

	// Get all global/top-level identifiers
	for _, file := range files {
		scope := globalScope
		for _, node := range file.ChildNodes {
			switch node := node.(type) {
			case *ast.DeclareStatement,
				*ast.OpStatement,
				*ast.ArrayAppendStatement,
				*ast.Expression,
				*ast.If,
				*ast.For,
				*ast.Block,
				*ast.HTMLBlock,
				*ast.Call,
				*ast.WorkspaceDefinition,
				*ast.Return:
				// no-op, these are checked in TypecheckFile()
			case *ast.ProcedureDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				p.typerProcedureDefinition(node, scope)
			case *ast.StructDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				if node.Name.Kind == token.Unknown {
					p.AddError(node.Name, fmt.Errorf("Cannot declare anonymous \":: struct\" block."))
					continue
				}
				name := node.Name.String()
				symbol := scope.getOrCreateSymbol(name)
				if definition := symbol.structDefinition; definition != nil {
					errorMessage := fmt.Errorf("Cannot redeclare \"%s :: struct\" more than once in global scope.", name)
					p.AddError(definition.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}
				p.typerStruct(node, scope)
				symbol.structDefinition = node
				p.typeinfo.register(name, p.typeinfo.NewStructInfo(node))
			case *ast.HTMLComponentDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				if node.Name.Kind == token.Unknown {
					p.AddError(node.Name, fmt.Errorf("Cannot declare anonymous \":: html\" block."))
					continue
				}
				name := node.Name.String()
				symbol := scope.getOrCreateSymbol(name)
				if definition := symbol.htmlDefinition; definition != nil {
					errorMessage := fmt.Errorf("Cannot redeclare \"%s :: html\" more than once in global scope.", name)
					p.AddError(definition.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}
				symbol.htmlDefinition = node
				globalScopeHtmlDefinitions = append(globalScopeHtmlDefinitions, node)
			case *ast.CSSDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				if node.Name.Kind == token.Unknown {
					continue
				}
				name := node.Name.String()
				symbol := scope.getOrCreateSymbol(name)
				if definition := symbol.cssDefinition; definition != nil {
					errorMessage := fmt.Errorf("Cannot redeclare \"%s :: html\" more than once in global scope.", name)
					p.AddError(definition.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}
				p.typerCSSDefinition(node)
				symbol.cssDefinition = node
			case *ast.CSSConfigDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				if node.Name.Kind == token.Unknown {
					p.AddError(node.Name, fmt.Errorf("Cannot declare anonymous \":: css_config\" block."))
					continue
				}
				name := node.Name.String()
				symbol := scope.getOrCreateSymbol(name)
				if definition := symbol.cssConfigDefinition; definition != nil {
					errorMessage := fmt.Errorf("Cannot redeclare \"%s :: css_config\" more than once in global scope.", name)
					p.AddError(definition.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}
				symbol.cssConfigDefinition = node
				globalScopeCssConfigDefinitions = append(globalScopeCssConfigDefinitions, node)
			default:
				panic(fmt.Sprintf("TypecheckAndFinalize: Unknown type %T", node))
			}
		}
	}

	//
	for _, htmlDefinition := range globalScopeHtmlDefinitions {
		p.typerHTMLDefinition(htmlDefinition, globalScope)
	}

	// Check if CSS config matches a HTML or CSS component. If not, throw error.
	for _, cssConfigDefinition := range globalScopeCssConfigDefinitions {
		name := cssConfigDefinition.Name.String()
		symbol := globalScope.GetSymbolFromThisScope(name)

		hasHTMLDefinition := symbol != nil &&
			symbol.htmlDefinition != nil
		hasCSSDefinition := symbol != nil &&
			symbol.cssDefinition != nil
		if !hasCSSDefinition || !hasHTMLDefinition {
			p.AddError(cssConfigDefinition.Name, fmt.Errorf("\"%s :: css_config\" requires both a matching \":: css\" or \":: html\" definition.", name))
		}
	}

	// Get nested dependencies
	for _, htmlDefinition := range globalScopeHtmlDefinitions {
		nodeStack := make([]*ast.Call, 0, 50)
		for _, subNode := range htmlDefinition.Dependencies {
			nodeStack = append(nodeStack, subNode)
		}
		for len(nodeStack) > 0 {
			node := nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]

			// Add child dependencies
			for _, subNode := range node.HTMLDefinition.Dependencies {
				name := subNode.Name.String()
				_, ok := htmlDefinition.Dependencies[name]
				if ok {
					continue
				}
				htmlDefinition.Dependencies[name] = subNode
				nodeStack = append(nodeStack, subNode)
			}
		}

		// Print deps
		// fmt.Printf("\n\nDependencies of %s\n", htmlDefinition.Name.String())
		// for name, _ := range htmlDefinition.Dependencies {
		// 	fmt.Printf("- %s\n", name)
		// }
	}

	// Lookup if component depends on itself
	for _, htmlDefinition := range globalScopeHtmlDefinitions {
		name := htmlDefinition.Name.String()
		node, ok := htmlDefinition.Dependencies[name]
		if !ok {
			continue
		}
		//
		// todo(Jake): 2018-04-13
		//
		// Add a better dependency solver wherein it has detailed information
		// about what isn't allowed, rather than a vague message.
		//
		p.AddError(node.Name, fmt.Errorf("Cannot use \"%s\". Cyclic references are not allowed.", name))
	}

	// Typecheck
	for _, file := range files {
		p.typecheckFile(file, globalScope)
	}
}

/*func (p *Typer) checkForRedeclareErrors(nameToken token.Token, scope *Scope) bool {
	name := nameToken.String()
	if symbol := scope.GetSymbolFromThisScope(name); symbol != nil {
		errorMessage := fmt.Errorf("Cannot redeclare \"%s\" more than once in global scope.", name)
		if structDef := symbol.structDefinition; structDef != nil {
			p.AddError(structDef.Name, errorMessage)
		} else if htmlDef := symbol.htmlDefinition; htmlDef != nil {
			p.AddError(htmlDef.Name, errorMessage)
		} else if cssDef := symbol.cssDefinition; cssDef != nil {
			p.AddError(cssDef.Name, errorMessage)
		}
		p.AddError(nameToken, errorMessage)
		return true
	}
	return false
}*/
