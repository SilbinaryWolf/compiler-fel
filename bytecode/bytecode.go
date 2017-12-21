package bytecode

import (
	"fmt"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	Label
	Store
	StoreStructField
	StoreInternalStructField
	Push
	PushStackVar
	PushAllocStruct
	PushAllocInternalStruct
	ConditionalEqual
	Add
	AddString
	Jump
	JumpIfFalse
)

var kindToString = []string{
	Unknown:                  "unknown/unset bytecode",
	Label:                    "Label",
	Store:                    "Store",
	StoreStructField:         "StoreStructField",
	StoreInternalStructField: "StoreInternalStructField",
	Push:                    "Push",
	PushStackVar:            "PushStackVar",
	PushAllocStruct:         "PushAllocStruct",
	PushAllocInternalStruct: "PushAllocInternalStruct",
	ConditionalEqual:        "ConditionalEqual",
	Add:                     "Add",
	AddString:               "AddString",
	Jump:                    "Jump",
	JumpIfFalse:             "JumpIfFalse",
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
