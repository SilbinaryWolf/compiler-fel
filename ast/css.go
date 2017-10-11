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
	Kind      CSSRuleKind
	Selectors []CSSSelector
	Base
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
