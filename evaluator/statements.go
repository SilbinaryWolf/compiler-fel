package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
)

func (program *Program) evaluateStatements(nodesQueue []ast.Node, scope *Scope) {
	// DEBUG
	/*json, _ := json.MarshalIndent(nodesQueue, "", "   ")
	fmt.Printf("%s", string(json))
	panic("evaluateBlock")*/

	for len(nodesQueue) > 0 {
		currentNode := nodesQueue[0]
		nodesQueue = nodesQueue[1:]

		switch node := currentNode.(type) {
		case *ast.DeclareStatement:
			name := node.Name.String()
			if _, exists := scope.GetCurrentScope(name); exists {
				panic(fmt.Sprintf("Cannot redeclare %v", name))
			}
			value := program.evaluateExpression(node.ChildNodes, scope)
			scope.Set(name, value)
		default:
			panic(fmt.Sprintf("evaluateBlock(): Unhandled type: %T", node))
		}

		// Add children
		// NOTE(Jake): I only want this on ast.Block or similar, NOT ast.DeclareStatement
		//			   as that will just be buggy / odd behaviour
		//nodesQueue = append(nodesQueue, currentNode.Nodes()...)
	}
	//panic("Evaluator::evaluateBlock(): todo(Jake): The rest of the function")
}
