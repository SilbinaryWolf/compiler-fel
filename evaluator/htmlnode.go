package evaluator

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

func (program *Program) evaluateTemplate(node *ast.File, scope *Scope) *data.HTMLNode {
	htmlNode := ast.HTMLNode{}
	htmlNode.ChildNodes = node.Nodes()
	result := program.evaluateHTMLNode(&htmlNode, scope)
	result.Name = ""
	return result
}

func (program *Program) evaluateHTMLNode(node *ast.HTMLNode, scope *Scope) *data.HTMLNode {
	resultDataNode := new(data.HTMLNode)
	resultDataNode.Name = node.Name.String()

	// Evaluate parameters
	if parameterSet := node.Parameters; parameterSet != nil {
		for _, parameter := range parameterSet {
			value := program.evaluateExpression(parameter.Nodes(), scope)
			attributeNode := data.HTMLAttribute{
				Name:  parameter.Name.String(),
				Value: value.String(),
			}
			resultDataNode.Attributes = append(resultDataNode.Attributes, attributeNode)
		}
	}

	scope = NewScope(scope)
	for _, itNode := range node.Nodes() {
		switch node := itNode.(type) {
		case *ast.HTMLNode:
			subResultDataNode := program.evaluateHTMLNode(node, scope)
			resultDataNode.ChildNodes = append(resultDataNode.ChildNodes, subResultDataNode)
		case *ast.Expression:
			value := program.evaluateExpression(node.ChildNodes, scope)
			subResultDataNode := &data.HTMLText{
				Value: value.String(),
			}
			resultDataNode.ChildNodes = append(resultDataNode.ChildNodes, subResultDataNode)
		default:
			program.evaluateStatement(itNode, scope)
		}
	}
	return resultDataNode
}
