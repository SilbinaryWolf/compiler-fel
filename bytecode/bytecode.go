package bytecode

import (
	"fmt"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	DebugString
	AllocStruct
	Store
	StoreStructField
	Push
	PushStackVar
	ConditionalEqual
	Add
	Jump
	JumpIfFalse
)

var kindToString = []string{
	Unknown:          "unknown/unset bytecode",
	DebugString:      "DebugString",
	AllocStruct:      "AllocStruct",
	Store:            "Store",
	StoreStructField: "StoreStructField",
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

type Struct struct {
	Fields           []interface{}
	StructDefinition interface{}
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
		switch code.Value.(type) {
		case string:
			result += fmt.Sprintf(" \"%v\"", code.Value)
		default:
			result += fmt.Sprintf(" %v", code.Value)
		}
	}
	return result
}
