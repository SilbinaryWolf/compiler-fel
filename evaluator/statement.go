package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
)

func (program *Program) evaluateStatement(topNode ast.Node, scope *Scope) {
	switch node := topNode.(type) {
	case *ast.DeclareStatement:
		name := node.Name.String()
		if _, exists := scope.GetThisScope(name); exists {
			panic(fmt.Sprintf("Cannot redeclare %v in same scope.", name))
		}
		value := program.evaluateExpression(node.ChildNodes, scope)
		scope.Set(name, value)
	default:
		panic(fmt.Sprintf("evaluateStatement(): Unhandled type: %T", node))
	}
}
