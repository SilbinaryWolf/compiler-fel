package generate

import (
	"bytes"

	"github.com/silbinarywolf/compiler-fel/data"
)

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

func PrettyHTML(node *data.HTMLNode) string {
	gen := new(Generator)

	gen.getHTMLNode(node)

	return gen.String()
}

func (gen *Generator) getHTMLNode(node *data.HTMLNode) {
	gen.WriteByte('<')
	gen.WriteString(node.Name)
	for _, attribute := range node.Attributes {
		gen.WriteByte(' ')
		gen.WriteString(attribute.Name)
		gen.WriteString("=\"")
		gen.WriteString(attribute.Value)
		gen.WriteByte('"')
	}
	gen.WriteByte('>')

	if len(node.ChildNodes) > 0 {
		gen.indent++
		for _, subNode := range node.ChildNodes {
			gen.WriteLine()
			gen.getHTMLNode(subNode)
		}
	} else {
		gen.WriteLine()
	}

	if len(node.ChildNodes) > 0 {
		gen.indent--
		gen.WriteLine()
	}
	gen.WriteString("</")
	gen.WriteString(node.Name)
	gen.WriteByte('>')
}
