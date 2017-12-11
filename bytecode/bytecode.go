package bytecode

import (
	"fmt"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	Set
	Push
)

var kindToString = []string{
	Unknown: "unknown/unset bytecode",
	Set:     "Set",
	Push:    "Push",
}

type Code struct {
	kind     Kind
	Value    interface{}
	StackPos int // Stack offset where interface{} is stored in memory
}

func Init(kind Kind) Code {
	code := Code{}
	code.kind = kind
	return code
}

func (code *Code) Kind() Kind {
	return code.kind
}

func (kind Kind) String() string {
	return kindToString[kind]
}

func (code *Code) String() string {
	result := code.Kind().String()
	if code.Value != nil {
		switch value := code.Value.(type) {
		default:
			result += fmt.Sprintf(" %v", value)
		}
	}
	if code.StackPos > 0 {
		result += fmt.Sprintf(" %d", code.StackPos)
	}
	return result
}
