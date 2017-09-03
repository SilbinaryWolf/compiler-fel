/*package vm

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
)

func (vm *VM) allocStruct() {
	structDefinitionInterface := vm.pop()
	structDefinition, ok := structDefinitionInterface.(*StructDefinition)
	if DEBUG && !ok {
		panic(fmt.Sprintf("allocStruct: Expected *StructDefinition type not %T.", structDefinitionInterface))
	}
	if DEBUG && structDefinition == nil {
		panic(fmt.Sprintf("allocStruct: Expected not nil *StructDefinition type."))
	}

	instance := new(StructInstance)
	instance.Definition = structDefinition
	instance.Properties = make([]interface{}, len(structDefinition.defaultValues))
	copy(instance.Properties, structDefinition.defaultValues)
	vm.push(instance)
}

// Push a regular <div>, <span>, <a>, etc tag onto the return node stack.
func (vm *VM) pushNode() {
	argCountInterface := vm.pop()
	argCount, ok := argCountInterface.(int)
	if DEBUG && !ok {
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
	vm.push(htmlNode)
}
*/