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
type Bool struct {
	Value bool
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
	type_    Type // NOTE: Store type via interface{}
}

func NewArray(t Type) *Array {
	res := new(Array)
	res.type_ = t
	return res
}

func (array *Array) Type() Type {
	return array.type_
}

func (array *Array) Push(value Type) {
	array.Elements = append(array.Elements, value)
}

func (array *Array) String() string {
	panic("(array *MixedArray) String(): Not implemented")
}
