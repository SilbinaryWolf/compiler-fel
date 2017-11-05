package parser

import (
	"fmt"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
	"github.com/silbinarywolf/compiler-fel/util"
)

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

func (p *Parser) typecheckExpression(scope *Scope, expression *ast.Expression) {
	var typeInfo types.TypeInfo
	defer func() {
		expression.TypeInfo = typeInfo
	}()

	if t := expression.TypeIdentifier; t.Kind != token.Unknown {
		typeIdentString := t.String()
		typeInfo = types.GetTypeFromString(typeIdentString)
		if types.HasNoType(typeInfo) {
			p.addErrorToken(fmt.Errorf("Undeclared type %s", typeIdentString), t)
			return
		}
		/*switch typeTokenString {
		case "string":
			exprType = data.KindString
		case "int", "int64":
			exprType = data.KindInteger64
		case "float", "float64":
			exprType = data.KindFloat64
		case "html_node":
			exprType = data.KindHTMLNode
		default:
			p.addErrorToken(fmt.Errorf("Unknown data type %s", typeTokenString), typeToken)
			panic(fmt.Sprintf("typecheckExpression: TODO: Handle explicit type decl. Unknown type \"%s\"", typeTokenString))
		}*/
	}

	for _, itNode := range expression.Nodes() {
		switch node := itNode.(type) {
		case *ast.Token:
			switch node.Kind {
			case token.String:
				expectedTypeInfo := types.String()
				if types.HasNoType(typeInfo) {
					typeInfo = expectedTypeInfo
				}
				if !types.Equals(typeInfo, expectedTypeInfo) {
					p.addErrorToken(fmt.Errorf("Cannot mix %s \"%s\" with %s", expectedTypeInfo.Name(), node.String(), typeInfo.Name()), node.Token)
				}
			case token.Number:
				if types.HasNoType(typeInfo) {
					typeInfo = types.Int()
					if strings.ContainsRune(node.Data, '.') {
						//exprType = types.Float()
						panic("todo(Jake): Fix float")
					}
				}
				if !types.Equals(typeInfo, types.Int()) && !types.Equals(typeInfo, types.Float()) {
					p.addErrorToken(fmt.Errorf("Cannot mix number \"%s\" with %s", node.String(), typeInfo.Name()), node.Token)
				}
			case token.KeywordTrue, token.KeywordFalse:
				expectedTypeInfo := types.Bool()
				if types.HasNoType(typeInfo) {
					typeInfo = expectedTypeInfo
				}
				if !types.Equals(typeInfo, expectedTypeInfo) {
					p.addErrorToken(fmt.Errorf("Cannot mix %s \"%s\" with %s", expectedTypeInfo.Name(), node.String(), typeInfo.Name()), node.Token)
				}
			case token.Identifier:
				name := node.String()
				variableTypeInfo, ok := scope.Get(name)
				if !ok {
					_, ok := scope.GetHTMLDefinition(name)
					if ok {
						p.addErrorToken(fmt.Errorf("Undeclared identifier \"%s\". Did you mean \"%s()\" or \"%s{ }\" to reference the \"%s :: html\" component?", name, name, name, name), node.Token)
						continue
					}
					p.addErrorToken(fmt.Errorf("Undeclared identifier \"%s\".", name), node.Token)
					continue
				}
				if types.HasNoType(typeInfo) {
					typeInfo = variableTypeInfo
				}
				if !types.Equals(typeInfo, variableTypeInfo) {
					p.addErrorToken(fmt.Errorf("Identifier \"%s\" must be a %s not %s.", name, typeInfo.Name(), variableTypeInfo.Name()), node.Token)
				}
			default:
				if node.IsOperator() {
					continue
				}
				panic(fmt.Sprintf("typecheckExpression: Unhandled token kind: \"%s\" with value: %s", node.Kind.String(), node.String()))
			}
		case *ast.HTMLBlock:
			panic("todo(Jake): Fix")
			/*variableType := data.KindHTMLNode
			if exprType == data.KindUnknown {
				exprType = variableType
			}
			if exprType != variableType {
				p.addErrorToken(fmt.Errorf("\":: html\" must be a %s not %s.", exprType.String(), variableType.String()), node.HTMLKeyword)
			}
			p.typecheckHTMLBlock(node, scope)*/
		default:
			panic(fmt.Sprintf("typecheckExpression: Unhandled type %T", node))
		}
	}
}

