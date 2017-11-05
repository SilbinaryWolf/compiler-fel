package data

import (
	"strconv"
)

type Kind int

const (
	KindUnknown Kind = 0 + iota
	KindString
	KindInteger64
	KindFloat64
	KindBoolean
	KindHTMLNode
	KindHTMLText
	KindHTMLComponentNode
	KindStruct
	// todo(Jake): Add CSS nodes to data
	KindMixedArray
)

var kindToString = []string{
	KindUnknown:    "unknown type",
	KindString:     "string",
	KindInteger64:  "integer",
	KindFloat64:    "float",
	KindHTMLNode:   "html node",
	KindHTMLText:   "html text",
	KindStruct:     "struct",
	KindMixedArray: "mixed array",
}

func (kind Kind) String() string {
	return kindToString[kind]
}

type Type interface {
	String() string
	Kind() Kind
}

// String

type String struct {
	Value string
}

func (s *String) Kind() Kind {
	return KindString
}

func (s *String) String() string {
	return s.Value
}

// Boolean

type Bool struct {
	Value bool
}

func (s *Bool) Kind() Kind {
	return KindBoolean
}

func (i *Bool) String() string {
	if i.Value {
		return "true"
	}
	return "false"
}

// Integer64

type Integer64 struct {
	Value int64
}

func (s *Integer64) Kind() Kind {
	return KindInteger64
}

func (i *Integer64) String() string {
	return strconv.FormatInt(i.Value, 10)
}

// Float 64

type Float64 struct {
	Value float64
}

func (s *Float64) Kind() Kind {
	return KindFloat64
}

func (f *Float64) String() string {
	return strconv.FormatFloat(f.Value, 'f', 6, 64)
}

// Mixed Array

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
	panic("(array *MixedArray) String(): Not implemented")
	/*var buffer bytes.Buffer
	for _, record := range array.Array {
		buffer.WriteString(record.String())
	}
	return buffer.String()*/
}
