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
				selectorList = append(selectorList, &data.CSSSelectorIdentifier{
					Name: selectorPartNode.String(),
				})
			case token.Declare:
				selectorList = append(selectorList, &data.CSSSelectorOperator{
					Operator: ":",
				})
			case token.Define:
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

func (program *Program) evaluateCSSDefinition(topNode *ast.CSSDefinition, scope *Scope) {
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
			ruleNode.Selectors = selectorRuleList
			ruleNode.Properties = propertyList
			resultList = append(resultList, ruleNode)
			//{
			//	json, _ := json.MarshalIndent(ruleNode, "", "   ")
			//	fmt.Printf("%s", string(json))
			//}
			//panic("evaluateCSSDefinition(): finish function")
		default:
			{
				json, _ := json.MarshalIndent(node, "", "   ")
				fmt.Printf("%s", string(json))
			}
			panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled type: %T", node))
		}
	}
	panic(fmt.Sprintf("evaluateCSSDefinition(): Finish handling CSS statement"))
}
