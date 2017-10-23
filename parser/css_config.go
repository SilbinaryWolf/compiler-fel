package parser

import (
	"fmt"
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseCSSConfigRuleDefinition(name token.Token) *ast.CSSConfigDefinition {
	p.SetScanMode(scanner.ModeCSS)
	nodes := p.parseCSSStatements()
	p.SetScanMode(scanner.ModeDefault)

	// Check / read data from ast
	cssConfigDefinition := new(ast.CSSConfigDefinition)
	cssConfigDefinition.Name = name
	//cssConfigDefinition.Rules = make([]ast.CSSConfigMatchPart, 0, len(nodes))
	for _, itNode := range nodes {
		switch node := itNode.(type) {
		case *ast.CSSRule:
			configRule := ast.NewCSSConfigRule()

			// Get rules
			for _, itNode := range node.ChildNodes {
				switch node := itNode.(type) {
				case *ast.CSSProperty:
					name := node.Name.String()
					switch name {
					case "modify_name":
						value, ok := p.getBoolFromCSSConfigProperty(node)
						if !ok {
							return nil
						}
						configRule.ModifyName = value
					default:
						p.addErrorToken(fmt.Errorf("Invalid config key \"%s\". Expected \"modify_name\".", name), node.Name)
						return nil
					}
				case *ast.DeclareStatement:
					p.addErrorToken(fmt.Errorf("Cannot declare variables in a css_config block. Did you mean to use : instead of :="), node.Name)
					return nil
				default:
					panic(fmt.Sprintf("parseCSSConfigRuleDefinition:propertyLoop: Unknown type %T", node))
				}
			}

			// Get matching parts
			for _, selector := range node.Selectors {
				rulePartList := make(ast.CSSConfigMatchPart, 0, len(selector.ChildNodes))
				for _, itSelectorPart := range selector.ChildNodes {
					switch selectorPartNode := itSelectorPart.(type) {
					case *ast.Token:
						if selectorPartNode.IsOperator() {
							operator := selectorPartNode.Kind.String()
							switch selectorPartNode.Kind {
							case token.Multiply:
								rulePartList = append(rulePartList, operator)
							default:
								p.addErrorToken(fmt.Errorf("Only supports * wildcard, not %s", operator), selectorPartNode.Token)
								return nil
							}
							continue
						}
						if selectorPartNode.Kind != token.Identifier {
							p.addErrorToken(fmt.Errorf("Expected identifier, instead got %s", selectorPartNode.Kind.String()), selectorPartNode.Token)
							return nil
						}
						name := selectorPartNode.String()
						rulePartList = append(rulePartList, name)
					default:
						panic(fmt.Sprintf("parseCSSConfigRuleDefinition:selectorPartLoop: Unknown type %T", selectorPartNode))
					}
				}
				configRule.Selectors = append(configRule.Selectors, rulePartList)
			}

			// Generate string. (For easy feeding into path.Match() function)
			for _, selector := range configRule.Selectors {
				pattern := ""
				for _, part := range selector {
					pattern += part
				}
				configRule.SelectorsAsPattern = append(configRule.SelectorsAsPattern, pattern)
			}

			cssConfigDefinition.Rules = append(cssConfigDefinition.Rules, configRule)
		case *ast.DeclareStatement:
			p.addErrorToken(fmt.Errorf("Cannot declare variables in a css_config block."), node.Name)
			return nil
		default:
			panic(fmt.Sprintf("parseCSSConfigRuleDefinition: Unknown type %T", node))
		}
	}

	// Test
	//config := cssConfigDefinition.GetRule(".js-")
	//fmt.Printf("\n\n %v \n\n", config)
	//panic("parser/css_config.go test")

	return cssConfigDefinition
}

func (p *Parser) getBoolFromCSSConfigProperty(node *ast.CSSProperty) (bool, bool) {
	if len(node.ChildNodes) == 0 && len(node.ChildNodes) > 1 {
		p.addErrorToken(fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", node.Name.String()), node.Name)
		return false, false
	}
	itNode := node.ChildNodes[0]
	switch node := itNode.(type) {
	case *ast.Token:
		t := node.Token
		if t.Kind != token.Identifier {
			p.addErrorToken(fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", t.String()), t)
			return false, false
		}
		valueString := node.String()
		var value bool
		var ok bool
		if valueString == "true" {
			value = true
			ok = true
		}
		if !ok && valueString == "false" {
			value = false
			ok = true
		}
		if !ok {
			p.addErrorToken(fmt.Errorf("Expected \"true\" or \"false\" after \"%s\".", t.String()), t)
			return false, false
		}
		return value, ok
	default:
		panic(fmt.Sprintf("parseCSSConfigRuleDefinition:propertyValueLoop: Unknown type %T", node))
	}
	return false, false
}

/*

func getNewSelectorPartList() []ast.Node {
	return make([]ast.Node, 0, 10)
}

func (p *Parser) parseCSSConfigSelector(firstToken token.Token) ast.CSSSelectorWildcard {
	selectorWildcardParts := make([]ast.Node, 0, 10)
	selectorWildcardParts = append(selectorWildcardParts, &ast.Token{Token: firstToken})

Loop:
	for {
		t := p.PeekNextToken()
		switch t.Kind {
		case token.Identifier, token.Multiply:
			p.GetNextToken()
			selectorWildcardParts = append(selectorWildcardParts, &ast.Token{Token: t})
		case token.Comma:
			p.GetNextToken()
			break Loop
		case token.BraceOpen:
			break Loop
		case token.Whitespace, token.Newline:
			// no-op
			p.GetNextToken()
		default:
			panic(fmt.Sprintf("parseCSSConfigSelector(): Unhandled token type(%d): \"%s\" (value: %s) on Line %d", t.Kind, t.Kind.String(), t.String(), t.Line))
		}
	}

	result := ast.CSSSelectorWildcard{}
	result.ChildNodes = selectorWildcardParts
	return result
}

func (p *Parser) parseCSSConfigRules() []ast.Node {
	selectorWildcardList := make([]ast.CSSSelectorWildcard, 0, 3)

	for {
		p.eatWhitespace()
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			selector := p.parseCSSConfigSelector(t)
			selectorWildcardList = append(selectorWildcardList, selector)
		case token.BraceOpen:

		case token.Whitespace, token.Newline:
			// no-op
		default:
			panic(fmt.Sprintf("parseCSSConfigRules(): Unhandled token type(%d): \"%s\" (value: %s) on Line %d", t.Kind, t.Kind.String(), t.String(), t.Line))
		}
	}
	return nil
}*/
