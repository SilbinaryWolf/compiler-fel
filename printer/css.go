package printer

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
)

func PrettyCSS(node *data.CSSDefinition) string {
	gen := new(Generator)
	for _, itNode := range node.Rules() {
		gen.WriteCSSRuleNode(itNode)
	}
	gen.WriteByte('\n')
	return gen.String()
}

func (gen *Generator) WriteCSSRuleNode(node *data.CSSRule) {
	selectors := node.Selectors()
	if len(selectors) == 0 {
		panic("getCSSRuleNode(): CSSRule with no selectors???")
	}

	// Print selectors
	for i, selectorNodes := range selectors {
		if i != 0 {
			gen.WriteByte(',')
			gen.WriteLine()
		}

		//lastSelectorWasOperator := false
		for _, node := range selectorNodes {
			switch nodeKind := node.Kind(); nodeKind {
			case data.SelectorPartKindAttribute:
				//if i != 0 && lastSelectorWasOperator == false {
				//	gen.WriteByte(' ')
				//}
				gen.WriteByte('[')
				gen.WriteString(node.Name())
				if node.Operator() != "" {
					gen.WriteString(node.Operator())
					gen.WriteByte('"')
					gen.WriteString(node.Value())
					gen.WriteByte('"')
				}
				gen.WriteByte(']')
			// todo(Jake): Fix this, this is used for paren'd component values. ie ([controls])
			/*case data.CSSSelector:
			if i != 0 && lastSelectorWasOperator == false {
				gen.WriteByte(' ')
			}
			gen.WriteByte('(')
			gen.WriteString(node.String())
			gen.WriteByte(')')
			//panic(fmt.Sprintf("getCSSRuleNode(): Unhandled node type: %T, value: %s", node, node.String()))*/
			default:
				if nodeKind.IsOperator() ||
					nodeKind.IsIdentifier() {
					gen.WriteString(node.String())
					continue
				}
				panic(fmt.Sprintf("getCSSRuleNode(): Unhandled node type: %T", node))
			}
		}
	}
	gen.WriteByte(' ')
	gen.WriteByte('{')
	gen.indent++
	gen.WriteLine()

	// Print properties
	for i, property := range node.Properties() {
		if i != 0 {
			gen.WriteLine()
		}
		gen.WriteString(property.String())
	}

	// Print nested rules
	for i, rule := range node.Rules() {
		if i != 0 {
			gen.WriteLine()
		}
		gen.WriteCSSRuleNode(rule)
	}

	gen.indent--
	gen.WriteLine()
	gen.WriteByte('}')
	gen.WriteLine()
}

/*func (gen *Generator) getHTMLNode(node *data.HTMLNode) {
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
}*/
