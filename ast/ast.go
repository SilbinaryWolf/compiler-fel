package ast

import "github.com/silbinarywolf/compiler-fel/token"

type Node interface {
	Nodes() []Node
}

type Base struct {
	ChildNodes []Node
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

/*type NamedBlock struct {
	Name token.Token
	Block
}*/

type HTMLNode struct {
	Name       token.Token
	Parameters []Parameter
	Base
}

type Expression struct {
	Base
	//InfixNodes *Base
}

type DeclareStatement struct {
	Name token.Token
	Base
}

type Token struct {
	token.Token
}

func (node *Base) Nodes() []Node {
	return node.ChildNodes
}

func (node *Token) Nodes() []Node {
	return nil
}
