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

func getDataTypeFromToken(t token.Token) data.Kind {
	switch t.Kind {
	case token.Identifier:
		typename := t.String()
		switch typename {
		case "string":
			return data.KindString
		default:
			panic(fmt.Sprintf("Unknown type name: %s", typename))
		}
	default:
		panic(fmt.Sprintf("Cannot use token kind %s in type declaration", t.Kind.String()))
	}
}

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
		default:
			p.addErrorLine(fmt.Errorf("Unknown data type %s", typeTokenString), typeToken.Line)
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
					p.addErrorLine(fmt.Errorf("Cannot mix string \"%s\" with %s", node.String(), exprType.String()), node.Line)
				}
			case token.Number:
				if exprType == data.KindUnknown {
					exprType = data.KindInteger64
					if strings.ContainsRune(node.Data, '.') {
						exprType = data.KindFloat64
					}
				}
				if exprType != data.KindInteger64 && exprType != data.KindFloat64 {
					p.addErrorLine(fmt.Errorf("Cannot mix number (\"%s\") with %s", node.String(), exprType.String()), node.Line)
				}
			case token.Identifier:
				name := node.String()
				variableType, ok := scope.Get(name)
				if !ok {
					p.addErrorLine(fmt.Errorf("Undeclared identifier \"%s\".", name), node.Line)
					continue
				}
				if exprType == data.KindUnknown {
					exprType = variableType
				}
				if exprType != variableType {
					p.addErrorLine(fmt.Errorf("Identifier \"%s\" must be a %s not %s.", name, exprType.String(), variableType.String()), node.Line)
				}
			default:
				if node.IsOperator() {
					continue
				}
				panic(fmt.Sprintf("typecheckExpression: Unhandled token kind: \"%s\" with value: %s", node.Kind.String(), node.String()))
			}
		default:
			panic(fmt.Sprintf("typecheckExpression: Unhandled type %T", node))
		}
	}

	expression.Type = exprType
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
					p.addErrorLine(fmt.Errorf("Cannot use \"children\" as it's a reserved property."), propertyNode.Name.Line)
					continue
				}
				p.addErrorLine(fmt.Errorf("Property \"%s\" declared twice.", name), propertyNode.Name.Line)
				continue
			}
			scope.Set(name, propertyNode.Type)
		}
	}

	p.typecheckHtmlNodeDependencies = make(map[string]*ast.HTMLNode)
	p.TypecheckStatements(htmlDefinition, scope)

	// Put list of dependencies on the AST
	dependencies := make([]*ast.HTMLNode, 0, len(p.typecheckHtmlNodeDependencies))
	for _, htmlNode := range p.typecheckHtmlNodeDependencies {
		dependencies = append(dependencies, htmlNode)
		if htmlDefinition == htmlNode.HTMLDefinition {
			p.addError(fmt.Errorf("Cannot reference self in \"%s :: html\".", htmlDefinition.Name.String()))
		}
	}
	htmlDefinition.Dependencies = dependencies
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
					p.addErrorLine(fmt.Errorf("\"%s\" is not a valid HTML5 element or HTML component", name), node.Name.Line)
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
								p.addErrorLine(fmt.Errorf("\"%s\" must be of type %s, not %s", paramName, componentParamNode.Type.String(), parameterNode.Type.String()), parameterNode.Name.Line)
							}
							continue ParameterCheckLoop
						}
					}
					p.addErrorLine(fmt.Errorf("\"%s\" is not a property on \"%s :: html\"", paramName, name), parameterNode.Name.Line)
					continue
				}
			}

		case *ast.DeclareStatement:
			p.typecheckExpression(scope, &node.Expression)
			name := node.Name.String()
			_, ok := scope.GetFromThisScope(name)
			if ok {
				p.addErrorLine(fmt.Errorf("Cannot redeclare \"%s\".", name), node.Name.Line)
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
			case *ast.HTMLNode, *ast.DeclareStatement:
				// no-op
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

	// Check for circular dependencies in components to avoid recursion at runtime
CircularDependencyLoopCheck:
	for _, htmlDefinition := range globalScope.htmlDefinitions {
		hasChecked := new(map[string]bool)
		for _, otherHtmlDefinition := range globalScope.htmlDefinitions {
			// Don't compare dependencies of self
			if htmlDefinition == otherHtmlDefinition {
				continue
			}
			for _, outerDepHtmlNode := range otherHtmlDefinition.Dependencies {
				if htmlDefinition != outerDepHtmlNode.HTMLDefinition {
					continue
				}
				for _, innerDepHtmlNode := range htmlDefinition.Dependencies {
					if otherHtmlDefinition != innerDepHtmlNode.HTMLDefinition {
						continue
					}
					p.addErrorLine(fmt.Errorf("Cannot use \"%s\" in circular reference.", outerDepHtmlNode.Name.String()), outerDepHtmlNode.Name.Line)
					p.addErrorLine(fmt.Errorf("Cannot use \"%s\" in circular reference.", innerDepHtmlNode.Name.String()), innerDepHtmlNode.Name.Line)
					break CircularDependencyLoopCheck
				}
			}
		}
	}

	// Typecheck
	for _, file := range files {
		p.TypecheckFile(file, globalScope)
	}
}
