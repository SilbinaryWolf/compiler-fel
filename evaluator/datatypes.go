package evaluator

import "strconv"

type Kind int

const (
	Unknown Kind = 0 + iota
	KindString
	KindInteger64
	KindFloat64
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