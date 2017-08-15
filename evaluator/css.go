package evaluator

import (
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
					default:
						panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled selector node type: %T", selectorPartNode))
					}
					selector.Tokens = append(selector.Tokens, value)
				}
				selectorList = append(selectorList, selector)
			}

			// Evaluate child nodes / properties
			for _, itNode := range node.Nodes() {
				switch itNode.(type) {
				default:
					panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled child node type: %T", itNode))
				}
				//property := data.CSSProperty{
				//	Name: itNode.Name.String(),
				//}
			}

			ruleNode := new(data.CSSRule)
			ruleNode.Selectors = selectorList
			{
				json, _ := json.MarshalIndent(ruleNode, "", "   ")
				fmt.Printf("%s", string(json))
			}
			panic("evaluateCSSDefinition(): finish function")
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
