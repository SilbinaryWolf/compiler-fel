package vm

import (
	"bytes"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/bytecode"
)

type DebugPrinter struct {
	bytes.Buffer
	indent int
}

func (printer *DebugPrinter) writeLine() {
	printer.WriteByte('\n')
	for i := 0; i < printer.indent; i++ {
		printer.WriteString("    ")
	}
}

func (printer *DebugPrinter) writeValue(value interface{}) {
	switch value := value.(type) {
	case *bytecode.Struct:
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

func debugPrintStack(stack []interface{}) {
	printer := new(DebugPrinter)
	fmt.Printf("----------------\nVM Stack Values:\n----------------\n")
	for i := 0; i < len(stack); i++ {
		printer.writeValue(stack[i])
		printer.writeLine()
	}
	fmt.Print(printer.String())
	fmt.Printf("----------------\n")
}
