package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (program *Program) evaluateExpression(expressionNodes []ast.Node, scope *Scope) data.Type {
	var stack []data.Type

	// todo(Jake): Rewrite string concat to use `var stringBuffer bytes.Buffer` and see if
	//			   there is a speedup

	for _, itNode := range expressionNodes {
		switch node := itNode.(type) {
		case *ast.Token:
			switch node.Kind {
			case token.String:
				value := &data.String{Value: node.String()}
				stack = append(stack, value)
			case token.Identifier:
				name := node.String()
				value, exists := scope.GetAllScopes(name)
				if !exists {
					panic(fmt.Sprintf("Variable isn't declared '%v'", name))
				}
				stack = append(stack, value)
			default:
				if node.IsOperator() {
					rightValue := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					if len(stack) == 0 {
						panic(fmt.Sprintf("Only got %s %s", rightValue, node.String()))
					}
					leftValue := stack[len(stack)-1]
					stack = stack[:len(stack)-1]

					rightType := rightValue.Kind()
					leftType := leftValue.Kind()

					switch node.Kind {
					case token.Add:
						if leftType == data.KindString && rightType == data.KindString {
							result := &data.String{
								Value: leftValue.String() + rightValue.String(),
							}
							stack = append(stack, result)
							continue
						}
						panic("evaluateExpression(): Unhandled type computation in +")
					default:
						panic(fmt.Sprintf("evaluateExpression(): Unhandled operator type: %s", node.Kind.String()))
					}
				}
				panic(fmt.Sprintf("Evaluator::evaluateExpression(): Unhandled *.astToken kind: %s", node.Kind.String()))
			}
		default:
			panic(fmt.Sprintf("Unhandled type: %T", node))
		}
	}
	if len(stack) == 0 || len(stack) > 1 {
		panic("evaluateExpression(): Invalid stack. Either 0 or above 1")
	}
	result := stack[0]

	return result
}
