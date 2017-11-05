package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
)

type TypeKind int

const (
	TypeUnknown TypeKind = 0 + iota
	TypeString
	TypeInteger64
	TypeFloat64
	TypeHTMLDefinitionNode
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
	Expression
}

type Expression struct {
	TypeInfo       types.TypeInfo
	TypeIdentifier token.Token
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
