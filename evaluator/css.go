package evaluator

/*import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (program *Program) evaluateSelector(cssDefinition *data.CSSDefinition, nodes []ast.Node) data.CSSSelector {
	var cssConfigDefinition *ast.CSSConfigDefinition
	if htmlComponentDefinition := program.CurrentComponentScope(); htmlComponentDefinition != nil {
		cssConfigDefinition = htmlComponentDefinition.CSSConfigDefinition
	}

	selectorList := make(data.CSSSelector, 0, len(nodes))
	for _, itSelectorPartNode := range nodes {
		//var value string
		switch selectorPartNode := itSelectorPartNode.(type) {
		case *ast.Token:
			switch selectorPartNode.Kind {
			case token.Identifier:
				name := selectorPartNode.String()

				var selectorKind data.CSSSelectorPartKind
				switch name[0] {
				case '.':
					selectorKind = data.SelectorKindClass

					// Prefix component namespace
					config := cssConfigDefinition.GetSettings(name)
					if config.Modify && len(cssDefinition.Name) > 0 {
						name = "." + PrefixNamespace(cssDefinition.Name, name[1:])
					}
				case '#':
					selectorKind = data.SelectorKindID
				default:
					selectorKind = data.SelectorKindTag
				}

				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: selectorKind,
					Name: name,
				})
			case token.AtKeyword:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindAtKeyword,
					Name: selectorPartNode.String(),
				})
			case token.Number:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindNumber,
					Name: selectorPartNode.String(),
				})
			case token.Colon:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindColon,
				})
			case token.DoubleColon:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindDoubleColon,
				})
			case token.Whitespace:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindAncestor,
				})
			case token.GreaterThan:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindChild,
				})
			case token.Add:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindAdjacent,
				})
			case token.Tilde:
				selectorList = append(selectorList, data.CSSSelectorPart{
					Kind: data.SelectorKindSibling,
				})
			default:
				if selectorPartNode.IsOperator() {
					panic("todo(Jake): Fixme (or add support for operator in above `switch`)")
					// 	selectorPartString := selectorPartNode.String()
					// 	selectorList = append(selectorList, data.CSSSelectorOperator{
					// 		Operator: selectorPartString,
					// 	})
					// 	continue
				}
				panic(fmt.Sprintf("evaluateSelector(): Unhandled selector sub-node kind: %s", selectorPartNode.Kind.String()))
			}
		case *ast.CSSSelector:
			// todo(Jake)
			panic(fmt.Sprintf("todo(Jake): Fix this, %v", selectorPartNode.Nodes()))
			//subSelectorList := program.evaluateSelector(selectorPartNode.Nodes())
			//selectorList = append(selectorList, subSelectorList)

			//for _, token := range selectorPartNode.ChildNodes {
			//	value += token.String() + " "
			//}
			//value = value[:len(value)-1]
		case *ast.CSSAttributeSelector:
			if selectorPartNode.Operator.Kind != 0 {
				value := data.CSSSelectorPart{
					Kind:     data.SelectorKindAttribute,
					Name:     selectorPartNode.Name.String(),
					Operator: selectorPartNode.Operator.String(),
					Value:    selectorPartNode.Value.String(),
				}
				selectorList = append(selectorList, value)
				break
			}
			value := data.CSSSelectorPart{
				Kind: data.SelectorKindAttribute,
				Name: selectorPartNode.Name.String(),
			}
			selectorList = append(selectorList, value)
		default:
			panic(fmt.Sprintf("evaluateSelector(): Unhandled selector node type: %T", selectorPartNode))
		}
	}
	return selectorList
}

func (program *Program) evaluateCSSRule(cssDefinition *data.CSSDefinition, topNode *ast.CSSRule, parentCSSRule *data.CSSRule, scope *Scope) {
	scope = NewScope(scope)

	ruleNode := new(data.CSSRule)
	nextNodeToAppend := ruleNode

	// Evaluate selectors
	ruleNode.Selectors = make([]data.CSSSelector, 0, 10)
	if parentCSSRule != nil {
		switch topNode.Kind {
		case ast.CSSKindRule:
			// Handle nested selectors
			for _, parentSelectorListNode := range parentCSSRule.Selectors {
				for _, selectorListNode := range topNode.Selectors {
					selectorList := make(data.CSSSelector, 0, len(parentSelectorListNode))
					selectorList = append(selectorList, parentSelectorListNode...)
					selectorList = append(selectorList, program.evaluateSelector(cssDefinition, selectorListNode.Nodes())...)
					ruleNode.Selectors = append(ruleNode.Selectors, selectorList)
				}
			}
		case ast.CSSKindAtKeyword:
			// Setup rule node
			mediaRuleNode := new(data.CSSRule)
			for _, selectorListNode := range topNode.Selectors {
				selectorList := program.evaluateSelector(cssDefinition, selectorListNode.Nodes())
				mediaRuleNode.Selectors = append(mediaRuleNode.Selectors, selectorList)
			}

			// Get parent selector
			for _, parentSelectorListNode := range parentCSSRule.Selectors {
				selectorList := make(data.CSSSelector, 0, len(parentSelectorListNode))
				selectorList = append(selectorList, parentSelectorListNode...)
				ruleNode.Selectors = append(ruleNode.Selectors, selectorList)
			}
			mediaRuleNode.Rules = append(mediaRuleNode.Rules, ruleNode)

			// Become the wrapping @media query
			nextNodeToAppend = mediaRuleNode
		default:
			panic("evaluateCSSRule(): Unhandled CSSType.")
		}
	} else {
		for _, selectorListNode := range topNode.Selectors {
			selectorList := program.evaluateSelector(cssDefinition, selectorListNode.Nodes())
			ruleNode.Selectors = append(ruleNode.Selectors, selectorList)
		}
	}
	cssDefinition.ChildNodes = append(cssDefinition.ChildNodes, nextNodeToAppend)

	// Evaluate child nodes / properties
	ruleNode.Properties = make([]data.CSSProperty, 0, 10)
	for _, itNode := range topNode.Nodes() {
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
						identName := node.String()

						// If a variable is declared with this name, use it instead.
						variable, ok := scope.Get(identName)
						if ok {
							value.WriteString(variable.String())
							//fmt.Printf("%v\n", value)
							//panic("todo(jake): Make it use this variable value")
							continue
						}

						value.WriteString(identName)
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
			ruleNode.Properties = append(ruleNode.Properties, property)
		case *ast.CSSRule:
			program.evaluateCSSRule(cssDefinition, node, ruleNode, scope)
		default:
			panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled child node type: %T", itNode))
		}
	}
}

func (program *Program) evaluateCSSDefinition(topNode *ast.CSSDefinition, scope *Scope) *data.CSSDefinition {
	if topNode == nil {
		panic("evaluateCSSDefinition: Cannot pass nil CSSDefinition")
	}
	cssDefinition := new(data.CSSDefinition)
	if topNode != nil && topNode.Name.Kind != token.Unknown {
		cssDefinition.Name = topNode.Name.String()
	}
	cssDefinition.ChildNodes = make([]*data.CSSRule, 0, 10)
	program.globalScope.cssDefinitions = append(program.globalScope.cssDefinitions, cssDefinition)

	scope = NewScope(scope)
	for _, itNode := range topNode.Nodes() {
		switch node := itNode.(type) {
		case *ast.DeclareStatement:
			program.evaluateDeclareSet(node, scope)
		case *ast.CSSRule:
			program.evaluateCSSRule(cssDefinition, node, nil, scope)
		default:
			{
				json, _ := json.MarshalIndent(node, "", "   ")
				fmt.Printf("%s", string(json))
			}
			panic(fmt.Sprintf("evaluateCSSDefinition(): Unhandled type: %T", node))
		}
	}

	// If was unnamed, give it the filename here (to avoid it namespaces classes with the filename)
	// NOTE: We do this so the CSS definition outputs the file the CSS rules belong to
	if len(cssDefinition.Name) == 0 {
		cssDefinition.Name = program.Filepath
	}

	if scope == nil {
		panic("evaluateCSSDefinition(): Null scope provided.")
	}
	//if scope.parent != nil {
	//	{
	//		json, _ := json.MarshalIndent(scope.parent, "", "   ")
	//		fmt.Printf("%s", string(json))
	//	}
	//	panic("evaluateCSSDefinition(): Todo! Can only have CSS blocks at top-level")
	//}
	return cssDefinition
}*/
