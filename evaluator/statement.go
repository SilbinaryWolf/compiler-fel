package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
)

func (program *Program) evaluateDeclareSet(node *ast.DeclareStatement, scope *Scope) {
	name := node.Name.String()
	if _, exists := scope.GetThisScope(name); exists {
		panic(fmt.Sprintf("Cannot redeclare %v in same scope.", name))
	}
	value := program.evaluateExpression(node.ChildNodes, scope)
	scope.Set(name, value)
}

func (program *Program) evaluateStatement(topNode ast.Node, scope *Scope) {
	switch node := topNode.(type) {
	case *ast.DeclareStatement:
		program.evaluateDeclareSet(node, scope)
	case *ast.CSSDefinition:
		program.evaluateCSSDefinition(node, scope)
		panic("evaluateStatement(): Finish handling CSS statement")
	default:
		panic(fmt.Sprintf("evaluateStatement(): Unhandled type: %T", node))
	}
}
