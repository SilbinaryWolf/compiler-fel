package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (program *Program) evaluateTemplate(node *ast.File, scope *Scope) []data.Type {
	program.Filepath = node.Filepath
	result := program.evaluateHTMLNodeChildren(node.Nodes(), nil, scope)
	program.Filepath = ""
	return result
}

func (program *Program) evaluateHTMLNodeChildren(nodes []ast.Node, parentNode *data.HTMLNode, scope *Scope) []data.Type {
	resultNodes := make([]data.Type, 0, 5)

	scope = NewScope(scope)
	for _, itNode := range nodes {
		switch node := itNode.(type) {
		case *ast.HTMLNode:
			subResultDataNode := program.evaluateHTMLNode(node, scope)
			resultNodes = append(resultNodes, subResultDataNode)
		case *ast.Expression:
			valueInterface := program.evaluateExpression(node, scope)
			switch value := valueInterface.(type) {
			case *data.String:
				subResultDataNode := &data.HTMLText{
					Value: value.String(),
				}
				resultNodes = append(resultNodes, subResultDataNode)
			case *data.MixedArray:
				for _, subValue := range value.Array {
					resultNodes = append(resultNodes, subValue)
				}
			case *data.HTMLNode:
				if value != nil {
					resultNodes = append(resultNodes, value)
				}
			default:
				panic(fmt.Sprintf("Unhandled value result in HTMLNode: %T", value))
			}
			/*if value.Kind() == data.KindMixedArray {
				panic("test")
			} else {
				subResultDataNode := &data.HTMLText{
					Value: value.String(),
				}
				resultNodes = append(resultNodes, subResultDataNode)
			}*/
		case *ast.CSSDefinition:
			if node.Name.Kind == token.Unknown {
				program.anonymousCSSDefinitionsUsed = append(program.anonymousCSSDefinitionsUsed, node)
			}
		default:
			program.evaluateStatement(itNode, scope)
		}
	}
	return resultNodes
}

func (program *Program) evaluateHTMLBlock(node *ast.HTMLBlock, scope *Scope) *data.HTMLNode {
	nodes := program.evaluateHTMLNodeChildren(node.Nodes(), nil, scope)
	resultNode, ok := nodes[0].(*data.HTMLNode)
	if !ok {
		panic("evaluateHTMLBlock: Failed to type-assert to data.HTMLNode")
	}
	return resultNode
}

func (program *Program) evaluateHTMLNode(node *ast.HTMLNode, scope *Scope) *data.HTMLNode {
	if node.HTMLDefinition != nil {
		return program.evaluteHTMLComponent(node, scope)
	}

	resultDataNode := new(data.HTMLNode)
	resultDataNode.Name = node.Name.String()

	// Evaluate parameters
	if parameterSet := node.Parameters; parameterSet != nil {
		for _, parameter := range parameterSet {
			value := program.evaluateExpression(&parameter.Expression, scope)
			attributeNode := data.HTMLAttribute{
				Name:  parameter.Name.String(),
				Value: value.String(),
			}
			resultDataNode.Attributes = append(resultDataNode.Attributes, attributeNode)
		}
	}

	childNodes := program.evaluateHTMLNodeChildren(node.Nodes(), resultDataNode, scope)
	resultDataNode.SetNodes(childNodes)
	return resultDataNode
}
