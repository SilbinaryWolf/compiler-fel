package vm

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/types"
)

type Program struct {
	stack         []interface{}
	registerStack []interface{}

	//htmlNodeStack   []*data.HTMLElement
	returnHTMLNodes []*data.HTMLElement // used only in ":: html" blocks / template files
	//nodeStackContext []interface{}           // stack of node contexts for tracking CSS rules / current HTML node.
}

func ExecuteNewProgram(codeBlock *bytecode.Block) interface{} {
	program := new(Program)
	program.stack = make([]interface{}, 32)
	program.registerStack = make([]interface{}, 0, 4)

	program.executeBytecode(codeBlock)
	if codeBlock.HasReturnValue {
		return program.pop()
	}
	return nil
}

func (program *Program) pop() interface{} {
	result := program.registerStack[len(program.registerStack)-1]
	program.registerStack = program.registerStack[:len(program.registerStack)-1]
	return result
}

func (program *Program) executeBytecode(codeBlock *bytecode.Block) {
	opcodes := codeBlock.Opcodes
	offset := 0
	for offset < len(opcodes) {
		code := opcodes[offset]

		switch kind := code.Kind; kind {
		case bytecode.Label:
			// no-op
		case bytecode.Push:
			program.registerStack = append(program.registerStack, code.Value)
		case bytecode.Pop:
			program.registerStack = program.registerStack[:len(program.registerStack)-1]
		case bytecode.PopN:
			popAmount := code.Value.(int)
			program.registerStack = program.registerStack[:len(program.registerStack)-popAmount]
		//
		// Array Structures
		//
		case bytecode.PushAllocArrayString:
			capacity := code.Value.(int)
			value := make([]string, 0, capacity)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocArrayInt:
			capacity := code.Value.(int)
			value := make([]int, 0, capacity)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocArrayFloat:
			capacity := code.Value.(int)
			value := make([]float64, 0, capacity)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocArrayStruct:
			capacity := code.Value.(int)
			value := make([]data.Struct, 0, capacity)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocHTMLFragment:
			value := data.NewHTMLFragment()
			program.registerStack = append(program.registerStack, value)
		//
		// CSS Structures
		//
		/*case bytecode.PushAllocCSSDefinition:
			// NOTE(Jake): 2018-04-17
			//
			// We probably don't need to retain the "name"
			// of the CSS definition, but eh, might help with debugging
			// and other things
			//
			name := code.Value.(string)
			value := data.NewCSSDefinition(name)
			program.registerStack = append(program.registerStack, value)
		case bytecode.PushAllocCSSSelector:
			_ = program.registerStack[len(program.registerStack)-1].(*data.CSSDefinition)
			size := code.Value.(int)
			value := data.NewCSSSelector(size)
			program.registerStack = append(program.registerStack, value)*/
		/*case bytecode.AppendPopCSSSelectorPart:
		selectorPart := code.Value.(*data.CSSSelectorPart)
		selector := program.registerStack[len(program.registerStack)-1].(*data.CSSSelector)
		selector.AddPart(selectorPart)*/
		case bytecode.AppendPopArrayString:
			array := program.registerStack[len(program.registerStack)-2].([]string)
			value := program.registerStack[len(program.registerStack)-1].(string)

			// Pop value
			program.registerStack = program.registerStack[:len(program.registerStack)-1]

			array = append(array, value)
			program.registerStack[len(program.registerStack)-1] = array
		case bytecode.PushStackVar:
			stackOffset := code.Value.(int)
			program.registerStack = append(program.registerStack, program.stack[stackOffset])
		case bytecode.PushStructFieldVar:
			fieldOffset := code.Value.(int)
			value := program.registerStack[len(program.registerStack)-1]
			structData := value.(*data.Struct)
			fieldData := structData.GetField(fieldOffset)
			program.registerStack = append(program.registerStack, fieldData)
		case bytecode.ReplaceStructFieldVar:
			fieldOffset := code.Value.(int)
			value := program.registerStack[len(program.registerStack)-1]
			structData := value.(*data.Struct)
			fieldData := structData.GetField(fieldOffset)
			program.registerStack[len(program.registerStack)-1] = fieldData
		case bytecode.PushAllocStruct:
			structTypeInfo := code.Value.(*types.Struct)
			structData := data.NewStruct(len(structTypeInfo.Fields()), structTypeInfo)
			program.registerStack = append(program.registerStack, structData)
		case bytecode.PushAllocInternalStruct:
			panic("No support, to be removed.")
			/*internalType := code.Value.(reflect.Type)
			structData := reflect.Indirect(reflect.New(internalType)).Interface()
			program.registerStack = append(program.registerStack, structData)*/
		//
		// HTML Structures
		//
		case bytecode.PushAllocHTMLNode:
			tagName := code.Value.(string)
			htmlElementNode := data.NewHTMLElement(tagName)
			program.registerStack = append(program.registerStack, htmlElementNode)
		case bytecode.CastToHTMLText:
			value := program.registerStack[len(program.registerStack)-1]
			switch value := value.(type) {
			case string:
				program.registerStack[len(program.registerStack)-1] = data.NewHTMLText(value)
			default:
				panic(fmt.Sprintf("CastToHTMLText: Cannot convert from %T. This should be caught in the typechecker.", value))
			}
		case bytecode.AppendPopHTMLElementToHTMLElement:
			if len(program.registerStack) < 2 {
				panic(fmt.Sprintf("offset %d", offset))
			}
			parentNode := program.registerStack[len(program.registerStack)-2].(*data.HTMLElement)
			node := program.registerStack[len(program.registerStack)-1].(*data.HTMLElement)

			switch node.Kind() {
			case data.HTMLKindElement,
				data.HTMLKindText,
				data.HTMLKindFragment:
				node.SetParent(parentNode)
			default:
				panic(fmt.Sprintf("StoreAppendToHTMLElement: Unsupported type %v.", node.Kind()))
			}

			// Pop
			program.registerStack = program.registerStack[:len(program.registerStack)-1]
		//
		// Expressions
		//
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
		case bytecode.Store:
			value := program.registerStack[len(program.registerStack)-1]

			stackOffset := code.Value.(int)
			if stackOffset >= len(program.stack) {
				panic(fmt.Sprintf("Array out of bounds on index #%d, At opcode offset #%d", stackOffset, offset))
			}
			program.stack[stackOffset] = value
		case bytecode.StorePopHTMLAttribute:
			attrValueInterface := program.registerStack[len(program.registerStack)-1]
			node := program.registerStack[len(program.registerStack)-2].(*data.HTMLElement)

			// Only pop attribute
			program.registerStack = program.registerStack[:len(program.registerStack)-1]

			// Convert expression result into string for HTML attribute
			var attrValue string
			switch attrValueInterface := attrValueInterface.(type) {
			case string:
				attrValue = attrValueInterface
			case nil:
				// todo(Jake): 2018-01-16
				//
				// For null/nil values, we *probably* want then to mean
				// that the attribute is no longer set, but we'll see.
				//
				//attrName := code.Value.(string)
				//node.RemoveAttribute(attrName)
				panic("executeBytecode:StorePopHTMLAttribute: Add logic to handle nil")
				//continue
			default:
				panic(fmt.Sprintf("executeBytecode:StorePopHTMLAttribute: Unhandled attribute type cast for %T", attrValueInterface))
			}

			attrName := code.Value.(string)
			node.SetAttribute(attrName, attrValue)
		case bytecode.AppendPopHTMLNodeReturn:
			value := program.registerStack[len(program.registerStack)-1].(*data.HTMLElement)
			program.registerStack = program.registerStack[:len(program.registerStack)-1]

			program.returnHTMLNodes = append(program.returnHTMLNodes, value)
		case bytecode.PushReturnHTMLNodeArray:
			// NOTE(jJake): 2018-01-31
			//
			// Used for ":: html" components
			//
			returnHTMLNodes := program.returnHTMLNodes
			program.returnHTMLNodes = make([]*data.HTMLElement, 0, 10)
			program.registerStack = append(program.registerStack, returnHTMLNodes)
		case bytecode.StorePopStructField:
			fieldData := program.registerStack[len(program.registerStack)-1]
			structData := program.registerStack[len(program.registerStack)-2].(*data.Struct)

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
		case bytecode.Call, bytecode.CallHTML:
			// NOTE(Jake): 2018-01-30
			//
			// Move the stack ahead of used areas by cutting a new slice
			// and revert back so the stack can be reclaimed.
			//
			oldStack := program.stack
			program.stack = program.stack[codeBlock.StackSize:]
			if value := program.stack[0]; value != nil {
				debugPrintStack("VM Stack Values", program.stack)
				panic("bytecode.Call: Stack already has items in it, need to make sure we dont break the stack.")
			}

			block := code.Value.(*bytecode.Block)
			program.executeBytecode(block)
			// Clear for better debuggability
			for i := 0; i < block.StackSize; i++ {
				program.stack[i] = nil
			}
			program.stack = oldStack

			//if kind == bytecode.CallHTML {
			//	htmlNodes := program.pop().([]*data.HTMLElement)
			//	program.returnHTMLNodes = append(program.returnHTMLNodes, htmlNodes...)
			//}

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

	expectedRegisterStackSize := 0
	if codeBlock.HasReturnValue {
		expectedRegisterStackSize = 1
	}

	if len(program.registerStack) > expectedRegisterStackSize {
		debugPrintStack("VM Stack Values", program.stack)
		debugPrintStack(fmt.Sprintf("VM Register Stack (Size: %d)", len(program.registerStack)), program.registerStack)
		panic(fmt.Sprintf("Register Stack should have %d items.", expectedRegisterStackSize))
	}

	// Debug
	/*debugPrintStack("VM Stack Values", program.stack)
	if len(program.returnHTMLNodes) > 0 {
		fmt.Printf("Result HTML Nodes:\n")
		for _, node := range program.returnHTMLNodes {
			fmt.Printf("- %s\n", node.Debug())
		}
	}*/
}
