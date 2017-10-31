package evaluator

import (
	"fmt"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
)

func PrefixNamespace(componentName string, className string) string {
	return componentName + "_" + className
}

func (program *Program) evaluateTemplate(node *ast.File) (*data.HTMLComponentNode, error) {
	result := new(data.HTMLComponentNode)
	result.Name = node.Filepath

	//
	program.Filepath = node.Filepath
	result.ChildNodes = program.evaluateHTMLNodeChildren(node.Nodes(), NewScope(nil))
	program.Filepath = ""

	if len(result.ChildNodes) == 0 {
		return nil, fmt.Errorf("Unexpected error. evaluateTemplate returned 0 nodes which should not happen if the AST is typechecked.")
	}

	// Add template
	program.AddHTMLTemplateUsed(result)

	return result, nil
}

func (program *Program) evaluateHTMLExpression(node *ast.Expression, scope *Scope, resultNodes []data.Type) []data.Type {
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
		panic(fmt.Sprintf("evaluateHTMLNodeChildren: Unhandled value result in HTMLNode: %T", value))
	}
	/*if value.Kind() == data.KindMixedArray {
		panic("test")
	} else {
		subResultDataNode := &data.HTMLText{
			Value: value.String(),
		}
		resultNodes = append(resultNodes, subResultDataNode)
	}*/
	return resultNodes
}

func (program *Program) evaluateHTMLNodeChildren(nodes []ast.Node, scope *Scope) []data.Type {
	resultNodes := make([]data.Type, 0, 5)

	scope = NewScope(scope)
	for _, itNode := range nodes {
		switch node := itNode.(type) {
		case *ast.HTMLNode:
			if node.HTMLDefinition != nil {
				subResultDataNode := program.evaluteHTMLComponent(node, scope)
				resultNodes = append(resultNodes, subResultDataNode)
				continue
			}
			subResultDataNode := program.evaluateHTMLNode(node, scope)
			resultNodes = append(resultNodes, subResultDataNode)
		case *ast.Expression:
			resultNodes = program.evaluateHTMLExpression(node, scope, resultNodes)
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

func (program *Program) evaluteHTMLComponent(topNode *ast.HTMLNode, scope *Scope) *data.HTMLComponentNode {
	// Get children nodes
	childNodes := program.evaluateHTMLNodeChildren(topNode.Nodes(), scope)

	//
	componentScope := NewScope(nil)
	program.currentComponentScope = append(program.currentComponentScope, topNode.HTMLDefinition)
	defer func() {
		program.currentComponentScope = program.currentComponentScope[:len(program.currentComponentScope)-1]
	}()

	// Get valid parameters
	if properties := topNode.HTMLDefinition.Properties; properties != nil {
		if propertySet := properties.Statements; propertySet != nil {
			for _, decl := range propertySet {
				name := decl.Name.String()
				if len(decl.ChildNodes) == 0 {
					componentScope.Set(name, program.CreateDataType(decl.TypeToken))
				} else {
					componentScope.Set(name, program.evaluateExpression(&decl.Expression, nil))
				}
			}
		}
	}

	// Evaluate parameters
	if parameterSet := topNode.Parameters; parameterSet != nil {
		for _, parameter := range parameterSet {
			name := parameter.Name.String()
			_, ok := componentScope.Get(name)
			if !ok {
				panic(fmt.Sprintf("Cannot pass \"%s\" as parameter as it's not a defined property on \"%s\".", name, topNode.Name))
			}
			value := program.evaluateExpression(&parameter.Expression, scope)
			componentScope.Set(name, value)
		}
	}

	// Add children to component scope if they exist
	if len(childNodes) > 0 {
		componentScope.Set("children", data.NewMixedArray(childNodes))
	} else {
		componentScope.Set("children", data.NewMixedArray(nil))
	}

	// Get resultDataNode
	var resultNodes []data.Type
	resultDataNode := new(data.HTMLComponentNode)
	resultDataNode.Name = topNode.HTMLDefinition.Name.String()
	for _, itNode := range topNode.HTMLDefinition.ChildNodes {
		switch node := itNode.(type) {
		case *ast.HTMLNode:
			if node.HTMLDefinition != nil {
				subResultDataNode := program.evaluteHTMLComponent(node, componentScope)
				resultNodes = append(resultNodes, subResultDataNode)
				continue
			}
			dataNode := program.evaluateHTMLNode(node, componentScope)
			resultNodes = append(resultNodes, dataNode)
		case *ast.Expression:
			resultNodes = program.evaluateHTMLExpression(node, scope, resultNodes)
		case *ast.HTMLProperties, *ast.CSSDefinition:
			// ignore
		default:
			program.evaluateStatement(node, componentScope)
			panic(fmt.Sprintf("evaluteHTMLComponent(): Unhandled type %T", node))
		}
	}
	resultDataNode.ChildNodes = resultNodes
	if len(resultDataNode.ChildNodes) == 0 {
		panic("evaluteHTMLComponent(): Component must contain one top-level HTML node.")
	}

	program.AddHTMLDefinitionUsed(topNode.Name.String(), topNode.HTMLDefinition, resultDataNode)
	return resultDataNode
}

func (program *Program) evaluateHTMLBlock(node *ast.HTMLBlock, scope *Scope) *data.HTMLNode {
	nodes := program.evaluateHTMLNodeChildren(node.Nodes(), scope)
	resultNode, ok := nodes[0].(*data.HTMLNode)
	if !ok {
		panic("evaluateHTMLBlock: Failed to type-assert to data.HTMLNode")
	}
	return resultNode
}

func (program *Program) evaluateHTMLNode(node *ast.HTMLNode, scope *Scope) *data.HTMLNode {
	resultDataNode := new(data.HTMLNode)
	resultDataNode.Name = node.Name.String()

	//
	currentComponentName := ""
	var cssConfigDefinition *ast.CSSConfigDefinition
	if currentComponentScope := program.CurrentComponentScope(); currentComponentScope != nil {
		cssConfigDefinition = currentComponentScope.CSSConfigDefinition
		currentComponentName = currentComponentScope.Name.String()
	}

	// Evaluate parameters
	if parameterSet := node.Parameters; parameterSet != nil {
		for _, parameter := range parameterSet {
			name := parameter.Name.String()
			value := program.evaluateExpression(&parameter.Expression, scope).String()
			if len(currentComponentName) > 0 && name == "class" {
				// Namespace
				classList := strings.Fields(value)
				newValue := ""
				for i, className := range classList {
					if i > 0 {
						newValue += " "
					}
					config := cssConfigDefinition.GetSettings("." + className)
					if config.Modify {
						newValue += PrefixNamespace(currentComponentName, className)
						continue
					}
					newValue += className
				}
				value = newValue
			}

			attributeNode := data.HTMLAttribute{
				Name:  name,
				Value: value,
			}
			resultDataNode.Attributes = append(resultDataNode.Attributes, attributeNode)
		}
	}

	childNodes := program.evaluateHTMLNodeChildren(node.Nodes(), scope)
	resultDataNode.SetNodes(childNodes)
	return resultDataNode
}
