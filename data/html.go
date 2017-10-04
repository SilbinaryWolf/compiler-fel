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
	// todo(Jake): Make new type, HTMLComponentNode.
	HTMLDefinitionName string

	//parentIndex int
	//parentNode *HTMLNode
}

type HTMLAttribute struct {
	Name  string
	Value string
}

func (node *HTMLNode) Kind() Kind {
	return KindHTMLNode
}

//func (node *HTMLNode) Parent() *HTMLNode {
//	return node.parentNode
//}

/*func (node *HTMLNode) AsSelectorString() string {
	result := ""
	result = node.Name
	for _, attribute := range node.Attributes {
		if attribute.Name != "class" {
			continue
		}
		result += "." + attribute.Value
	}
	return result
}*/

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

func (topNode *HTMLNode) HasMatchRecursive(selectorParts CSSSelector, htmlDefinitionName string) bool {
	nodeIterationStack := make([]*HTMLNode, 0, 50)
	nodeIterationStack = append(nodeIterationStack, topNode)

	// This is for matching parents of the node being iterated on.
	nodeScopeStack := make([]*HTMLNode, 0, 20)

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

		if node == nil {
			nodeScopeStack = nodeScopeStack[:len(nodeScopeStack)-1]
			continue
		}

		// Skip nodes that weren't created by the specified HTMLComponentDefinition
		if len(node.HTMLDefinitionName) > 0 &&
			len(htmlDefinitionName) > 0 &&
			htmlDefinitionName != node.HTMLDefinitionName {
			continue
		}

		// Add scope
		nodeScopeStack = append(nodeScopeStack, node)
		nodeIterationStack = append(nodeIterationStack, nil)

		//
		if node.HasSelectorPartMatch(lastSelectorPart) {
			if len(selectorParts) == 1 {
				return true
			}

			// NOTE(Jake): We only want to modify this slice within the context
			//			   of selector matching, not across all nodes we're checking.
			innerNodeScopeStack := nodeScopeStack
		SelectorPartMatchingLoop:
			for p := len(selectorParts) - 2; p >= 0; p-- {
				if len(innerNodeScopeStack) == 0 {
					break
				}
				currentNode := innerNodeScopeStack[len(innerNodeScopeStack)-1]
				innerNodeScopeStack = innerNodeScopeStack[:len(innerNodeScopeStack)-1]

				//
				selectorPart := &selectorParts[p]
				switch selectorPart.Kind {
				case SelectorKindAncestor:
					p--
					if p < 0 {
						panic("Missing identifier before [descendant whitespace].")
						continue NodeLoop
					}
					selectorPart = &selectorParts[p]
					if !selectorPart.Kind.IsIdentifier() {
						panic(fmt.Sprintf("Expected selector identifier, not \"%s\"", selectorPart.Kind))
					}
					fmt.Printf("- Does descendent selector part -- %s\n", selectorPart.String())
					//{
					//	fmt.Printf("...match stack:\n-----\n")
					//	for i := len(innerNodeScopeStack) - 1; i >= 0; i-- {
					//		node := innerNodeScopeStack[i]
					//		fmt.Printf("- %s\n", node.String())
					//	}
					//}
					for i := len(innerNodeScopeStack) - 1; i >= 0; i-- {
						node := innerNodeScopeStack[i]
						fmt.Printf("...match %s\n", node.String())
						if !node.HasSelectorPartMatch(selectorPart) {
							innerNodeScopeStack = innerNodeScopeStack[:i]
							continue
						}
						// Has matched!
						//if p == 0 {
						//	return true
						//}

						continue SelectorPartMatchingLoop
					}
				case SelectorKindChild:
					// Get previous selector part
					p--
					if p < 0 {
						panic("Missing identifier before >.")
						continue NodeLoop
					}
					selectorPart = &selectorParts[p]
					if !selectorPart.Kind.IsIdentifier() {
						panic(fmt.Sprintf("Expected selector identifier, not \"%s\"", selectorPart.Kind))
					}
					if len(innerNodeScopeStack) == 0 {
						panic("??? This should not happen maybe?")
						continue NodeLoop
					}
					node := innerNodeScopeStack[len(innerNodeScopeStack)-1]
					//innerNodeScopeStack = innerNodeScopeStack[:len(innerNodeScopeStack)-1]
					fmt.Printf("- Does node: %s, match part: %s \n", node.String(), selectorPart.String())
					if !node.HasSelectorPartMatch(selectorPart) {
						continue NodeLoop
					}
					// Has matched!
					//if p == 0 {
					//	return true
					//}
					continue SelectorPartMatchingLoop
				default:
					if selectorPart.Kind == SelectorKindAttribute ||
						selectorPart.Kind.IsIdentifier() {
						if !currentNode.HasSelectorPartMatch(selectorPart) {
							continue NodeLoop
						}
						continue SelectorPartMatchingLoop
					}
					panic(fmt.Sprintf("HTMLNode::HasMatchRecursive():inner: Unhandled type \"%s\"", selectorPart.Kind.String()))
				}
				continue NodeLoop
			}

			// If got to end of loop, success!
			return true
			//panic(fmt.Sprintf("todo(Jake): Handle multiple selector - %s", selectorParts.String()))
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
			nodeIterationStack = append(nodeIterationStack, node)
		}
	}
	return false
}
