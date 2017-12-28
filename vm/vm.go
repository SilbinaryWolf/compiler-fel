package vm

import (
	"fmt"
	"reflect"

	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/data"
)

type Program struct {
	stack            []interface{}
	nodeStackContext []interface{} // stack of node contexts for tracking CSS rules / current HTML node.
}

func ExecuteBytecode(codeBlock *bytecode.Block) {
	program := new(Program)
	program.stack = make([]interface{}, codeBlock.StackSize)

	registerStack := make([]interface{}, 0, 4)

	opcodes := codeBlock.Opcodes
	offset := 0
	for offset < len(opcodes) {
		code := opcodes[offset]

		switch code.Kind {
		case bytecode.Label:
			// no-op
		case bytecode.Push:
			registerStack = append(registerStack, code.Value)
		case bytecode.PushArrayString:
			value := make([]int, 0)
			registerStack = append(registerStack, value)
		case bytecode.PushArrayInt:
			value := make([]int, 0)
			registerStack = append(registerStack, value)
		case bytecode.PushArrayFloat:
			value := make([]int, 0)
			registerStack = append(registerStack, value)
		case bytecode.PushArrayStruct:
			value := make([]bytecode.Struct, 0)
			registerStack = append(registerStack, value)
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
		case bytecode.PushNewContextNode:
			var node interface{}
			switch code.Value.(bytecode.NodeContextType) {
			case bytecode.NodeCSSDefinition:
				node = new(data.CSSDefinition)
			default:
				panic("Unhandled NodeContextType")
			}
			program.nodeStackContext = append(program.nodeStackContext, node)
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
		case bytecode.Pop:
			//popAmount := code.Value.(int) + 1
			registerStack = registerStack[:len(registerStack)-1]
		case bytecode.Store:
			value := registerStack[len(registerStack)-1]

			stackOffset := code.Value.(int)
			program.stack[stackOffset] = value
		case bytecode.StorePopStructField:
			fieldData := registerStack[len(registerStack)-1]
			structData := registerStack[len(registerStack)-2].(*bytecode.Struct)

			// NOTE(Jake): Only pop `fieldData`
			registerStack = registerStack[:len(registerStack)-1]

			fieldOffset := code.Value.(int)
			structData.SetField(fieldOffset, fieldData)
		case bytecode.StoreInternalStructField:
			panic("No longer supported")
			/*fieldData := registerStack[len(registerStack)-1]
			structData := registerStack[len(registerStack)-2]

			// NOTE(Jake): Only pop `fieldData`
			registerStack = registerStack[:len(registerStack)-1]

			// NOTE(Jake): This might not work as I think it does... need to investigate
			fieldOffset := []int{code.Value.(int)}
			structField := reflect.ValueOf(structData).FieldByIndex(fieldOffset)
			structField.Set(reflect.ValueOf(fieldData))
			panic("todo(Jake): Add reflect.GetField or whatever here")*/
		default:
			panic(fmt.Sprintf("executeBytecode: Unhandled kind in vm: \"%s\"", code.Kind.String()))
		}
		offset++
	}

	// Debug
	debugPrintStack(program.stack)
}
