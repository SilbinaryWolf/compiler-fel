package generate

import (
	"bytes"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/util"
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
	isNamedHTMLNode := len(node.Name) > 0
	isSelfClosing := util.IsSelfClosingTagName(node.Name)

	if isNamedHTMLNode {
		gen.WriteByte('<')
		gen.WriteString(node.Name)
		for _, attribute := range node.Attributes {
			gen.WriteByte(' ')
			gen.WriteString(attribute.Name)
			gen.WriteString("=\"")
			gen.WriteString(attribute.Value)
			gen.WriteByte('"')
		}
		if isSelfClosing {
			gen.WriteByte('/')
		}
		gen.WriteByte('>')

		if !isSelfClosing && len(node.ChildNodes) > 0 {
			gen.indent++
		}
	}

	if len(node.ChildNodes) == 0 && !isSelfClosing {
		gen.WriteLine()
	} else {
		for _, itNode := range node.ChildNodes {
			gen.WriteLine()
			switch subNode := itNode.(type) {
			case *data.HTMLNode:
				gen.getHTMLNode(subNode)
			case *data.HTMLText:
				gen.WriteString(subNode.String())
			default:
				panic(fmt.Sprintf("getHTMLNode(): Unhandled type: %T", itNode))
			}
		}
	}

	if !isSelfClosing && isNamedHTMLNode {
		if len(node.ChildNodes) > 0 {
			gen.indent--
			gen.WriteLine()
		}
		gen.WriteString("</")
		gen.WriteString(node.Name)
		gen.WriteByte('>')
	}
}
