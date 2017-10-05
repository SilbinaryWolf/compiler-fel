package generate

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/util"
)

func PrettyHTML(nodes []data.Type) string {
	gen := new(Generator)

	for _, itNode := range nodes {
		switch childNode := itNode.(type) {
		case *data.HTMLNode:
			gen.WriteHTMLNode(childNode)
		case *data.HTMLText:
			gen.WriteString(childNode.String())
		default:
			panic(fmt.Sprintf("PrettyHTML(): Unhandled type: %T", itNode))
		}
	}

	gen.WriteByte('\n')

	return gen.String()
}

func (gen *Generator) WriteHTMLNode(node *data.HTMLNode) {
	isSelfClosing := util.IsSelfClosingTagName(node.Name)
	childNodes := node.Nodes()

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

	if !isSelfClosing && len(childNodes) > 0 {
		gen.indent++
	}

	if len(childNodes) == 0 && !isSelfClosing {
		gen.WriteLine()
	} else {
		for _, itNode := range childNodes {
			gen.WriteLine()
			switch childNode := itNode.(type) {
			case *data.HTMLNode:
				gen.WriteHTMLNode(childNode)
			case *data.HTMLText:
				gen.WriteString(childNode.String())
			default:
				panic(fmt.Sprintf("getHTMLNode(): Unhandled type: %T", itNode))
			}
		}
	}

	if !isSelfClosing {
		if len(childNodes) > 0 {
			gen.indent--
			gen.WriteLine()
		}
		gen.WriteString("</")
		gen.WriteString(node.Name)
		gen.WriteByte('>')
	}
}