func (p *Parser) typecheckHTMLBlock(htmlBlock *ast.HTMLBlock, scope *Scope) {
	scope = NewScope(scope)
	p.typecheckStatements(htmlBlock, scope)
}

func (p *Parser) typecheckHTMLDefinition(htmlDefinition *ast.HTMLComponentDefinition, parentScope *Scope) {
	// Attach CSSDefinition if found
	name := htmlDefinition.Name.String()
	cssDefinition, ok := parentScope.GetCSSDefinition(name)
	if ok {
		htmlDefinition.CSSDefinition = cssDefinition
	}

	// Attach CSSConfigDefinition if found
	cssConfigDefinition, ok := parentScope.GetCSSConfigDefinition(name)
	if ok {
		htmlDefinition.CSSConfigDefinition = cssConfigDefinition
	}

	//
	var globalScopeNoVariables Scope = *parentScope
	globalScopeNoVariables.identifiers = nil
	scope := NewScope(&globalScopeNoVariables)
	scope.Set("children", types.HTML())

	if htmlDefinition.Properties != nil {
		for i, _ := range htmlDefinition.Properties.Statements {
			var propertyNode *ast.DeclareStatement = htmlDefinition.Properties.Statements[i]
			p.typecheckExpression(scope, &propertyNode.Expression)
			name := propertyNode.Name.String()
			_, ok := scope.Get(name)
			if ok {
				if name == "children" {
					p.addErrorToken(fmt.Errorf("Cannot use \"children\" as it's a reserved property."), propertyNode.Name)
					continue
				}
				p.addErrorToken(fmt.Errorf("Property \"%s\" declared twice.", name), propertyNode.Name)
				continue
			}
			scope.Set(name, propertyNode.TypeInfo)
		}
	}

	if p.typecheckHtmlNodeDependencies != nil {
		panic("typecheckHtmlNodeDependencies must be nil before being re-assigned")
	}
	p.typecheckHtmlNodeDependencies = make(map[string]*ast.HTMLNode)
	p.typecheckStatements(htmlDefinition, scope)
	htmlDefinition.Dependencies = p.typecheckHtmlNodeDependencies
	p.typecheckHtmlNodeDependencies = nil
}

func (p *Parser) typecheckStatements(topNode ast.Node, scope *Scope) {
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

		if itNode == nil {
			scope = scope.parent
			continue
		}

		switch node := itNode.(type) {
		case *ast.CSSDefinition,
			*ast.CSSConfigDefinition,
			*ast.HTMLComponentDefinition,
			*ast.HTMLProperties:
			// Skip nodes and child nodes
			continue
		case *ast.HTMLBlock:
			p.typecheckHTMLBlock(node, scope)
		case *ast.HTMLNode:
			for i, _ := range node.Parameters {
				p.typecheckExpression(scope, &node.Parameters[i].Expression)
			}

			name := node.Name.String()
			isValidHTML5TagName := util.IsValidHTML5TagName(name)
			if !isValidHTML5TagName {
				htmlComponentDefinition, ok := scope.GetHTMLDefinition(name)
				if !ok {
					p.addErrorToken(fmt.Errorf("\"%s\" is not a valid element or component name.", name), node.Name)
					continue
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
				// Check if parameters exist
			ParameterCheckLoop:
				for i, _ := range node.Parameters {
					parameterNode := &node.Parameters[i]
					paramName := parameterNode.Name.String()
					for _, componentParamNode := range node.HTMLDefinition.Properties.Statements {
						if paramName == componentParamNode.Name.String() {
							parameterType := parameterNode.TypeInfo
							componentStructType := componentParamNode.TypeInfo
							if parameterType != componentStructType {
								p.addErrorToken(fmt.Errorf("\"%s\" must be of type %s, not %s", paramName, componentStructType.Name(), parameterType.Name()), parameterNode.Name)
							}
							continue ParameterCheckLoop
						}
					}
					p.addErrorToken(fmt.Errorf("\"%s\" is not a property on \"%s :: html\"", paramName, name), parameterNode.Name)
					continue
				}
			}

		case *ast.DeclareStatement:
			p.typecheckExpression(scope, &node.Expression)
			name := node.Name.String()
			_, ok := scope.GetFromThisScope(name)
			if ok {
				p.addErrorToken(fmt.Errorf("Cannot redeclare \"%s\".", name), node.Name)
				continue
			}
			scope.Set(name, node.Expression.TypeInfo)
			continue
		case *ast.Expression:
			p.typecheckExpression(scope, node)
			continue
		default:
			panic(fmt.Sprintf("TypecheckStatements: Unknown type %T", node))
		}

		// Nest scope
		scope = NewScope(scope)
		nodeStack = append(nodeStack, nil)

		// Add children
		nodes := itNode.Nodes()
		for i := len(nodes) - 1; i >= 0; i-- {
			node := nodes[i]
			nodeStack = append(nodeStack, node)
		}
	}
}

