package ast

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
)

type CallKind int

const (
	CallUnknown CallKind = 0 + iota
	CallProcedure
	CallHTMLNode
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

// todo(Jake): 2018-01-09
//
// Refactor HTMLNode into Call
//
type Call struct {
	kind CallKind
	// Shared
	Name       token.Token
	Parameters []*Parameter
	Definition *ProcedureDefinition
	// HTMLNode only
	HTMLDefinition *HTMLComponentDefinition // optional
	IfExpression   Expression               // optional
	Base
}

func NewCall() *Call {
	node := new(Call)
	node.kind = CallProcedure
	return node
}

func NewHTMLNode() *Call {
	node := new(Call)
	node.kind = CallHTMLNode
	return node
}

func (node *Call) Kind() CallKind {
	return node.kind
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

func (exprNode *Expression) String() string {
	result := ""
	for i, node := range exprNode.ChildNodes {
		if i != 0 {
			result += " | "
		}
		switch node := node.(type) {
		case *Call:
			result += node.Name.String() + "("
			for i, parameter := range node.Parameters {
				if i != 0 {
					result += ","
				}
				result += parameter.Expression.String()
			}
			result += ")"
		case *Token:
			result += node.String()
		case *TokenList:
			result += node.String()
		default:
			panic(fmt.Sprintf("Expression:String: Unhandled type %T", node))
		}
	}
	return result
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

func (node *Token) Nodes() []Node {
	return nil
}

type TokenList struct {
	tokens []token.Token
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

func (node *TokenList) String() string {
	tokens := node.Tokens()
	concatPropertyName := tokens[0].String()
	for i := 1; i < len(tokens); i++ {
		concatPropertyName += "." + tokens[i].String()
	}
	return concatPropertyName
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
