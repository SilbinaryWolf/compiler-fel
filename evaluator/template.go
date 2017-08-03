package evaluator

import (
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
)

func (program *Program) evaluateTemplate(nodesQueue []ast.Node, scope *Scope) []*HTMLNode {
	var resultDataNodeSet []*HTMLNode
	program.debugLevel++

	for len(nodesQueue) > 0 {
		currentNode := nodesQueue[0]
		nodesQueue = nodesQueue[1:]

		{
			json, _ := json.MarshalIndent(nodesQueue, "", "   ")
			fmt.Printf("%s", string(json))
		}

		switch node := currentNode.(type) {
		case *ast.HTMLNode:
			resultDataNode := new(HTMLNode)
			resultDataNode.Name = node.Name.String()

			// Evaluate parameters
			if parameterSet := node.Parameters; parameterSet != nil {
				for _, parameter := range parameterSet {
					value := program.evaluateExpression(parameter.Nodes(), scope)
					attributeNode := HTMLAttribute{
						Name:  parameter.Name.String(),
						Value: value.String(),
					}
					resultDataNode.Attributes = append(resultDataNode.Attributes, attributeNode)
				}
			}

			// Add child nodes
			subScope := NewScope(scope)
			fmt.Printf("- (Level %d) Getting children of %s.\n", program.debugLevel, resultDataNode.Name)
			for _, subNode := range node.Nodes() {
				// Debug
				/*for _, subSubNode := range subNode.Nodes() {
					switch node_test := subSubNode.(type) {
					case *ast.HTMLNode:
						fmt.Printf("-- %T -- %s \n", subSubNode, node_test.Name)
					default:
						panic(fmt.Sprintf("evaluateTemplate(): Unhandled ast type: %T", node_test))
					}
				}*/
				subResultDataNodeSet := program.evaluateTemplate(subNode.Nodes(), subScope)
				if subResultDataNodeSet != nil {
					for _, subResultDataNode := range subResultDataNodeSet {
						resultDataNode.ChildNodes = append(resultDataNode.ChildNodes, subResultDataNode)
					}
				}
			}

			resultDataNodeSet = append(resultDataNodeSet, resultDataNode)
		default:
			panic("OY")
			//program.evaluateStatements(nodesQueue, scope)
			break
		}
	}

	program.debugLevel--

	return resultDataNodeSet
}
