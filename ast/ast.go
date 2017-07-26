package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
)

type Node interface {
}

type Base struct {
	Start int
	End   int
}

type File struct {
	Base
	Filepath string
}

type Block struct {
	Base
}

type Expression struct {
	Base
	Nodes []Node
	//InfixNodes *Base
}

type DeclareStatement struct {
	Name       token.Token
	Expression *Expression
}

type Token struct {
	token.Token
}
