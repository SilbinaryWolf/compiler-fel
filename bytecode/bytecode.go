package bytecode

import (
	"fmt"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	Store
	Push
	PushStackVar
	ConditionalEqual
	Add
	Jump
	JumpIfFalse
)

var kindToString = []string{
	Unknown:          "unknown/unset bytecode",
	Store:            "Store",
	Push:             "Push",
	PushStackVar:     "PushStackVar",
	ConditionalEqual: "ConditionalEqual",
	Add:              "Add",
	Jump:             "Jump",
	JumpIfFalse:      "JumpIfFalse",
}

type Code struct {
	kind  Kind
	Value interface{}
}

// ie. a function, block-scope, HTMLComponent
type Block struct {
	Opcodes   []Code
	StackSize int
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
		result += fmt.Sprintf(" %v", code.Value)
	}
	return result
}
