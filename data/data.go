package data

import "strconv"

type Type interface {
	String() string
}

type String struct {
	Value string
}

func (s *String) String() string {
	return s.Value
}

// Boolean
var boolFalse = &Bool{value: false}
var boolTrue = &Bool{value: true}

type Bool struct {
	value bool
}

func NewBool(value bool) *Bool {
	if value {
		return boolTrue
	}
	return boolFalse
}

func (b *Bool) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b *Bool) Value() bool {
	return b.value
}

// Integer64
type Integer64 struct {
	Value int64
}

func (i *Integer64) String() string {
	return strconv.FormatInt(i.Value, 10)
}

// Float 64
type Float64 struct {
	Value float64
}

func (f *Float64) String() string {
	return strconv.FormatFloat(f.Value, 'f', 6, 64)
}

// Array
type Array struct {
	Elements []Type
	typeinfo interface{}
}

func NewArray(t interface{}) *Array {
	res := new(Array)
	res.typeinfo = t
	return res
}

func (array *Array) Type() interface{} {
	return array.typeinfo
}

func (array *Array) Push(value Type) {
	array.Elements = append(array.Elements, value)
}

func (array *Array) String() string {
	panic("(array *Array) String(): Not implemented")
}

// Struct
type Struct struct {
	Fields   []Type
	typeinfo interface{}
}

func (value *Struct) String() string {
	result := "("
	for i, field := range value.Fields {
		if i != 0 {
			result += ","
		}
		str := field.String()
		if str == "" {
			str = "\"\""
		}
		result += str
	}
	result += ")"
	return result
	//panic("(array *Struct) String(): Not implemented")
}

func NewStruct(t interface{}) *Struct {
	res := new(Struct)
	res.typeinfo = t
	return res
}
