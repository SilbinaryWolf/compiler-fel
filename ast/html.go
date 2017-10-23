package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
)

type HTMLBlock struct {
	HTMLKeyword token.Token // NOTE(Jake): Used to determine line number/etc
	Base
}

type HTMLProperties struct {
	Statements []*DeclareStatement
}

func (node *HTMLProperties) Nodes() []Node {
	return nil
}

type HTMLComponentDefinition struct {
	Name                token.Token
	Dependencies        map[string]*HTMLNode
	Properties          *HTMLProperties
	CSSDefinition       *CSSDefinition       // optional
	CSSConfigDefinition *CSSConfigDefinition // optional
	Base
}

type HTMLNode struct {
	Name           token.Token
	Parameters     []Parameter
	HTMLDefinition *HTMLComponentDefinition // optional
	Base
}