func (p *Parser) TypecheckFile(file *ast.File, globalScope *Scope) {
	scope := NewScope(globalScope)
	p.typecheckStatements(file, scope)
}

func (p *Parser) TypecheckAndFinalize(files []*ast.File) {
	globalScope := NewScope(nil)

	// Get all global/top-level identifiers
	for _, file := range files {
		scope := globalScope
		for _, itNode := range file.ChildNodes {
			switch node := itNode.(type) {
			case *ast.HTMLNode, *ast.DeclareStatement, *ast.Expression, *ast.HTMLBlock:
				// no-op, these are checked in TypecheckFile()
			case *ast.HTMLComponentDefinition:
				if node == nil {
					continue
				}
				if node.Name.Kind == token.Unknown {
					p.addErrorToken(fmt.Errorf("Cannot declare anonymous \":: html\" block."), node.Name)
					continue
				}
				name := node.Name.String()
				existingNode, ok := scope.htmlDefinitions[name]
				if ok {
					errorMessage := fmt.Errorf("Cannot declare \"%s :: html\" more than once in global scope.", name)
					p.addErrorToken(errorMessage, existingNode.Name)
					p.addErrorToken(errorMessage, node.Name)
					continue
				}
				scope.htmlDefinitions[name] = node
			case *ast.CSSDefinition:
				if node == nil {
					continue
				}
				if node.Name.Kind == token.Unknown {
					continue
				}
				name := node.Name.String()
				existingNode, ok := scope.cssDefinitions[name]
				if ok {
					errorMessage := fmt.Errorf("Cannot declare \"%s :: css\" more than once in global scope.", name)
					p.addErrorToken(errorMessage, existingNode.Name)
					p.addErrorToken(errorMessage, node.Name)
					continue
				}
				scope.cssDefinitions[name] = node
			case *ast.CSSConfigDefinition:
				if node == nil {
					continue
				}
				if node.Name.Kind == token.Unknown {
					p.addErrorToken(fmt.Errorf("Cannot declare anonymous \":: css_config\" block."), node.Name)
					continue
				}
				name := node.Name.String()
				existingNode, ok := scope.cssConfigDefinitions[name]
				if ok {
					errorMessage := fmt.Errorf("Cannot declare \"%s :: css_config\" more than once in global scope.", name)
					p.addErrorToken(errorMessage, existingNode.Name)
					p.addErrorToken(errorMessage, node.Name)
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
		p.addErrorToken(fmt.Errorf("\"%s :: css_config\" has no matching \":: css\" or \":: html\" block.", name), cssConfigDefinition.Name)
	}

	// Get nested dependencies
	for _, htmlDefinition := range globalScope.htmlDefinitions {
		nodeStack := make([]*ast.HTMLNode, 0, 50)
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
		p.addErrorToken(fmt.Errorf("Cannot use \"%s\". Cyclic references are not allowed.", name), node.Name)
	}

	// Typecheck
	for _, file := range files {
		p.TypecheckFile(file, globalScope)
	}
}
