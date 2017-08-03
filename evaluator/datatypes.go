package evaluator

import (
	"bytes"
	"strconv"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	KindString
	KindInteger64
	KindFloat64
	KindHTMLNode
)

type DataType interface {
	String() string
	Kind() Kind
}

type String struct {
	Value string
}

func (s *String) String() string {
	return s.Value
}

func (s *String) Kind() Kind {
	return KindString
}

type Integer64 struct {
	Value int64
}

func (i *Integer64) String() string {
	return strconv.FormatInt(i.Value, 10)
}

func (s *Integer64) Kind() Kind {
	return KindInteger64
}

type Float64 struct {
	Value float64
}

func (f *Float64) String() string {
	return strconv.FormatFloat(f.Value, 'f', 6, 64)
}

func (s *Float64) Kind() Kind {
	return KindFloat64
}

type HTMLNode struct {
	Name       string
	Attributes []HTMLAttribute
	ChildNodes []*HTMLNode
}

type HTMLAttribute struct {
	Name  string
	Value string
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

func (node *HTMLNode) Kind() Kind {
	return KindHTMLNode
}
