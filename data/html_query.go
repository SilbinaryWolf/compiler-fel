package data

import (
	"fmt"
	"strings"
)

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
		selectorValue := selectorPart.Value
		for _, attribute := range node.Attributes {
			if attribute.Name != selectorPart.Name {
				continue
			}
			switch selectorPart.Operator {
			case "=":
				return attribute.Value == selectorValue
			case "^=":
				return strings.HasPrefix(attribute.Value, selectorValue)
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

func (topNode *HTMLComponentNode) QuerySelectorAll(selectorParts CSSSelector) []*HTMLNode {
	nodeResultStack := make([]*HTMLNode, 0, 5)

	// Get nodes
	nodes := topNode.Nodes()
	for _, itNode := range nodes {
		switch node := itNode.(type) {
		case *HTMLNode:
			foundNodes := node.querySelectorAllWithOptions(selectorParts, true)
			if foundNodes != nil {
				nodeResultStack = append(nodeResultStack, foundNodes...)
			}
		case *HTMLComponentNode:
			foundNodes := node.QuerySelectorAll(selectorParts)
			if foundNodes != nil {
				nodeResultStack = append(nodeResultStack, foundNodes...)
			}
		case *HTMLText:
			// skip
		default:
			panic(fmt.Sprintf("QuerySelectorAll: Unknown type %T", node))
		}
	}

	return nodeResultStack
}

func (topNode *HTMLNode) QuerySelectorAll(selectorParts CSSSelector) []*HTMLNode {
	return topNode.querySelectorAllWithOptions(selectorParts, false)
}

func (topNode *HTMLNode) querySelectorAllWithOptions(selectorParts CSSSelector, onlyScanCurrentHTMLComponentScope bool) []*HTMLNode {
	nodeResultStack := make([]*HTMLNode, 0, 5)
	nodeIterationStack := make([]*HTMLNode, 0, 50)

	//
	{
		childNodes := topNode.Nodes()
		for i := len(childNodes) - 1; i >= 0; i-- {
			switch node := childNodes[i].(type) {
			case *HTMLNode:
				nodeIterationStack = append(nodeIterationStack, node)
			case *HTMLComponentNode:
				if onlyScanCurrentHTMLComponentScope {
					continue
				}
				panic("todo(Jake): Add HTMLNode items from HTMLComponentNode")
				// todo(Jake):
				//nodeIterationStack = append(nodeIterationStack, node.Nodes())
			case *HTMLText:
				// skip
			default:
				panic(fmt.Sprintf("QuerySelectorAll: Unhandled type: %T", node))
			}
		}
		nodeIterationStack = append(nodeIterationStack, topNode)
	}

	lastSelectorPart := &selectorParts[len(selectorParts)-1]
	if lastSelectorPart.Kind != SelectorKindAttribute &&
		!lastSelectorPart.Kind.IsIdentifier() {
		panic(fmt.Sprintf("HTMLNode::HasMatchRecursive(): Unhandled type \"%s\" in selector \"%s\"", lastSelectorPart.Kind.String(), selectorParts.String()))
	}

NodeRecursionLoop:
	for len(nodeIterationStack) > 0 {
		node := nodeIterationStack[len(nodeIterationStack)-1]
		nodeIterationStack = nodeIterationStack[:len(nodeIterationStack)-1]

		//
		if node.HasSelectorPartMatch(lastSelectorPart) {
			currentNode := node

		SelectorPartMatchingLoop:
			for p := len(selectorParts) - 1; p >= 0; p-- {
				//
				if selectorPart := &selectorParts[p]; selectorPart.Kind == SelectorKindAttribute ||
					selectorPart.Kind.IsIdentifier() {
					if !currentNode.HasSelectorPartMatch(selectorPart) {
						continue NodeRecursionLoop
					}
					continue SelectorPartMatchingLoop
				}

				// If wasn't identifier, handle operator
				selectorPartOperator := &selectorParts[p]
				p--
				if p < 0 {
					panic(fmt.Sprintf("Missing identifier before %s.", selectorPartOperator.Kind.String()))
					continue NodeRecursionLoop
				}
				selectorPart := &selectorParts[p]
				if !selectorPart.Kind.IsIdentifier() {
					panic(fmt.Sprintf("Expected selector identifier, not \"%s\"", selectorPartOperator.Kind))
					continue NodeRecursionLoop
				}

				switch selectorPartOperator.Kind {
				case SelectorKindAncestor:
					// Has matched!
					//continue SelectorPartMatchingLoop
					for {
						currentNode = currentNode.Parent()
						if currentNode == nil {
							continue NodeRecursionLoop
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
						continue NodeRecursionLoop
					}
					if !currentNode.HasSelectorPartMatch(selectorPart) {
						continue NodeRecursionLoop
					}
					// Has matched!
					continue SelectorPartMatchingLoop
				case SelectorKindAdjacent:
					currentNode = currentNode.Previous()
					if currentNode == nil {
						continue NodeRecursionLoop
					}
					if !currentNode.HasSelectorPartMatch(selectorPart) {
						continue NodeRecursionLoop
					}
					continue SelectorPartMatchingLoop
				case SelectorKindSibling:
					for {
						currentNode = currentNode.Previous()
						if currentNode == nil {
							continue NodeRecursionLoop
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
				continue NodeRecursionLoop
			}

			// If got to end of loop, then it matched!
			nodeResultStack = append(nodeResultStack, node)
		}

		// Add children
		childNodes := node.Nodes()
		for i := len(childNodes) - 1; i >= 0; i-- {
			switch node := childNodes[i].(type) {
			case *HTMLNode:
				nodeIterationStack = append(nodeIterationStack, node)
			case *HTMLComponentNode:
				if onlyScanCurrentHTMLComponentScope {
					continue
				}
				panic("todo(Jake): Add HTMLNode items from HTMLComponentNode")
			case *HTMLText:
				// skip
			default:
				panic(fmt.Sprintf("HasMatchRecursive()::loop: Unhandled type: %T", node))
			}
		}
	}

	if len(nodeResultStack) == 0 {
		return nil
	}
	return nodeResultStack
}
