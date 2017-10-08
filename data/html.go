package data

import (
	"bytes"
	"fmt"
	"strings"
)

type HTMLComponentNode struct {
	Name       string
	ChildNodes []Type
}

func (node *HTMLComponentNode) Kind() Kind {
	return KindHTMLComponentNode
}

func (node *HTMLComponentNode) String() string {
	return fmt.Sprintf("(%s :: html)", node.Name)
}

func (node *HTMLComponentNode) Nodes() []Type {
	return node.ChildNodes
}

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
	childNodes []Type

	// NOTE: Used for context where the htmlnode was processed
	// todo(Jake): Make new type, HTMLComponentNode.
	//HTMLDefinitionName string

	parentNode   *HTMLNode
	previousNode *HTMLNode
	nextNode     *HTMLNode
}

func (topNode *HTMLNode) SetNodes(nodes []Type) {
	// Apply Next()/Prev()/Parent() values
	var previousNode *HTMLNode
	nodesLength := len(nodes)
	for i := 0; i < nodesLength; i++ {
		node, ok := nodes[i].(*HTMLNode)
		if !ok {
			// Skip HTMLText, etc.
			continue
		}
		node.previousNode = previousNode
		node.parentNode = topNode

		// Get and set next node
		for j := i + 1; j < nodesLength; j++ {
			nextNode, ok := nodes[j].(*HTMLNode)
			if !ok {
				// Skip HTMLText, etc.
				continue
			}
			node.nextNode = nextNode
		}

		// Track
		previousNode = node
	}

	// Attach to node
	topNode.childNodes = nodes
}

func (node *HTMLNode) Nodes() []Type {
	return node.childNodes
}

func (node *HTMLNode) Parent() *HTMLNode {
	return node.parentNode
}

func (node *HTMLNode) Previous() *HTMLNode {
	return node.previousNode
}

func (node *HTMLNode) Next() *HTMLNode {
	return node.nextNode
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
	selectorString := selectorPart.String()
	switch selectorPart.Kind {
	case SelectorKindClass:
		selectorString = selectorString[1:]
		for _, attribute := range node.Attributes {
			if attribute.Name != "class" {
				continue
			}
			className := attribute.Value
			// NOTE(Jake): This technically has a bug, should split
			//			   into words and then check if equal
			return strings.Contains(className, selectorString)
		}
		return false
	case SelectorKindID:
		selectorString = selectorString[1:]
		for _, attribute := range node.Attributes {
			if attribute.Name != "id" {
				continue
			}
			ID := attribute.Value
			return ID == selectorString
		}
		return false
	case SelectorKindAttribute:
		attributeName := selectorPart.Name
		attributeValue := selectorPart.Value
		for _, attribute := range node.Attributes {
			if attribute.Name != attributeName {
				continue
			}
			switch selectorPart.Operator {
			case "=":
				return attributeValue == attribute.Value
			default:
				panic(fmt.Sprintf("HasSelectorPartMatch: Unhandled attribute operator: %s", selectorPart.Operator))
			}
			return false
		}
	case SelectorKindTag:
		return node.Name == selectorString
	default:
		panic(fmt.Sprintf("HasSelectorPartMatch: Unhandled selector part kind: %s", selectorPart.Kind.String()))
	}
	return false
}

