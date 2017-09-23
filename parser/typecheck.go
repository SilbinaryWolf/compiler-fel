package parser

import (
	"fmt"
	//"io/ioutil"
	//"encoding/json"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
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
	var exprType data.Kind

	typeToken := expression.TypeToken
	if typeToken.Kind != token.Unknown {
		typeTokenString := typeToken.String()
		switch typeTokenString {
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
		}
	}

	for _, itNode := range expression.Nodes() {
		switch node := itNode.(type) {
		case *ast.Token:
			switch node.Kind {
			case token.String:
				if exprType == data.KindUnknown {
					exprType = data.KindString
				}
				if exprType != data.KindString {
					p.addErrorToken(fmt.Errorf("Cannot mix string \"%s\" with %s", node.String(), exprType.String()), node.Token)
				}
			case token.Number:
				if exprType == data.KindUnknown {
					exprType = data.KindInteger64
					if strings.ContainsRune(node.Data, '.') {
						exprType = data.KindFloat64
					}
				}
				if exprType != data.KindInteger64 && exprType != data.KindFloat64 {
					p.addErrorToken(fmt.Errorf("Cannot mix number (\"%s\") with %s", node.String(), exprType.String()), node.Token)
				}
			case token.Identifier:
				name := node.String()
				variableType, ok := scope.Get(name)
				if !ok {
					_, ok := scope.GetHTMLDefinition(name)
					if ok {
						p.addErrorToken(fmt.Errorf("Undeclared identifier \"%s\". Did you mean \"%s()\" or \"%s{ }\" to reference the \"%s :: html\" component?", name, name, name), node.Token)
						continue
					}
					p.addErrorToken(fmt.Errorf("Undeclared identifier \"%s\".", name), node.Token)
					continue
				}
				if exprType == data.KindUnknown {
					exprType = variableType
				}
				if exprType != variableType {
					p.addErrorToken(fmt.Errorf("Identifier \"%s\" must be a %s not %s.", name, exprType.String(), variableType.String()), node.Token)
				}
			default:
				if node.IsOperator() {
					continue
				}
				panic(fmt.Sprintf("typecheckExpression: Unhandled token kind: \"%s\" with value: %s", node.Kind.String(), node.String()))
			}
		case *ast.HTMLBlock:
			variableType := data.KindHTMLNode
			if exprType == data.KindUnknown {
				exprType = variableType
			}
			if exprType != variableType {
				p.addErrorToken(fmt.Errorf("\":: html\" must be a %s not %s.", exprType.String(), variableType.String()), node.HTMLKeyword)
			}
			p.TypecheckHTMLBlock(node, scope)
		default:
			panic(fmt.Sprintf("typecheckExpression: Unhandled type %T", node))
		}
	}

	expression.Type = exprType
}

func (p *Parser) TypecheckHTMLBlock(htmlBlock *ast.HTMLBlock, scope *Scope) {
	scope = NewScope(scope)
	p.TypecheckStatements(htmlBlock, scope)
}

func (p *Parser) TypecheckHTMLDefinition(htmlDefinition *ast.HTMLComponentDefinition, parentScope *Scope) {
	// Attach CSSDefinition if found
	name := htmlDefinition.Name.String()
	cssDefinition, ok := parentScope.GetCSSDefinition(name)
	if ok {
		htmlDefinition.CSSDefinition = cssDefinition
	}

	//
	var globalScopeNoVariables Scope = *parentScope
	globalScopeNoVariables.identifiers = nil
	scope := NewScope(&globalScopeNoVariables)
	scope.Set("children", data.KindHTMLNode)

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
			scope.Set(name, propertyNode.Type)
		}
	}

	if p.typecheckHtmlNodeDependencies != nil {
		panic("typecheckHtmlNodeDependencies must be nil before being re-assigned")
	}
	p.typecheckHtmlNodeDependencies = make(map[string]*ast.HTMLNode)
	p.TypecheckStatements(htmlDefinition, scope)
	htmlDefinition.Dependencies = p.typecheckHtmlNodeDependencies
	p.typecheckHtmlNodeDependencies = nil
}

func (p *Parser) TypecheckStatements(topNode ast.Node, scope *Scope) {
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
		case *ast.CSSDefinition, *ast.HTMLComponentDefinition, *ast.HTMLProperties:
			// Skip nodes and child nodes
			continue
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
							if componentParamNode.Type != parameterNode.Type {
								p.addErrorToken(fmt.Errorf("\"%s\" must be of type %s, not %s", paramName, componentParamNode.Type.String(), parameterNode.Type.String()), parameterNode.Name)
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
			scope.Set(name, node.Expression.Type)
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
	p.TypecheckStatements(file, scope)
}

func (p *Parser) TypecheckAndFinalize(files []*ast.File) {
	globalScope := NewScope(nil)

	// Get all global/top-level identifiers
	for _, file := range files {
		scope := globalScope
		for _, itNode := range file.ChildNodes {
			switch node := itNode.(type) {
			case *ast.HTMLNode, *ast.DeclareStatement, *ast.Expression:
				// no-op, these are checked in TypecheckFile()
			case *ast.HTMLComponentDefinition:
				name := node.Name.String()
				_, ok := scope.htmlDefinitions[name]
				if ok {
					p.addError(fmt.Errorf("Cannot declare \"%s :: html\" more than once in global scope.", name))
					continue
				}
				scope.htmlDefinitions[name] = node
			case *ast.CSSDefinition:
				if node.Name.Kind == token.Unknown {
					continue
				}
				name := node.Name.String()
				_, ok := scope.cssDefinitions[name]
				if ok {
					p.addError(fmt.Errorf("Cannot declare \"%s :: css\" more than once in global scope.", name))
					continue
				}
				scope.cssDefinitions[name] = node
			default:
				panic(fmt.Sprintf("TypecheckAndFinalize: Unknown type %T", node))
			}
		}
	}

	//
	for _, htmlDefinition := range globalScope.htmlDefinitions {
		p.TypecheckHTMLDefinition(htmlDefinition, globalScope)
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
