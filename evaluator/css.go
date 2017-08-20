package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (program *Program) evaluateSelector(nodes []ast.Node) data.CSSSelector {
	selectorList := make(data.CSSSelector, 0, len(nodes))
	for _, itSelectorPartNode := range nodes {
		//var value string
		switch selectorPartNode := itSelectorPartNode.(type) {
		case *ast.Token:
			switch selectorPartNode.Kind {
			case token.Identifier:
				name := selectorPartNode.String()
				selectorList = append(selectorList, &data.CSSSelectorIdentifier{
					Name: name,
				})
			case token.Colon:
				selectorList = append(selectorList, &data.CSSSelectorOperator{
					Operator: ":",
				})
			case token.DoubleColon:
				selectorList = append(selectorList, &data.CSSSelectorOperator{
					Operator: "::",
				})
			default:
				if selectorPartNode.IsOperator() {
					selectorList = append(selectorList, &data.CSSSelectorOperator{
						Operator: selectorPartNode.String(),
					})
					continue
				}
				panic(fmt.Sprintf("evaluateSelector(): Unhandled selector sub-node kind: %s", selectorPartNode.Kind.String()))
			}
		case *ast.CSSSelector:
			subSelectorList := program.evaluateSelector(selectorPartNode.Nodes())
			selectorList = append(selectorList, subSelectorList)
			//for _, token := range selectorPartNode.ChildNodes {
			//	value += token.String() + " "
			//}
			//value = value[:len(value)-1]
		case *ast.CSSAttributeSelector:
			if selectorPartNode.Operator.Kind != 0 {
				value := &data.CSSSelectorAttribute{
					Name:     selectorPartNode.Name.String(),
					Operator: selectorPartNode.Operator.String(),
					Value:    selectorPartNode.Value.String(),
				}
				selectorList = append(selectorList, value)
				break
			}
			value := &data.CSSSelectorAttribute{
				Name: selectorPartNode.Name.String(),
			}
			selectorList = append(selectorList, value)
			//value = fmt.Sprintf("[%s]", selectorPartNode.Name)
			//panic("evaluateSelector(): Handle attribute selector")
		default:
			panic(fmt.Sprintf("evaluateSelector(): Unhandled selector node type: %T", selectorPartNode))
		}
	}
	return selectorList
}

func (program *Program) evaluateCSSDefinition(topNode *ast.CSSDefinition, scope *Scope) *data.CSSDefinition {
	resultList := make([]*data.CSSRule, 0, 10)
	scope = NewScope(scope)
	for _, itNode := range topNode.Nodes() {
		switch node := itNode.(type) {
		case *ast.DeclareStatement:
			program.evaluateDeclareSet(node, scope)
		case *ast.CSSRule:
			// Evaluate selectors
			selectorRuleList := make([]data.CSSSelector, 0, 10)
			for _, selectorListNode := range node.Selectors {
				selectorList := program.evaluateSelector(selectorListNode.Nodes())
				selectorRuleList = append(selectorRuleList, selectorList)
			}

			// Evaluate child nodes / properties
			propertyList := make([]data.CSSProperty, 0, 10)
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
							switch node.Kind {
							case token.Identifier:
								name := node.String()

								// If a variable is declared with this name, use it instead.
								value, ok := scope.Get(name)
								if ok {
									fmt.Printf("%v\n", value)
									panic("todo(jake): Make it use this variable value")
									//name = value
								}
							default:
								value.WriteString(node.String())
							}
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
			ruleNode.Selectors = selectorRuleList
			ruleNode.Properties = propertyList
			resultList = append(resultList, ruleNode)
		default:
			{
				json, _ := json.MarshalIndent(node, "", "   ")
				fmt.Printf("%s", string(json))
			}
			panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled type: %T", node))
		}
	}

	cssDefinition := new(data.CSSDefinition)
	cssDefinition.Name = topNode.Name.String()
	cssDefinition.ChildNodes = resultList
	if len(cssDefinition.Name) == 0 {
		panic("evaluateCSSDefinition(): Todo! Cannot have named CSS blocks yet.")
	}
	if scope == nil {
		panic("evaluateCSSDefinition(): Null scope provided.")
	}
	/*if scope.parent != nil {
		{
			json, _ := json.MarshalIndent(scope.parent, "", "   ")
			fmt.Printf("%s", string(json))
		}
		panic("evaluateCSSDefinition(): Todo! Can only have CSS blocks at top-level")
	}*/
	program.globalScope.cssDefinitions = append(program.globalScope.cssDefinitions, cssDefinition)
	return cssDefinition
}
