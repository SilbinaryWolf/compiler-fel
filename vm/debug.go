package vm

import (
	"bytes"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/data"
)

type DebugPrinter struct {
	bytes.Buffer
	indent      int
	seenPointer map[string]bool
}

func (printer *DebugPrinter) writeLine() {
	printer.WriteByte('\n')
	for i := 0; i < printer.indent; i++ {
		printer.WriteString("    ")
	}
}

func debugOpcodes(opcodes []bytecode.Code, offset int) {
	fmt.Printf("Opcode Debug:\n-----------\n")
	for i, code := range opcodes {
		if offset == i {
			fmt.Printf("**%d** - %s\n", i, code.String())
			continue
		}
		fmt.Printf("%d - %s\n", i, code.String())
	}
	fmt.Printf("-----------\n")
}

func (printer *DebugPrinter) writeValue(value interface{}) {
	switch value := value.(type) {
	case *data.HTMLElement:
		printer.WriteString(value.Debug())
	case *data.Struct:
		addr := fmt.Sprintf("%p", value)
		if _, ok := printer.seenPointer[addr]; ok {
			printer.WriteString("(--recursive--)")
			return
		}
		printer.seenPointer[addr] = true

		printer.indent++
		printer.WriteString(fmt.Sprintf("%T w/ %d fields", value, value.FieldCount()))
		printer.WriteString("(")
		printer.writeLine()
		fieldCount := value.FieldCount()
		for i := 0; i < fieldCount; i++ {
			if i != 0 {
				printer.writeLine()
			}
			field := value.GetField(i)
			printer.writeValue(field)
			printer.WriteString(",")
		}
		printer.indent--
		printer.writeLine()
		printer.WriteString(")")
		//printer.WriteString(fmt.Sprintf("%T w/ %d fields", value, value.FieldCount()))
	case string:
		printer.WriteString(fmt.Sprintf("\"%s\" - %T", value, value))
	default:
		printer.WriteString(fmt.Sprintf("%v - %T", value, value))
	}
}

func debugPrintStack(message string, stack []interface{}) {
	printer := new(DebugPrinter)
	printer.seenPointer = make(map[string]bool)
	fmt.Printf("----------------\n%s:\n----------------\n", message)
	defer func() {
		fmt.Print(printer.String())
		fmt.Printf("----------------\n")
	}()
	for i := 0; i < len(stack); i++ {
		printer.writeValue(stack[i])
		printer.writeLine()

		// Reset per top level value
		printer.seenPointer = make(map[string]bool)
	}
}
