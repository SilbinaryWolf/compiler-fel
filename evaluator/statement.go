package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (program *Program) evaluateDeclareSet(node *ast.DeclareStatement, scope *Scope) {
	name := node.Name.String()
	if _, exists := scope.GetThisScope(name); exists {
		panic(fmt.Sprintf("Cannot redeclare %v in same scope.", name))
	}
	value := program.evaluateExpression(&node.Expression, scope)
	scope.DeclareSet(name, value)
}

func (program *Program) evaluateStatement(topNode ast.Node, scope *Scope) {
	switch node := topNode.(type) {
	case *ast.DeclareStatement:
		program.evaluateDeclareSet(node, scope)
	case *ast.OpStatement:
		name := node.Name.String()
		_, exists := scope.Get(name)
		if !exists {
			panic(fmt.Sprintf("evaluateStatement(): Typechecker missed undeclared variable \"%s\".", name))
		}
		value := program.evaluateExpression(&node.Expression, scope)
		switch node.Operator.Kind {
		case token.Equal:
			scope.Set(name, value)
		default:
			panic(fmt.Sprintf("evaluateStatement(): Unhandled set-operator: %s", node.Operator.Kind))
		}
	case *ast.HTMLBlock, *ast.HTMLComponentDefinition, *ast.StructDefinition:
		// ignore
	case *ast.CSSDefinition:
		panic("todo(Jake): Handle CSS definition in statement")
	default:
		panic(fmt.Sprintf("evaluateStatement(): Unhandled type: %T.", node))
	}
}

/*func (program *Program) evaluateFor(rootNode *ast.For, scope *Scope) {
	iValue := program.evaluateExpression(&rootNode.Array, scope)
	value := iValue.(*data.Array)

	//scope = NewScope(scope)
	name := rootNode.RecordName.String()
	nodes := rootNode.Nodes()
	for _, val := range value.Elements {
		scope.Set(name, val)

		for _, node := range nodes {
			// todo(Jake): Fix this to evaluate properly
			program.evaluateStatement(node, scope)
		}
	}
	//if _, exists := scope.GetThisScope(rootNode.RecordName.String()); exists {
	//	panic(fmt.Sprintf("Cannot redeclare %v in same scope.", name))
	//}
	panic("todo(Jake): finish for loop, currently debugging why the For-Loop has no children nodes.")
}*/
