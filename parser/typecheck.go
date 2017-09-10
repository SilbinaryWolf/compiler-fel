package parser

import (
	"fmt"
	//"io/ioutil"
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
		case "int":
			exprType = data.KindInteger64
		case "float":
			exprType = data.KindFloat64
		default:
			p.addErrorLine(fmt.Errorf("Unknown data type %s", typeTokenString), typeToken.Line)
			panic(fmt.Sprintf("todo: Handle explicit type decl. Unknown type %s", typeTokenString))
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

func (p *Parser) TypecheckAndFinalize(files []*ast.File) {
	globalScope := NewScope(nil)

	// Get all global/top-level identifiers
	for _, file := range files {
		scope := globalScope
		for _, itNode := range file.ChildNodes {
			switch node := itNode.(type) {
			case *ast.HTMLNode:
				// no-op
			case *ast.HTMLComponentDefinition:
				name := node.Name.String()
				_, ok := scope.htmlDefinitions[name]
				if ok {
					p.addError(fmt.Errorf("Cannot declare \"%s :: html\" more than once.", name))
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
					p.addError(fmt.Errorf("Cannot declare \"%s :: css\" more than once.", name))
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
		scope := NewScope(globalScope)

		nodes := file.Nodes()
		for i := len(nodes) - 1; i >= 0; i-- {
			node := nodes[i]
			nodeStack = append(nodeStack, node)
		}

		for len(nodeStack) > 0 {
			itNode := nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]

			if itNode == nil {
				scope = scope.parent
				continue
			}

			switch node := itNode.(type) {
			case *ast.CSSDefinition, *ast.HTMLComponentDefinition:
				// Skip nodes and child nodes
				continue
			case *ast.HTMLNode:
				for _, parameterNode := range node.Parameters {
					p.typecheckExpression(scope, &parameterNode.Expression)
				}

				name := node.Name.String()
				isValidHTML5TagName := util.IsValidHTML5TagName(name)
				if !isValidHTML5TagName {
					htmlDefinition, ok := scope.Get(name)
					if !ok {
						p.addErrorLine(fmt.Errorf("\"%s\" is not a valid HTML5 element or HTML component", name), node.Name.Line)
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
				panic(fmt.Sprintf("Unknown type %T", node))
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
}
