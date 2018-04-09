package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
	"github.com/silbinarywolf/compiler-fel/util"
)

/*type Checker struct {
	*Typer
	scope *Scope
}

func (checker *Checker) PushScope() {
	scope := NewScope(checker.scope)
	checker.scope = scope
}

func (checker *Checker) PopScope() {
	parentScope := checker.scope.parent
	if parentScope == nil {
		panic("Cannot pop last scope item.")
	}
	checker.scope = parentScope
}*/

// func getDataTypeFromTokaen(t token.Token) data.Kind {
// 	switch t.Kind {
// 	case token.Identifier:
// 		typename := t.String()
// 		switch typename {
// 		case "string":
// 			return data.KindString
// 		case "int", "int64":
// 			return data.KindInteger64
// 		case "float", "float64":
// 			return data.KindFloat64
// 		case "html_node":
// 			return data.KindHTMLNode
// 		default:
// 			panic(fmt.Sprintf("Unknown type name: %s", typename))
// 		}
// 	default:
// 		panic(fmt.Sprintf("Cannot use token kind %s in type declaration", t.Kind.String()))
// 	}
// }

func (p *Typer) typecheckStructLiteral(scope *Scope, literal *ast.StructLiteral) {
	name := literal.Name.String()
	def, ok := scope.GetStructDefinition(name)
	if !ok {
		p.AddError(literal.Name, fmt.Errorf("Undeclared \"%s :: struct\"", name))
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
			p.typecheckExpression(scope, &property.Expression)
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

func (p *Typer) typecheckArrayLiteral(scope *Scope, literal *ast.ArrayLiteral) {
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
	resultTypeInfo, ok := typeInfo.(*TypeInfo_Array)
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
		//			   when we call `typecheckExpression`
		//			   ie. Won't infer, will mark as invalid.
		if node.TypeInfo == nil {
			node.TypeInfo = underlyingTypeInfo
		}
		p.typecheckExpression(scope, node)

		if node.TypeInfo == nil {
			panic(fmt.Sprintf("typecheckArrayLiteral: Missing type on array literal item #%d.", i))
		}
	}
}

func (p *Typer) typecheckCall(scope *Scope, node *ast.Call) {
	switch node.Kind() {
	case ast.CallProcedure:
		p.typecheckProcedureCall(scope, node)
	case ast.CallHTMLNode:
		p.typecheckHTMLNode(scope, node)
	default:
		p.PanicError(node.Name, fmt.Errorf("Unhandled ast.Call kind: %s", node.Name))
	}
}

func (p *Typer) typecheckProcedureCall(scope *Scope, node *ast.Call) {
	typeInfo := p.typeinfo.getByName(node.Name.String())
	callTypeInfo, ok := typeInfo.(*TypeInfo_Procedure)
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
		p.typecheckExpression(scope, &parameter.Expression)
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
				// This should be applied in p.typecheckExpression(scope, &parameter.Expression)
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
				// This should be applied in p.typecheckExpression(scope, &parameter.Expression)
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

func (p *Typer) typecheckExpression(scope *Scope, expression *ast.Expression) {
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
			p.typecheckStructLiteral(scope, node)
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
			p.typecheckArrayLiteral(scope, node)
			expectedTypeInfo := node.TypeInfo
			if resultTypeInfo == nil {
				resultTypeInfo = expectedTypeInfo
			}
			if !TypeEquals(resultTypeInfo, expectedTypeInfo) {
				p.AddError(node.TypeIdentifier.Name, fmt.Errorf("Cannot mix array literal %s with %s", expectedTypeInfo.String(), resultTypeInfo.String()))
			}
			continue
		case *ast.Call:
			p.typecheckCall(scope, node)
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
				panic(fmt.Sprintf("typecheckExpression:Call: Unhandled call kind: %s", node.Kind()))
			}

			continue
		case *ast.HTMLBlock:
			panic("typecheckExpression: todo(Jake): Fix HTMLBlock")
			/*variableType := data.KindHTMLNode
			if exprType == data.KindUnknown {
				exprType = variableType
			}
			if exprType != variableType {
				p.addErrorToken(fmt.Errorf("\":: html\" must be a %s not %s.", exprType.String(), variableType.String()), node.HTMLKeyword)
			}
			p.typecheckHTMLBlock(node, scope)*/
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
				variableTypeInfo, ok := scope.Get(name)
				if !ok {
					_, ok := scope.GetHTMLDefinition(name)
					if ok {
						p.AddError(node.Token, fmt.Errorf("Undeclared identifier \"%s\". Did you mean \"%s()\" or \"%s{ }\" to reference the \"%s :: html\" component?", name, name, name, name))
						continue
					}
					p.AddError(node.Token, fmt.Errorf("Undeclared identifier \"%s\".", name))
					continue
				}
				if resultTypeInfo == nil {
					resultTypeInfo = variableTypeInfo
				}
				if !TypeEquals(resultTypeInfo, variableTypeInfo) {
					if variableTypeInfo == nil {
						p.PanicError(node.Token, fmt.Errorf("Unable to determine type, typechecker must have failed."))
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
				panic(fmt.Sprintf("typecheckExpression: Unhandled token kind: \"%s\" with value: %s", node.Kind.String(), node.String()))
			}
			leftToken = node.Token
			continue
		}
		panic(fmt.Sprintf("typecheckExpression: Unhandled type %T", itNode))
	}

	expression.TypeInfo = resultTypeInfo
}

