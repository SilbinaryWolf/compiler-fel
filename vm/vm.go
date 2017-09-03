/*package vm

import (
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
)

type Kind byte

const (
	DEBUG = true
)

// Instructions
const (
	Nil Kind = iota + 0
	Push
	Pop
	Call
	StoreInStruct

	start_builtin_call
	CallAllocStruct
	CallPushNode
	end_builtin_call
)

type StructDefinition struct {
	name          string
	properties    []StructDefinitionProperty
	defaultValues []interface{}
}

type StructDefinitionProperty struct {
	Name         string
	Type         data.Kind
	DefaultValue string
}

type StructInstance struct {
	Properties []interface{}
	Definition *StructDefinition
}

type Opcode struct {
	Code Kind
	Data interface{}
}

type VM struct {
	OpcodeIndex int
	Opcodes     []Opcode

	StackIndex int
	Stack      []interface{}

	Return            []interface{}
	StructDefinitions map[string]*StructDefinition
}

func New() *VM {
	vm := new(VM)
	vm.Opcodes = make([]Opcode, 0, 256)
	vm.Stack = make([]interface{}, 512)
	vm.Return = make([]interface{}, 256)
	vm.StructDefinitions = make(map[string]*StructDefinition)
	return vm
}

func (vm *VM) NewStructDefinition(name string, properties ...StructDefinitionProperty) *StructDefinition {
	structDef := new(StructDefinition)
	structDef.name = name
	structDef.properties = properties

	structDef.defaultValues = make([]interface{}, len(properties))
	for i, property := range properties {
		structDef.defaultValues[i] = property.DefaultValue
	}

	vm.StructDefinitions[structDef.name] = structDef
	return structDef
}

func (vm *VM) AddOpcode(code Kind, data interface{}) {
	vm.Opcodes = append(vm.Opcodes, Opcode{
		Code: code,
		Data: data,
	})
}

func (vm *VM) PrintDebug() {
	json, _ := json.MarshalIndent(vm, "", "   ")
	fmt.Printf("%s\n", string(json))
}

func (vm *VM) push(data interface{}) {
	vm.Stack[vm.StackIndex] = data
	vm.StackIndex++
}

func (vm *VM) pop() interface{} {
	vm.StackIndex--
	return vm.Stack[vm.StackIndex]
}

func (vm *VM) Run() {
Loop:
	for {
		opcode := vm.Opcodes[vm.OpcodeIndex]
		vm.OpcodeIndex++
		switch opcode.Code {
		case Nil:
			break Loop
		case Push:
			vm.push(opcode.Data)
		case Pop:
			vm.StackIndex--
		case StoreInStruct:
			offset, ok := opcode.Data.(int)
			if DEBUG && !ok {
				panic("StoreInStruct: Expected int.")
			}

		case Call:
			switch function := opcode.Data.(type) {
			case func():
				function()
			default:
				panic(fmt.Errorf("Unhandled type for call: %T", opcode.Data))
			}
		case CallAllocStruct:
			vm.allocStruct()
		case CallPushNode:
			vm.pushNode()
		default:
			panic("Invalid opcode")
			break Loop
		}
	}
}*/
