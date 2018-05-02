package ast

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/token"
)

type CallKind int

const (
	CallUnknown CallKind = 0 + iota
	CallProcedure
	CallHTMLNode
)

var callKindToString = []string{
	CallUnknown:   "unknown call kind",
	CallProcedure: "procedure",
	CallHTMLNode:  "html node",
}

func (kind CallKind) String() string {
	return callKindToString[kind]
}

type LeftHandSide []token.Token

func (parts LeftHandSide) String() string {
	result := ""
	for i, val := range parts {
		if i != 0 {
			result += "." + val.String()
			continue
		}
		result += val.String()
	}
	return result
}

type Node interface {
	Nodes() []Node
}

type Base struct {
	ChildNodes []Node
}

//func (node *Base) SetNodes(nodes []Node) {
//	node.childNodes = nodes
//}

func (node *Base) Nodes() []Node {
	return node.ChildNodes
}

type File struct {
	Filepath     string
	Dependencies map[string]bool
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

//type Block struct {
//	Base
//}

type TypeInfo interface {
	String() string
	ImplementsTypeInfo()
}

type TypeIdent struct {
	Name       token.Token
	ArrayDepth int // [] = 1, [][] = 2, [][][] = 3
}

func (node *TypeIdent) String() string {
	result := ""
	for i := 0; i < node.ArrayDepth; i++ {
		result += "[]"
	}
	result += node.Name.String()
	return result
}

func (node *TypeIdent) Nodes() []Node {
	return nil
}

type Parameter struct {
	Name token.Token
	Expression
}

type ProcedureDefinition struct {
	Name           token.Token
	Parameters     []Parameter
	TypeInfo       TypeInfo
	TypeIdentifier TypeIdent
	Base
}

type For struct {
	IndexName    token.Token
	RecordName   token.Token
	Array        Expression
	IsDeclareSet bool
	Base
}

type Block struct {
	Base
}

type If struct {
	Condition Expression
	Base
	ElseNodes []Node
}

type ArrayLiteral struct {
	TypeInfo       TypeInfo
	TypeIdentifier TypeIdent
	Base
}

type StructLiteral struct {
	Name     token.Token
	Fields   []Parameter
	TypeInfo TypeInfo
}

type WorkspaceDefinition struct {
	Name              token.Token
	WorkspaceTypeInfo TypeInfo
	Base
}

func (node *StructLiteral) Nodes() []Node {
	return nil
}

type Return struct {
	Expression
}

type LeftHandSideExpression struct {
	tokens []token.Token
}

func (leftHandSideExpression *LeftHandSideExpression) Nodes() []token.Token {
	return leftHandSideExpression.tokens
}

func NewLeftHandSideExpression(tokens []token.Token) LeftHandSideExpression {
	return LeftHandSideExpression{
		tokens: tokens,
	}
}

func (leftHandSideExpression *LeftHandSideExpression) String() string {
	parts := leftHandSideExpression.Nodes()
	result := ""
	for i, val := range parts {
		if i != 0 {
			result += "." + val.String()
			continue
		}
		result += val.String()
	}
	return result
}

type Expression struct {
	TypeInfo       TypeInfo  // determined at typecheck time (2017-12-30)
	TypeIdentifier TypeIdent // optional, for declare statements
	childNodes     []Node
}

func NewExpression(nodes []Node) Expression {
	return Expression{
		childNodes: nodes,
	}
}

func (exprNode *Expression) Nodes() []Node {
	return exprNode.childNodes
}

func (exprNode *Expression) String() string {
	result := ""
	for i, node := range exprNode.Nodes() {
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

type ArrayAccessStatement struct {
	LeftHandSide LeftHandSideExpression
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
	name   token.Token
	fields []StructField
}

func (structDef *StructDefinition) Name() token.Token     { return structDef.name }
func (structDef *StructDefinition) Fields() []StructField { return structDef.fields }

func NewStructDefinition(name token.Token, fields []StructField) *StructDefinition {
	return &StructDefinition{
		name:   name,
		fields: fields,
	}
}

func (structDef *StructDefinition) GetFieldByName(name string) *StructField {
	fields := structDef.Fields()
	for i := 0; i < len(fields); i++ {
		field := &fields[i]
		if field.Name().String() == name {
			return field
		}
	}
	return nil
}

func (structDef *StructDefinition) Nodes() []Node {
	return nil
}

type StructField struct {
	name  token.Token
	index int
	Expression
}

func (field *StructField) Name() token.Token { return field.name }
func (field *StructField) Index() int        { return field.index }

func CreateStructField(name token.Token, index int, expression Expression, typeIdentifier TypeIdent) StructField {
	expression.TypeIdentifier = typeIdentifier
	return StructField{
		name:       name,
		index:      index,
		Expression: expression,
	}
}
