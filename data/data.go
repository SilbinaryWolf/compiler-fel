package data

import (
	"bytes"
	"fmt"
	"strconv"
)

type Kind int

type SelectorKind int

const (
	KindUnknown Kind = 0 + iota
	KindString
	KindInteger64
	KindFloat64
	KindHTMLNode
	KindHTMLText
	KindStruct
	// todo(Jake): Add CSS nodes to data
	KindMixedArray
)

type Type interface {
	String() string
	Kind() Kind
}

type String struct {
	Value string
}

func (s *String) Kind() Kind {
	return KindString
}

func (s *String) String() string {
	return s.Value
}

type Integer64 struct {
	Value int64
}

func (s *Integer64) Kind() Kind {
	return KindInteger64
}

func (i *Integer64) String() string {
	return strconv.FormatInt(i.Value, 10)
}

type Float64 struct {
	Value float64
}

func (s *Float64) Kind() Kind {
	return KindFloat64
}

func (f *Float64) String() string {
	return strconv.FormatFloat(f.Value, 'f', 6, 64)
}

type HTMLNode struct {
	Name       string
	Attributes []HTMLAttribute
	ChildNodes []Type
}

type HTMLAttribute struct {
	Name  string
	Value string
}

func (node *HTMLNode) Kind() Kind {
	return KindHTMLNode
}

func (node *HTMLNode) String() string {
	var buffer bytes.Buffer
	buffer.WriteByte('<')
	buffer.WriteString(node.Name)
	buffer.WriteByte(' ')
	for _, attribute := range node.Attributes {
		buffer.WriteString(attribute.Name)
		buffer.WriteString("=\"")
		buffer.WriteString(attribute.Value)
		buffer.WriteString("\" ")
	}
	buffer.WriteByte('>')
	return buffer.String()
}

type HTMLText struct {
	Value string
}

func (node *HTMLText) Kind() Kind {
	return KindHTMLText
}

func (node *HTMLText) String() string {
	return node.Value
}

type CSSDefinition struct {
	Name       string
	ChildNodes []*CSSRule
}

type CSSRule struct {
	Selectors  []CSSSelector
	Properties []CSSProperty
	Rules      []*CSSRule
}

type CSSSelector []CSSSelectorNode

type CSSSelectorNode interface {
	String() string
}

//type CSSSelector struct {
//	ChildNodes []CSSSelectorNode
//}

func (nodes CSSSelector) String() string {
	result := ""
	for _, node := range nodes {
		result += node.String() + " "
	}
	result = result[:len(result)-1]
	return result
}

type CSSSelectorIdentifier struct {
	Name string
}

func (node *CSSSelectorIdentifier) String() string {
	return node.Name
}

type CSSSelectorOperator struct {
	Operator string
}

func (node *CSSSelectorOperator) String() string {
	return node.Operator
}

type CSSSelectorAttribute struct {
	Name     string
	Operator string
	Value    string
}

func (node *CSSSelectorAttribute) String() string {
	return fmt.Sprintf("[%s%s%s]", node.Name, node.Operator, node.Value)
}

type CSSProperty struct {
	Name  string
	Value string
}

func (property *CSSProperty) String() string {
	return fmt.Sprintf("%s: %s;", property.Name, property.Value)
}

type MixedArray struct {
	Array []Type
}

func NewMixedArray(array []Type) *MixedArray {
	result := new(MixedArray)
	result.Array = array
	return result
}

func (array *MixedArray) Kind() Kind {
	return KindMixedArray
}

func (array *MixedArray) String() string {
	panic("No")
	var buffer bytes.Buffer
	for _, record := range array.Array {
		buffer.WriteString(record.String())
	}
	return buffer.String()
}
