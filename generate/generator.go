package generate

import "bytes"

type Generator struct {
	bytes.Buffer
	indent int
}

func (gen *Generator) WriteLine() {
	gen.WriteByte('\n')
	for i := 0; i < gen.indent; i++ {
		gen.WriteString("    ")
	}
}
