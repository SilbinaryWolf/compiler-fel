package vm

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

	start_builtin_call
	CallPushNode
	end_builtin_call
)

type Opcode struct {
	Code Kind
	Data interface{}
}

type VM struct {
	OpcodeIndex int
	Opcodes     []Opcode

	StackIndex int
	Stack      []interface{}

	Return []interface{}
}

func New() *VM {
	vm := new(VM)
	vm.Opcodes = make([]Opcode, 0, 256)
	vm.Stack = make([]interface{}, 512)
	vm.Return = make([]interface{}, 256)
	return vm
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
	return
}

func (vm *VM) pushNode() *data.HTMLNode {
	argCountInterface := vm.pop()
	argCount, ok := argCountInterface.(int)
	if DEBUG && !ok {
		//vm.PrintDebug()
		panic(fmt.Sprintf("callPushNode: Expected int type not %T.", argCountInterface))
	}

	attributes := make([]data.HTMLAttribute, argCount)

	for i := 0; i < argCount; i++ {
		attributes[i].Name, ok = vm.pop().(string)
		if DEBUG && !ok {
			panic(fmt.Sprintf("callPushNode: Expected string type for argument name %d.", i))
		}
		attributes[i].Value, ok = vm.pop().(string)
		if DEBUG && !ok {
			panic(fmt.Sprintf("callPushNode: Expected string type for argument value %d.", i))
		}
	}

	htmlNode := new(data.HTMLNode)
	htmlNode.Attributes = attributes
	return htmlNode
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
		case Call:
			switch function := opcode.Data.(type) {
			case func():
				function()
			default:
				panic(fmt.Errorf("Unhandled type for call: %T", opcode.Data))
			}
		case CallPushNode:
			vm.push(vm.pushNode())
		default:
			panic("Invalid opcode")
			break Loop
		}
	}
}
