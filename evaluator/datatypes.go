package evaluator

import "strconv"

type DataType interface {
	String() string
}

type String struct {
	Value string
}

func (s *String) String() string {
	return s.Value
}

type Integer64 struct {
	Value int64
}

func (i *Integer64) String() string {
	return strconv.FormatInt(i.Value, 10)
}

type Float64 struct {
	Value float64
}

func (f *Float64) String() string {
	return strconv.FormatFloat(f.Value, 'f', 6, 64)
}
