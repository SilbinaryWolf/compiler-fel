package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
)

type HTMLBlock struct {
	HTMLKeyword token.Token // NOTE(Jake): Used to determine line number/etc
	Base
}

type HTMLComponentDefinition struct {
	Name                token.Token
	Dependencies        map[string]*HTMLNode
	Properties          *Struct
	CSSDefinition       *CSSDefinition       // optional
	CSSConfigDefinition *CSSConfigDefinition // optional
	Base
}

type HTMLNode struct {
	Name           token.Token
	Parameters     []Parameter
	HTMLDefinition *HTMLComponentDefinition // optional
	IfExpression   Expression               // optional
	Base
}
