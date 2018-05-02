package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
)

type CSSRuleKind int

const (
	CSSKindUnknown CSSRuleKind = 0 + iota
	CSSKindRule
	CSSKindAtKeyword
)

type CSSDefinition struct {
	Name token.Token
	Base
}

type CSSRule struct {
	kind      CSSRuleKind
	selectors []CSSSelector
	Base
}

func (rule *CSSRule) Kind() CSSRuleKind        { return rule.kind }
func (rule *CSSRule) Selectors() []CSSSelector { return rule.selectors }

func NewCSSRule(kind CSSRuleKind, selectors []CSSSelector) *CSSRule {
	rule := new(CSSRule)
	rule.kind = kind
	rule.selectors = selectors
	return rule
}

type CSSSelector struct {
	Base
}

type CSSAttributeSelector struct {
	Name     token.Token
	Operator token.Token
	Value    token.Token
}

func (node *CSSAttributeSelector) Nodes() []Node {
	return nil
}

type CSSProperty struct {
	Name token.Token
	Base
}
