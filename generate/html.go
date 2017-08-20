package generate

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/util"
)

func PrettyHTML(node *data.HTMLNode) string {
	gen := new(Generator)

	if len(node.Name) > 0 {
		gen.getHTMLNode(node)
	} else {
		for _, itNode := range node.ChildNodes {
			switch childNode := itNode.(type) {
			case *data.HTMLNode:
				gen.getHTMLNode(childNode)
			default:
				panic(fmt.Sprintf("PrettyHTML(): Unhandled type: %T", itNode))
			}
		}
	}
	gen.WriteByte('\n')

	return gen.String()
}

func (gen *Generator) getHTMLNode(node *data.HTMLNode) {
	isSelfClosing := util.IsSelfClosingTagName(node.Name)

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

	if !isSelfClosing {
		if len(node.ChildNodes) > 0 {
			gen.indent--
			gen.WriteLine()
		}
		gen.WriteString("</")
		gen.WriteString(node.Name)
		gen.WriteByte('>')
	}
}
