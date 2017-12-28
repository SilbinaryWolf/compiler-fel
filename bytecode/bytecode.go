package bytecode

import (
	"fmt"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	Label
	Pop
	Store
	StorePopStructField
	StoreInternalStructField
	Push
	PushArrayString
	PushArrayInt
	PushArrayFloat
	PushArrayStruct
	PushStackVar
	PushAllocStruct
	PushAllocInternalStruct
	PushNewContextNode
	ConditionalEqual
	Add
	AddString
	Jump
	JumpIfFalse
)

type NodeContextType int

const (
	NodeUnknown NodeContextType = 0 + iota
	NodeCSSDefinition
)

var kindToString = []string{
	Unknown:                  "unknown/unset bytecode",
	Label:                    "Label",
	Pop:                      "Pop",
	Store:                    "Store",
	StorePopStructField:      "StorePopStructField",
	StoreInternalStructField: "StoreInternalStructField",
	Push:                    "Push",
	PushArrayString:         "PushArrayString",
	PushArrayInt:            "PushArrayInt",
	PushArrayFloat:          "PushArrayFloat",
	PushArrayStruct:         "PushArrayStruct",
	PushStackVar:            "PushStackVar",
	PushAllocStruct:         "PushAllocStruct",
	PushAllocInternalStruct: "PushAllocInternalStruct",
	PushNewContextNode:      "PushNewContextNode",
	ConditionalEqual:        "ConditionalEqual",
	Add:                     "Add",
	AddString:               "AddString",
	Jump:                    "Jump",
	JumpIfFalse:             "JumpIfFalse",
}

type Code struct {
	Kind  Kind
	Value interface{}
}

// ie. a function, block-scope, HTMLComponent
type Block struct {
	Opcodes   []Code
	StackSize int
}

type StructInterface interface {
	GetField(index int) interface{}
	SetField(index int, value interface{})
}

type Struct struct {
	fields []interface{}
}

func NewStruct(fieldCount int) *Struct {
	structData := new(Struct)
	structData.fields = make([]interface{}, fieldCount)
	return structData
}

func (structData *Struct) SetField(index int, value interface{}) {
	structData.fields[index] = value
}

func (structData *Struct) GetField(index int) interface{} {
	return structData.fields[index]
}

func (structData *Struct) FieldCount() int {
	return len(structData.fields)
}

func (kind Kind) String() string {
	return kindToString[kind]
}

func (code *Code) String() string {
	result := code.Kind.String()
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
