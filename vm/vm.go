package vm

import (
	"fmt"
	"reflect"

	"github.com/silbinarywolf/compiler-fel/bytecode"
)

type Program struct {
	stack            []interface{}
	registerStack    []interface{}
	nodeStackContext []interface{} // stack of node contexts for tracking CSS rules / current HTML node.
}

func (program *Program) PopRegisterStack() interface{} {
	result := program.registerStack[len(program.registerStack)-1]
	program.registerStack = program.registerStack[:len(program.registerStack)-1]
	return result
}

func ExecuteNewProgram(codeBlock *bytecode.Block) {
	program := new(Program)
	program.stack = make([]interface{}, codeBlock.StackSize)
	program.registerStack = make([]interface{}, 0, 4)

	program.executeBytecode(codeBlock)
}

func (program *Program) executeBytecode(codeBlock *bytecode.Block) {
	opcodes := codeBlock.Opcodes
	offset := 0
	for offset < len(opcodes) {
		code := opcodes[offset]

		switch code.Kind {
		case bytecode.Label:
			// no-op
		case bytecode.Push:
			program.registerStack = append(program.registerStack, code.Value)
		case bytecode.PushAllocArrayString:
			value := make([]int, 0)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocArrayInt:
			value := make([]int, 0)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocArrayFloat:
			value := make([]int, 0)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocArrayStruct:
			value := make([]bytecode.Struct, 0)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushStackVar:
			stackOffset := code.Value.(int)
			program.registerStack = append(program.registerStack, program.stack[stackOffset])
		case bytecode.PushStructFieldVar:
			fieldOffset := code.Value.(int)
			value := program.registerStack[len(program.registerStack)-1]
			structData := value.(*bytecode.Struct)
			fieldData := structData.GetField(fieldOffset)
			program.registerStack = append(program.registerStack, fieldData)
		case bytecode.ReplaceStructFieldVar:
			fieldOffset := code.Value.(int)
			value := program.registerStack[len(program.registerStack)-1]
			structData := value.(*bytecode.Struct)
			fieldData := structData.GetField(fieldOffset)
			program.registerStack[len(program.registerStack)-1] = fieldData
		case bytecode.PushAllocStruct:
			structFieldCount := code.Value.(int)
			structData := bytecode.NewStruct(structFieldCount)
			program.registerStack = append(program.registerStack, structData)
		case bytecode.PushAllocInternalStruct:
			internalType := code.Value.(reflect.Type)
			structData := reflect.Indirect(reflect.New(internalType)).Interface()
			program.registerStack = append(program.registerStack, structData)
		case bytecode.PushNewContextNode:
			panic("PushNewContextNode: Not currently supported")
			/*var node interface{}
			switch code.Value.(bytecode.NodeContextType) {
			case bytecode.NodeCSSDefinition:
				node = new(data.CSSDefinition)
			default:
				panic("Unhandled NodeContextType")
			}
			program.nodeStackContext = append(program.nodeStackContext, node)*/
		case bytecode.ConditionalEqual:
			valueA := program.registerStack[len(program.registerStack)-2].(int64)
			valueB := program.registerStack[len(program.registerStack)-1].(int64)
			program.registerStack = program.registerStack[:len(program.registerStack)-2]
			program.registerStack = append(program.registerStack, valueA == valueB)
		case bytecode.JumpIfFalse:
			boolValue := program.registerStack[len(program.registerStack)-1].(bool)
			program.registerStack = program.registerStack[:len(program.registerStack)-1]
			if !boolValue {
				offset = code.Value.(int)
				continue
			}
		case bytecode.Add:
			valueA := program.registerStack[len(program.registerStack)-2].(int64)
			valueB := program.registerStack[len(program.registerStack)-1].(int64)
			program.registerStack = program.registerStack[:len(program.registerStack)-2]
			program.registerStack = append(program.registerStack, valueA+valueB)
		case bytecode.AddString:
			valueA := program.registerStack[len(program.registerStack)-2].(string)
			valueB := program.registerStack[len(program.registerStack)-1].(string)
			program.registerStack = program.registerStack[:len(program.registerStack)-2]
			program.registerStack = append(program.registerStack, valueA+valueB)
		case bytecode.Pop:
			program.registerStack = program.registerStack[:len(program.registerStack)-1]
		case bytecode.PopN:
			popAmount := code.Value.(int)
			program.registerStack = program.registerStack[:len(program.registerStack)-popAmount]
		case bytecode.Store:
			value := program.registerStack[len(program.registerStack)-1]

			stackOffset := code.Value.(int)
			program.stack[stackOffset] = value
		case bytecode.StorePopStructField:
			fieldData := program.registerStack[len(program.registerStack)-1]
			structData := program.registerStack[len(program.registerStack)-2].(*bytecode.Struct)

			// NOTE(Jake): Only pop `fieldData`
			program.registerStack = program.registerStack[:len(program.registerStack)-1]

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
		case bytecode.Call:
			block := code.Value.(*bytecode.Block)
			program.executeBytecode(block)

			//debugPrintStack("VM Stack Values", program.stack)
			//debugPrintStack("VM Register Stack", program.registerStack)
			//panic("bytecode.Call debug")
		case bytecode.Return:
			return
		default:
			panic(fmt.Sprintf("executeBytecode: Unhandled kind in vm: \"%s\"", code.Kind.String()))
		}
		offset++
	}

	if len(program.registerStack) > 0 {
		debugPrintStack("VM Stack Values", program.stack)
		debugPrintStack("VM Register Stack", program.registerStack)
		panic("Register Stack should be empty.")
	}

	// Debug
	debugPrintStack("VM Stack Values", program.stack)
}
