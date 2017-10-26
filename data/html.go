package data

import (
	"bytes"
	"fmt"
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
			// Skip HTMLText, HTMLComponentNode etc.
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
