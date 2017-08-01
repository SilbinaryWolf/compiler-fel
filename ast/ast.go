package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
)

type Node interface {
	Nodes() []Node
}

type Base struct {
	Start      int
	End        int
	ChildNodes []Node
}

type File struct {
	Base
	Filepath string
}

type Block struct {
	Base
}

/*type NamedBlock struct {
	Name token.Token
	Block
}*/

type HTMLNode struct {
	Name token.Token
	Base
}

type Expression struct {
	Base
	//InfixNodes *Base
}

type DeclareStatement struct {
	Name token.Token
	*Expression
}

type Token struct {
	Node
	token.Token
}

func (node *Base) Nodes() []Node {
	return node.ChildNodes
}

func (node *Token) Nodes() []Node {
	return nil
}
