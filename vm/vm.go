package vm

import (
	"fmt"
	"github.com/silbinarywolf/compiler-fel/bytecode"
)

type Program struct {
	stack []interface{}
}

func ExecuteBytecode(codeBlock *bytecode.Block) {
	program := new(Program)
	program.stack = make([]interface{}, codeBlock.StackSize)

	registerStack := make([]interface{}, 0, 4)

	opcodes := codeBlock.Opcodes
	offset := 0
	for offset < len(opcodes) {
		code := opcodes[offset]

		switch code.Kind() {
		case bytecode.Push:
			registerStack = append(registerStack, code.Value)
		case bytecode.PushStackVar:
			stackOffset := code.Value.(int)
			registerStack = append(registerStack, program.stack[stackOffset])
		case bytecode.ConditionalEqual:
			valueA := registerStack[len(registerStack)-2].(int64)
			valueB := registerStack[len(registerStack)-1].(int64)
			registerStack = registerStack[:len(registerStack)-2]
			registerStack = append(registerStack, valueA == valueB)
		case bytecode.JumpIfFalse:
			boolValue := registerStack[len(registerStack)-1].(bool)
			registerStack = registerStack[:len(registerStack)-1]
			if !boolValue {
				offset = code.Value.(int)
				continue
			}
		case bytecode.Add:
			valueA := registerStack[len(registerStack)-2].(int64)
			valueB := registerStack[len(registerStack)-1].(int64)
			registerStack = registerStack[:len(registerStack)-2]
			registerStack = append(registerStack, valueA+valueB)
		case bytecode.Store:
			value := registerStack[len(registerStack)-1]
			registerStack = registerStack[:len(registerStack)-1]

			stackOffset := code.Value.(int)
			program.stack[stackOffset] = value
		default:
			panic(fmt.Sprintf("executeBytecode: Unhandled kind in vm: \"%s\"", code.Kind().String()))
		}
		offset++
	}
}