func (topNode *HTMLComponentNode) HasMatchRecursive(selectorParts CSSSelector, htmlDefinitionName string) bool {
	nodeIterationStack := make([]*HTMLNode, 0, 50)
	childNodes := topNode.Nodes()
	for i := len(childNodes) - 1; i >= 0; i-- {
		itNode, ok := childNodes[i].(*HTMLNode)
		if !ok {
			continue
		}
		nodeIterationStack = append(nodeIterationStack, itNode)
	}

	lastSelectorPart := &selectorParts[len(selectorParts)-1]
	if lastSelectorPart.Kind != SelectorKindAttribute &&
		!lastSelectorPart.Kind.IsIdentifier() {
		panic(fmt.Sprintf("HTMLNode::HasMatchRecursive(): Unhandled type \"%s\" in selector \"%s\"", lastSelectorPart.Kind.String(), selectorParts.String()))
	}
	fmt.Printf("Search for selector - \"%s\"\n\n", lastSelectorPart.String())
	//fmt.Printf("Selector - %s - Lastbit - %s\n", selectorParts, itLastSelectorPart)

NodeLoop:
	for len(nodeIterationStack) > 0 {
		node := nodeIterationStack[len(nodeIterationStack)-1]
		nodeIterationStack = nodeIterationStack[:len(nodeIterationStack)-1]

		// Skip nodes that weren't created by the specified HTMLComponentDefinition
		/*if len(node.HTMLDefinitionName) > 0 &&
			len(htmlDefinitionName) > 0 &&
			htmlDefinitionName != node.HTMLDefinitionName {
			continue
		}*/

		//
		if node.HasSelectorPartMatch(lastSelectorPart) {
			if len(selectorParts) == 1 {
				return true
			}

			currentNode := node
		SelectorPartMatchingLoop:
			for p := len(selectorParts) - 2; p >= 0; p-- {
				//
				if selectorPart := &selectorParts[p]; selectorPart.Kind == SelectorKindAttribute ||
					selectorPart.Kind.IsIdentifier() {
					if !currentNode.HasSelectorPartMatch(selectorPart) {
						continue NodeLoop
					}
					continue SelectorPartMatchingLoop
				}

				// If wasn't identifier, handle operator
				selectorPartOperator := &selectorParts[p]
				p--
				if p < 0 {
					panic(fmt.Sprintf("Missing identifier before %s.", selectorPartOperator.Kind.String()))
					continue NodeLoop
				}
				selectorPart := &selectorParts[p]
				if !selectorPart.Kind.IsIdentifier() {
					panic(fmt.Sprintf("Expected selector identifier, not \"%s\"", selectorPartOperator.Kind))
					continue NodeLoop
				}

				switch selectorPartOperator.Kind {
				case SelectorKindAncestor:
					// Has matched!
					//continue SelectorPartMatchingLoop
					for {
						currentNode = currentNode.Parent()
						if currentNode == nil {
							continue NodeLoop
						}
						if !currentNode.HasSelectorPartMatch(selectorPart) {
							continue
						}
						break
					}
					// Has matched!
					continue SelectorPartMatchingLoop
				case SelectorKindChild:
					currentNode = currentNode.Parent()
					if currentNode == nil {
						continue NodeLoop
					}
					if !currentNode.HasSelectorPartMatch(selectorPart) {
						continue NodeLoop
					}
					// Has matched!
					continue SelectorPartMatchingLoop
				case SelectorKindAdjacent:
					currentNode = currentNode.Previous()
					if currentNode == nil {
						continue NodeLoop
					}
					if !currentNode.HasSelectorPartMatch(selectorPart) {
						continue NodeLoop
					}
					continue SelectorPartMatchingLoop
				case SelectorKindSibling:
					for {
						currentNode = currentNode.Previous()
						if currentNode == nil {
							continue NodeLoop
						}
						if !currentNode.HasSelectorPartMatch(selectorPart) {
							continue
						}
						break
					}
					// Has matched!
					continue SelectorPartMatchingLoop
					//panic("todo(Jake): Sibling selector matching")
				default:
					panic(fmt.Sprintf("HTMLNode::HasMatchRecursive():inner: Unhandled type \"%s\"", selectorPart.Kind.String()))
				}
				continue NodeLoop
			}

			// If got to end of loop, then it matched!
			return true
		}

		// Add children
		childNodes := node.Nodes()
		for i := len(childNodes) - 1; i >= 0; i-- {
			node, ok := childNodes[i].(*HTMLNode)
			if !ok {
				continue
			}
			nodeIterationStack = append(nodeIterationStack, node)
		}
	}
	return false
}