//func (p *Typer) typecheckHTMLBlock(htmlBlock *ast.HTMLBlock, scope *Scope) {
//	scope = NewScope(scope)
//	p.typecheckStatements(htmlBlock, scope)
//}

func (p *Typer) getTypeFromLeftHandSide(leftHandSideTokens []token.Token, scope *Scope) types.TypeInfo {
	nameToken := leftHandSideTokens[0]
	if nameToken.Kind != token.Identifier {
		p.PanicError(nameToken, fmt.Errorf("Expected identifier on left hand side, instead got %s.", nameToken.Kind.String()))
		return nil
	}
	name := nameToken.String()
	variableTypeInfo, ok := scope.Get(name)
	if !ok {
		p.AddError(nameToken, fmt.Errorf("Undeclared variable \"%s\".", name))
		return nil
	}
	if variableTypeInfo == nil {
		p.PanicError(nameToken, fmt.Errorf("Variable declared \"%s\" but it has \"nil\" type information, this is a bug in typechecking scope.", name))
		return nil
	}
	var concatPropertyName bytes.Buffer
	concatPropertyName.WriteString(name)
	for i := 1; i < len(leftHandSideTokens); i++ {
		propertyName := leftHandSideTokens[i].String()
		concatPropertyName.WriteString(".")
		concatPropertyName.WriteString(propertyName)
		structInfo, ok := variableTypeInfo.(*TypeInfo_Struct)
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

func (p *Typer) typecheckHTMLNode(scope *Scope, node *ast.Call) {
	p.typecheckExpression(scope, &node.IfExpression)

	for i, _ := range node.Parameters {
		p.typecheckExpression(scope, &node.Parameters[i].Expression)
	}

	name := node.Name.String()
	isValidHTML5TagName := util.IsValidHTML5TagName(name)
	if isValidHTML5TagName {
		return
	}
	htmlComponentDefinition, ok := scope.GetHTMLDefinition(name)
	if !ok {
		if name != strings.ToLower(name) {
			p.AddError(node.Name, fmt.Errorf("\"%s\" is an undefined component. If you want to use a standard HTML5 element, the name must all be in lowercase.", name))
			return
		}
		p.AddError(node.Name, fmt.Errorf("\"%s\" is not a valid HTML5 element or defined component.", name))
		return
	}
	//fmt.Printf("%s -- %d\n", htmlComponentDefinition.Name.String(), len(p.typecheckHtmlDefinitionStack))
	//for _, itHtmlDefinition := range p.typecheckHtmlDefinitionStack {
	//	if htmlComponentDefinition == itHtmlDefinition {
	//		p.addErrorLine(fmt.Errorf("Cannot reference self in \"%s :: html\".", htmlComponentDefinition.Name.String()), node.Name.Line)
	//		//continue Loop
	//		return
	//	}
	//}
	if p.typecheckHtmlNodeDependencies != nil {
		p.typecheckHtmlNodeDependencies[name] = node
	}
	node.HTMLDefinition = htmlComponentDefinition
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

func (p *Typer) typecheckHTMLDefinition(htmlDefinition *ast.HTMLComponentDefinition, parentScope *Scope) {
	// Attach CSSDefinition if found
	name := htmlDefinition.Name.String()
	if cssDefinition, ok := parentScope.GetCSSDefinition(name); ok {
		htmlDefinition.CSSDefinition = cssDefinition
	}

	// Attach CSSConfigDefinition if found
	if cssConfigDefinition, ok := parentScope.GetCSSConfigDefinition(name); ok {
		htmlDefinition.CSSConfigDefinition = cssConfigDefinition
	}

	// Attach StructDefinition if found
	if structDef, ok := parentScope.GetStructDefinition(name); ok {
		if htmlDefinition.Struct != nil {
			anonymousStructDef := htmlDefinition.Struct
			p.AddError(anonymousStructDef.Name, fmt.Errorf("Cannot have \"%s :: struct\" and embedded \":: struct\" inside \"%s :: html\"", structDef.Name.String(), htmlDefinition.Name.String()))
		} else {
			htmlDefinition.Struct = structDef
		}
	}

	//
	var globalScopeNoVariables Scope = *parentScope
	globalScopeNoVariables.identifiers = nil
	scope := NewScope(&globalScopeNoVariables)
	scope.Set("children", p.typeinfo.NewHTMLNode())

	if structDef := htmlDefinition.Struct; structDef != nil {
		for i, _ := range structDef.Fields {
			var propertyNode *ast.StructField = &structDef.Fields[i]
			p.typecheckExpression(scope, &propertyNode.Expression)
			name := propertyNode.Name.String()
			_, ok := scope.Get(name)
			if ok {
				if name == "children" {
					p.AddError(propertyNode.Name, fmt.Errorf("Cannot use \"children\" as it's a reserved property."))
					continue
				}
				p.AddError(propertyNode.Name, fmt.Errorf("Property \"%s\" declared twice.", name))
				continue
			}
			scope.Set(name, propertyNode.TypeInfo)
		}
	}

	if p.typecheckHtmlNodeDependencies != nil {
		panic("typecheckHtmlNodeDependencies must be nil before being re-assigned")
	}
	p.typecheckHtmlNodeDependencies = make(map[string]*ast.Call)
	p.typecheckStatements(htmlDefinition, scope)
	htmlDefinition.Dependencies = p.typecheckHtmlNodeDependencies
	p.typecheckHtmlNodeDependencies = nil
}

func (p *Typer) typecheckStatements(topNode ast.Node, scope *Scope) {
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
			*ast.WorkspaceDefinition,
			*ast.ProcedureDefinition:
			// Skip nodes and child nodes
			continue
		case *ast.Call:
			p.typecheckCall(scope, node)
		case *ast.HTMLBlock:
			panic("todo(Jake): Remove below commented out line if this is unused.")
			//p.typecheckHTMLBlock(node, scope)
		case *ast.Return:
			p.typecheckExpression(scope, &node.Expression)
			continue
		case *ast.ArrayAppendStatement:
			nameToken := node.LeftHandSide[0]
			name := nameToken.String()
			arrayTypeInfo := p.getTypeFromLeftHandSide(node.LeftHandSide, scope)
			if arrayTypeInfo == nil {
				continue
			}
			variableTypeInfo, isArray := arrayTypeInfo.(*TypeInfo_Array)
			if !isArray {
				// todo(Jake): 2017-12-25
				//
				// Retrieve the whole string for the token list and use instead of "name"
				//
				p.AddError(nameToken, fmt.Errorf("Cannot do \"%s []=\" as it's not an array type.", name))
				continue
			}
			p.typecheckExpression(scope, &node.Expression)
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
			p.typecheckExpression(scope, &node.Expression)
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
			p.typecheckExpression(scope, expr)
			name := node.Name.String()
			_, ok := scope.GetFromThisScope(name)
			if ok {
				p.AddError(node.Name, fmt.Errorf("Cannot redeclare \"%s\".", name))
				continue
			}
			scope.Set(name, expr.TypeInfo)
			continue
		case *ast.Expression:
			p.typecheckExpression(scope, node)
			continue
		case *ast.If:
			expr := &node.Condition
			//expr.TypeInfo = types.Bool()
			p.typecheckExpression(scope, expr)

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
		case *ast.Block:
			// no-op, will jump to adding child nodes / new scope below
		case *ast.For:
			if !node.IsDeclareSet {
				panic("todo(Jake): handle array without declare set")
			}
			p.typecheckExpression(scope, &node.Array)
			iTypeInfo := node.Array.TypeInfo
			typeInfo, ok := iTypeInfo.(*TypeInfo_Array)
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
				_, ok = scope.GetFromThisScope(indexName)
				if ok {
					p.AddError(node.IndexName, fmt.Errorf("Cannot redeclare \"%s\" in for-loop.", indexName))
					continue
				}
				scope.Set(indexName, p.typeinfo.NewTypeInfoInt())
			}

			// Set left-hand value type
			name := node.RecordName.String()
			_, ok = scope.GetFromThisScope(name)
			if ok {
				p.AddError(node.RecordName, fmt.Errorf("Cannot redeclare \"%s\" in for-loop.", name))
				continue
			}
			scope.Set(name, typeInfo.Underlying())
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

func (p *Typer) typecheckStruct(node *ast.StructDefinition, scope *Scope) {
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
			p.typecheckExpression(NewScope(nil), &structField.Expression)
		}
	}
}

func (p *Typer) typecheckProcedureDefinition(node *ast.ProcedureDefinition, scope *Scope) {
	scope = NewScope(scope)

	for i := 0; i < len(node.Parameters); i++ {
		parameter := &node.Parameters[i]
		typeinfo := p.DetermineType(&parameter.TypeIdentifier)
		if typeinfo == nil {
			p.AddError(parameter.TypeIdentifier.Name, fmt.Errorf("Unknown type %s on parameter %s", parameter.TypeIdentifier.String(), parameter.Name))
			continue
		}
		parameter.TypeInfo = typeinfo
		scope.Set(parameter.Name.String(), typeinfo)
	}
	if p.HasErrors() {
		return
	}
	p.typecheckStatements(node, scope)

	var returnType types.TypeInfo
	if node.TypeIdentifier.Name.Kind != token.Unknown {
		returnType = p.DetermineType(&node.TypeIdentifier)
		node.TypeInfo = returnType
	}
	// Check return statements
	// NOTE(Jake): 2017-12-30
	//
	// The code below is probably super slow compared
	// to if we just added a flag/callback to `typecheckStatements`
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
	p.typeinfo.register(node.Name.String(), functionType)
}

func (p *Typer) typecheckWorkspaceDefinition(node *ast.WorkspaceDefinition) {
	scope := NewScope(nil)
	node.WorkspaceTypeInfo = p.typeinfo.InternalWorkspaceStruct()
	scope.Set("workspace", node.WorkspaceTypeInfo)
	p.typecheckStatements(node, scope)
}

func (p *Typer) TypecheckFile(file *ast.File, globalScope *Scope) {
	scope := NewScope(globalScope)
	p.typecheckStatements(file, scope)
}

func (p *Typer) TypecheckAndFinalize(files []*ast.File) {
	globalScope := NewScope(nil)

	// Get all global/top-level identifiers
	for _, file := range files {
		scope := globalScope
		for _, itNode := range file.ChildNodes {
			switch node := itNode.(type) {
			case *ast.DeclareStatement,
				*ast.OpStatement,
				*ast.ArrayAppendStatement,
				*ast.Expression,
				*ast.If,
				*ast.For,
				*ast.Block,
				*ast.HTMLBlock,
				*ast.Call,
				*ast.Return:
				// no-op, these are checked in TypecheckFile()
			case *ast.WorkspaceDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				p.typecheckWorkspaceDefinition(node)
			case *ast.ProcedureDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				p.typecheckProcedureDefinition(node, scope)
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
				existingNode, ok := scope.structDefinitions[name]
				if ok {
					errorMessage := fmt.Errorf("Cannot declare \"%s :: struct\" more than once in global scope.", name)
					p.AddError(existingNode.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}

				p.typecheckStruct(node, scope)

				scope.structDefinitions[name] = node
				typeinfo := p.typeinfo.NewStructInfo(node)
				p.typeinfo.register(name, typeinfo)
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
				existingNode, ok := scope.htmlDefinitions[name]
				if ok {
					errorMessage := fmt.Errorf("Cannot declare \"%s :: html\" more than once in global scope.", name)
					p.AddError(existingNode.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}
				scope.htmlDefinitions[name] = node
			case *ast.CSSDefinition:
				if node == nil {
					p.PanicMessage(fmt.Errorf("Found nil top-level %T.", node))
					continue
				}
				if node.Name.Kind == token.Unknown {
					continue
				}
				name := node.Name.String()
				existingNode, ok := scope.cssDefinitions[name]
				if ok {
					errorMessage := fmt.Errorf("Cannot declare \"%s :: css\" more than once in global scope.", name)
					p.AddError(existingNode.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}
				scope.cssDefinitions[name] = node
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
				existingNode, ok := scope.cssConfigDefinitions[name]
				if ok {
					errorMessage := fmt.Errorf("Cannot declare \"%s :: css_config\" more than once in global scope.", name)
					p.AddError(existingNode.Name, errorMessage)
					p.AddError(node.Name, errorMessage)
					continue
				}
				scope.cssConfigDefinitions[name] = node
			default:
				panic(fmt.Sprintf("TypecheckAndFinalize: Unknown type %T", node))
			}
		}
	}

	//
	for _, htmlDefinition := range globalScope.htmlDefinitions {
		p.typecheckHTMLDefinition(htmlDefinition, globalScope)
	}

	// Check if CSS config matches a HTML or CSS component. If not, throw error.
	for name, cssConfigDefinition := range globalScope.cssConfigDefinitions {
		_, ok := globalScope.GetCSSDefinition(name)
		if ok {
			continue
		}
		_, ok = globalScope.GetHTMLDefinition(name)
		if ok {
			continue
		}
		p.AddError(cssConfigDefinition.Name, fmt.Errorf("\"%s :: css_config\" has no matching \":: css\" or \":: html\" block.", name))
	}

	// Get nested dependencies
	for _, htmlDefinition := range globalScope.htmlDefinitions {
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
	for _, htmlDefinition := range globalScope.htmlDefinitions {
		name := htmlDefinition.Name.String()
		node, ok := htmlDefinition.Dependencies[name]
		if !ok {
			continue
		}
		p.AddError(node.Name, fmt.Errorf("Cannot use \"%s\". Cyclic references are not allowed.", name))
	}

	// Typecheck
	for _, file := range files {
		p.TypecheckFile(file, globalScope)
	}
}
