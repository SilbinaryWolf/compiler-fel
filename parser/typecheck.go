package parser

import (
	"fmt"
	//"io/ioutil"
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
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
		panic("todo: Handle explicit type")
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
				// todo(Jake): Support float
				if exprType == data.KindUnknown {
					exprType = data.KindInteger64
				}
				if exprType != data.KindInteger64 && exprType != data.KindFloat64 {
					p.addErrorLine(fmt.Errorf("Cannot mix integer (\"%s\") with %s", node.String(), exprType.String()), node.Line)
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

func (p *Parser) TypecheckAndFinalize(files []*ast.File) {
	scope := NewScope(nil)

	// Get all global/top-level identifiers
	for _, file := range files {
		for _, itNode := range file.ChildNodes {
			switch node := itNode.(type) {
			case *ast.HTMLNode:
				// no-cop
			case *ast.HTMLComponentDefinition:
				name := node.Name.String()
				_, ok := scope.htmlDefinitions[name]
				if ok {
					p.addError(fmt.Errorf("Cannot declare \"%s :: HTML\" more than once.", name))
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
					p.addError(fmt.Errorf("Cannot declare \"%s :: CSS\" more than once.", name))
					continue
				}
				scope.cssDefinitions[name] = node
			default:
				panic(fmt.Sprintf("Unknown type %T", node))
			}
		}
	}

	// Typecheck
	for _, file := range files {
		nodeStack := make([]ast.Node, 0, 50)
		scopeStack := make([]*Scope, 0, 50)
		scopeStack = append(scopeStack, scope)

		nodes := file.Nodes()
		for i := len(nodes) - 1; i >= 0; i-- {
			node := nodes[i]
			nodeStack = append(nodeStack, node)
		}

		for len(nodeStack) > 0 {
			itNode := nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]

			if itNode == nil {
				scopeStack = scopeStack[:len(scopeStack)-1]
				continue
			}

			scope := scopeStack[len(scopeStack)-1]

			switch node := itNode.(type) {
			case *ast.CSSDefinition, *ast.HTMLComponentDefinition:
				// Skip nodes and child nodes
				continue
			case *ast.HTMLNode:
				for _, parameterNode := range node.Parameters {
					p.typecheckExpression(scope, &parameterNode.Expression)
				}
			case *ast.DeclareStatement:
				p.typecheckExpression(scope, &node.Expression)
				name := node.Name.String()
				_, ok := scope.Get(name)
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
				panic(fmt.Sprintf("Unknown type %T", node))
			}

			// Nest scope
			scopeStack = append(scopeStack, NewScope(scope))
			nodeStack = append(nodeStack, nil)

			// Add children
			nodes := itNode.Nodes()
			for i := len(nodes) - 1; i >= 0; i-- {
				node := nodes[i]
				nodeStack = append(nodeStack, node)
			}
		}
	}
}
