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

func (p *Parser) TypecheckAndFinalize() {
	for _, node := range p.htmlComponentNodes {
		name := node.Name.String()

		// Attach HTML component definition to HTML node
		htmlComponentDef, ok := p.htmlComponentDefinitions[name]
		if !ok {
			p.addErrorLine(fmt.Errorf("\"%s\" is not a valid html tag or defined component.", name), node.Name.Line)
			continue
		}
		node.HTMLDefinition = htmlComponentDef

		// Attach CSS definition to HTML node
		cssDefinition, ok := p.cssComponentDefinitions[name]
		if ok {
			node.CSSDefinition = cssDefinition
		}
	}
	p.htmlComponentNodes = nil

	//
	for _, expressionNode := range p.exprNodes {
		// Read data type
		var dataType data.Kind
		typeToken := expressionNode.TypeToken
		if typeToken.Kind != token.Unknown {
			dataType = getDataTypeFromToken(typeToken)
		}

		// Check if expression contents match
		for _, itNode := range expressionNode.ChildNodes {
			switch node := itNode.(type) {
			case *ast.Token:
				switch node.Kind {
				case token.String:
					if dataType != data.KindUnknown && dataType != data.KindString {
						p.addError(fmt.Errorf("Cannot mix string type with numbers in expression"))
						return
					}
				//case token.Identifier:

				default:
					panic(fmt.Sprintf("Unhandled token kind: %s", node.Kind.String()))
				}
			default:
				panic(fmt.Sprintf("Unhandled expression item type %T", itNode))
			}
		}
	}
	panic("TypecheckAndFinalize: Todo, check typing on expressions")
}
