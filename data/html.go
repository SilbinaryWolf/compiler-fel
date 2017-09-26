package data

import (
	"bytes"
	"fmt"
	"strings"
)

type HTMLText struct {
	Value string
}

func (node *HTMLText) Kind() Kind {
	return KindHTMLText
}

func (node *HTMLText) String() string {
	return node.Value
}

type HTMLNode struct {
	Name       string
	Attributes []HTMLAttribute
	ChildNodes []Type

	// NOTE: Used for context where the htmlnode was processed
	HTMLDefinitionName string
}

type HTMLAttribute struct {
	Name  string
	Value string
}

func (node *HTMLNode) Kind() Kind {
	return KindHTMLNode
}

func (node *HTMLNode) String() string {
	var buffer bytes.Buffer
	buffer.WriteByte('<')
	buffer.WriteString(node.Name)
	buffer.WriteByte(' ')
	for _, attribute := range node.Attributes {
		buffer.WriteString(attribute.Name)
		buffer.WriteString("=\"")
		buffer.WriteString(attribute.Value)
		buffer.WriteString("\" ")
	}
	buffer.WriteByte('>')
	return buffer.String()
}

func (node *HTMLNode) HasSelectorPartMatch(selectorPart *CSSSelectorPart) bool {
	switch selectorPart.Kind {
	case SelectorKindIdentifier:
		selectorString := selectorPart.String()
		switch selectorString[0] {
		case '.':
			selectorString = selectorString[1:]
			for _, attribute := range node.Attributes {
				if attribute.Name != "class" {
					continue
				}
				className := attribute.Value
				return strings.Contains(className, selectorString)
			}
		case '#':
			panic("todo(Jake): Handle #")
		default:
			return node.Name == selectorString
		}
	default:
		panic(fmt.Sprintf("HasSelectorPartMatch: Unhandled selector part kind: %d", selectorPart.Kind))
	}
	return false
}

func (topNode *HTMLNode) HasMatchRecursive(selectorParts CSSSelector, htmlDefinitionName string) bool {
	nodeStack := make([]*HTMLNode, 0, 50)
	nodeStack = append(nodeStack, topNode)

	lastSelectorPart := &selectorParts[len(selectorParts)-1]
	//fmt.Printf("Selector - %s - Lastbit - %s\n", selectorParts, itLastSelectorPart)

	for len(nodeStack) > 0 {
		node := nodeStack[len(nodeStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]

		// Skip nodes that weren't created by the specified HTMLComponentDefinition
		if len(node.HTMLDefinitionName) > 0 &&
			len(htmlDefinitionName) > 0 &&
			htmlDefinitionName != node.HTMLDefinitionName {
			continue
		}

		switch lastSelectorPart.Kind {
		case SelectorKindIdentifier:
			if node.HasSelectorPartMatch(lastSelectorPart) {
				if len(selectorParts) == 1 {
					return true
				}
				for i := len(selectorParts) - 2; i >= 0; i++ {
					selectorPart := &selectorParts[i]
					if node.HasSelectorPartMatch(selectorPart) {
						continue
					}

					// If no matches, stop
					break
				}
				panic(fmt.Sprintf("todo(Jake): Handle multiple selector - %s", selectorParts.String()))
			}
		default:
			panic(fmt.Sprintf("HTMLNode::HasMatchRecursive(): Unhandled type %T", lastSelectorPart))
		}
		// fmt.Printf("Tag - %s", node.Name)
		// for _, attribute := range node.Attributes {
		// 	switch attribute.Name {
		// 	case "class":
		// 		fmt.Printf(" - Class - %s", attribute.Value)
		// 	}
		// }
		// fmt.Printf("\n")

		// Add children
		childNodes := node.ChildNodes
		for i := len(childNodes) - 1; i >= 0; i-- {
			node, ok := childNodes[i].(*HTMLNode)
			if !ok {
				continue
			}
			nodeStack = append(nodeStack, node)
		}
	}
	return false
}
