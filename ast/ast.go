package ast

import (
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
)

type CSSRuleKind int

const (
	CSSKindUnknown CSSRuleKind = 0 + iota
	CSSKindRule
	CSSKindAtKeyword
)

type Node interface {
	Nodes() []Node
}

type Base struct {
	ChildNodes []Node
}

func (node *Base) Nodes() []Node {
	return node.ChildNodes
}

type File struct {
	Filepath string
	Base
}

type Block struct {
	Base
}

type Parameter struct {
	Name token.Token
	Base
}

type Expression struct {
	TypeToken token.Token
	Type      data.Kind
	Base
}

/*type NamedBlock struct {
	Name token.Token
	Block
}*/

type HTMLDefinition struct {
	Base
}

type HTMLProperties struct {
	Statements []*DeclareStatement
}

func (node *HTMLProperties) Nodes() []Node {
	return nil
}

type HTMLComponentDefinition struct {
	Name       token.Token
	Properties *HTMLProperties
	Base
}

type HTMLNode struct {
	Name           token.Token
	Parameters     []Parameter
	HTMLDefinition *HTMLComponentDefinition // optional
	CSSDefinition  *CSSDefinition           // optional
	Base
}

type DeclareStatement struct {
	Name token.Token
	Expression
}

type Token struct {
	token.Token
}

func (node *Token) Nodes() []Node {
	return nil
}

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
