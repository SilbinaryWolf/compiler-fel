package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
)

/*type TypeKind int

const (
	TypeUnknown TypeKind = 0 + iota
	TypeString
	TypeInteger64
	TypeFloat64
	TypeHTMLDefinitionNode
)*/

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

type Call struct {
	Name       token.Token
	Parameters []*Parameter
	Definition *ProcedureDefinition
}

func (node *Call) Nodes() []Node {
	return nil
}

/*type Block struct {
	Base
}*/

type Type struct {
	Name       token.Token
	ArrayDepth int // [] = 1, [][] = 2, [][][] = 3
}

func (node *Type) String() string {
	result := ""
	for i := 0; i < node.ArrayDepth; i++ {
		result += "[]"
	}
	result += node.Name.String()
	return result
}

func (node *Type) Nodes() []Node {
	return nil
}

type Parameter struct {
	Name token.Token
	Expression
}

type ProcedureDefinition struct {
	Name           token.Token
	Parameters     []Parameter
	TypeInfo       types.TypeInfo
	TypeIdentifier Type
	Base
}

type For struct {
	IndexName    token.Token
	RecordName   token.Token
	Array        Expression
	IsDeclareSet bool
	Base
}

// ie. for block scoping with an `if`, `for`, etc
type Block struct {
	Base
}

type If struct {
	Condition Expression
	Base
	ElseNodes []Node
}

type ArrayLiteral struct {
	TypeInfo       types.TypeInfo
	TypeIdentifier Type
	Base
}

type StructLiteral struct {
	Name     token.Token
	Fields   []StructLiteralField
	TypeInfo types.TypeInfo
}

func (node *StructLiteral) Nodes() []Node {
	return nil
}

type StructLiteralField struct {
	Name token.Token
	Expression
}

type Return struct {
	Expression
}

type Expression struct {
	TypeInfo       types.TypeInfo // determined at typecheck time (2017-12-30)
	TypeIdentifier Type           // optional, for declare statements
	Base
}

type OpStatement struct {
	LeftHandSide []token.Token
	Operator     token.Token
	Expression
}

type ArrayAppendStatement struct {
	LeftHandSide []token.Token
	Expression
}

type DeclareStatement struct {
	Name token.Token
	Expression
}

type Token struct {
	token.Token
}

type TokenList struct {
	tokens []token.Token
}

func (node *Token) Nodes() []Node {
	return nil
}

func NewTokenList(tokens []token.Token) *TokenList {
	result := new(TokenList)
	result.tokens = tokens
	return result
}

func (node *TokenList) Tokens() []token.Token {
	return node.tokens
}

func (node *TokenList) Nodes() []Node {
	return nil
}

type StructDefinition struct {
	Name   token.Token
	Fields []StructField
}

func (node *StructDefinition) GetFieldByName(name string) *StructField {
	for i := 0; i < len(node.Fields); i++ {
		field := &node.Fields[i]
		if field.Name.String() == name {
			return field
		}
	}
	return nil
}

func (node *StructDefinition) Nodes() []Node {
	return nil
}

type StructField struct {
	Name  token.Token
	Index int
	Expression
}
