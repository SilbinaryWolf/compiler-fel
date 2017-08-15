package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

func (program *Program) evaluateCSSDefinition(topNode *ast.CSSDefinition, scope *Scope) {
	scope = NewScope(scope)
	for _, itNode := range topNode.Nodes() {
		switch node := itNode.(type) {
		case *ast.DeclareStatement:
			program.evaluateDeclareSet(node, scope)
		case *ast.CSSRule:
			// Evaluate selectors
			selectorList := make([]data.CSSSelector, 0, 5)
			for _, selectorListNode := range node.Selectors {
				selector := data.CSSSelector{}
				for _, itSelectorPartNode := range selectorListNode.Nodes() {
					var value string
					switch selectorPartNode := itSelectorPartNode.(type) {
					case *ast.Token:
						value = selectorPartNode.String()
					case *ast.CSSSelector:
						{
							json, _ := json.MarshalIndent(selectorPartNode, "", "   ")
							fmt.Printf("%s", string(json))
							panic("Tests")
						}
						//for _, token := range selectorPartNode.ChildNodes {
						//	value += token.String() + " "
						//}
						value = value[:len(value)-1]
					case *ast.CSSAttributeSelector:
						if selectorPartNode.Operator.Kind != 0 {
							value = fmt.Sprintf("[%s%s%s]", selectorPartNode.Name, selectorPartNode.Operator, selectorPartNode.Value)
							break
						}
						value = fmt.Sprintf("[%s]", selectorPartNode.Name)
					default:
						panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled selector node type: %T", selectorPartNode))
					}
					selector.Tokens = append(selector.Tokens, value)
				}
				selectorList = append(selectorList, selector)
			}

			// Evaluate child nodes / properties
			propertyList := make([]data.CSSProperty, 0, 5)
			for _, itNode := range node.Nodes() {
				switch node := itNode.(type) {
				case *ast.CSSProperty:
					property := data.CSSProperty{
						Name: node.Name.String(),
					}

					var value bytes.Buffer
					for _, itNode := range node.ChildNodes {
						switch node := itNode.(type) {
						case *ast.Token:
							value.WriteString(node.String())
							value.WriteByte(' ')
						default:
							panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled CSS property value node type: %T", itNode))
						}
					}

					property.Value = value.String()
					property.Value = property.Value[:len(property.Value)-1]
					propertyList = append(propertyList, property)
				default:
					panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled child node type: %T", itNode))
				}
			}

			ruleNode := new(data.CSSRule)
			ruleNode.Selectors = selectorList
			ruleNode.Properties = propertyList
			//{
			//	json, _ := json.MarshalIndent(ruleNode, "", "   ")
			//	fmt.Printf("%s", string(json))
			//}
			//panic("evaluateCSSDefinition(): finish function")
		default:
			{
				json, _ := json.MarshalIndent(node, "", "   ")
				fmt.Printf("%s", string(json))
				panic("CSSRULE")
			}
			panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled type: %T", node))
		}
	}
	panic(fmt.Sprintf("evaluateCSSDefinition(): Finish handling CSS statement"))
}
