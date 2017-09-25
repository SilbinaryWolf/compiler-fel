package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

func (program *Program) evaluteHTMLComponent(topNode *ast.HTMLNode, scope *Scope) *data.HTMLNode {
	// Get children nodes
	childNodes := program.evaluateHTMLNodeChildren(topNode.Nodes(), scope)

	var resultDataNode *data.HTMLNode
	componentScope := NewScope(nil)

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
	for _, itNode := range topNode.HTMLDefinition.ChildNodes {
		switch node := itNode.(type) {
		case *ast.HTMLNode:
			if resultDataNode != nil {
				panic(fmt.Sprintf("evaluteHTMLComponent(): Cannot return multiple html nodes from \"%s :: html\"", topNode.Name.String()))
			}
			resultDataNode = program.evaluateHTMLNode(node, componentScope)
		case *ast.HTMLProperties, *ast.CSSDefinition:
			// ignore
		default:
			program.evaluateStatement(node, componentScope)
			panic(fmt.Sprintf("evaluteHTMLComponent(): Unhandled type %T", node))
		}
	}
	if resultDataNode == nil {
		panic("evaluteHTMLComponent(): Component must contain one top-level HTML node.")
	}

	program.AddHTMLDefinitionUsed(topNode.Name.String(), topNode.HTMLDefinition, resultDataNode)
	return resultDataNode
}
