package vm

import (
	"fmt"
	"reflect"

	"github.com/silbinarywolf/compiler-fel/bytecode"
)

type Program struct {
	stack []interface{}
	// NOTE(Jake):
	//nodeContext []interface{}
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
		case bytecode.Label:
			// no-op
		case bytecode.Push:
			registerStack = append(registerStack, code.Value)
		case bytecode.PushStackVar:
			stackOffset := code.Value.(int)
			registerStack = append(registerStack, program.stack[stackOffset])
		case bytecode.PushAllocStruct:
			structFieldCount := code.Value.(int)
			structData := bytecode.NewStruct(structFieldCount)
			registerStack = append(registerStack, structData)
		case bytecode.PushAllocInternalStruct:
			internalType := code.Value.(reflect.Type)
			structData := reflect.Indirect(reflect.New(internalType)).Interface()
			registerStack = append(registerStack, structData)
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
		case bytecode.AddString:
			valueA := registerStack[len(registerStack)-2].(string)
			valueB := registerStack[len(registerStack)-1].(string)
			registerStack = registerStack[:len(registerStack)-2]
			registerStack = append(registerStack, valueA+valueB)
		case bytecode.Store:
			value := registerStack[len(registerStack)-1]
			registerStack = registerStack[:len(registerStack)-1]

			stackOffset := code.Value.(int)
			program.stack[stackOffset] = value
			//panic(program.stack[stackOffset])
		case bytecode.StoreStructField:
			fieldData := registerStack[len(registerStack)-1]
			structData := registerStack[len(registerStack)-2].(bytecode.StructInterface)

			// NOTE(Jake): Only pop `fieldData`
			registerStack = registerStack[:len(registerStack)-1]

			fieldOffset := code.Value.(int)
			structData.SetField(fieldOffset, fieldData)
		case bytecode.StoreInternalStructField:
			fieldData := registerStack[len(registerStack)-1]
			structData := registerStack[len(registerStack)-2]

			// NOTE(Jake): Only pop `fieldData`
			registerStack = registerStack[:len(registerStack)-1]

			// NOTE(Jake): This might not work as I think it does... need to investigate
			fieldOffset := []int{code.Value.(int)}
			structField := reflect.ValueOf(structData).FieldByIndex(fieldOffset)
			structField.Set(reflect.ValueOf(fieldData))
			panic("todo(Jake): Add reflect.GetField or whatever here")
		default:
			panic(fmt.Sprintf("executeBytecode: Unhandled kind in vm: \"%s\"", code.Kind().String()))
		}
		offset++
	}

	// Debug
	fmt.Printf("----------------\nVM Stack Values:\n----------------\n")
	for i := 0; i < len(program.stack); i++ {
		stackValue := program.stack[i]
		fmt.Printf("%v - %T\n", stackValue, stackValue)
	}
	fmt.Printf("----------------\n")
}
