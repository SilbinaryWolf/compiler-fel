package evaluator

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
)

func (program *Program) evaluateExpression(expressionNode *ast.Expression, scope *Scope) data.Type {
	var stack []data.Type

	nodes := expressionNode.Nodes()
	if len(nodes) == 0 {
		panic(fmt.Sprintf("evaluateExpression: Expression node is missing nodes."))
	}
	if types.HasNoType(expressionNode.TypeInfo) {
		panic(fmt.Sprintf("evaluateExpression: Expression node has not been type-checked. Type Token: %v\nExpression Node Data:\n%v", expressionNode.TypeIdentifier, expressionNode))
	}

	// todo(Jake): Rewrite string concat to use `var stringBuffer bytes.Buffer` and see if
	//			   there is a speedup
	for _, itNode := range nodes {

		switch node := itNode.(type) {
		case *ast.ArrayLiteral:
			panic("evaluateExpression: todo(Jake): Support Array literal, add \"parseType\" function and return token")
		case *ast.HTMLBlock:
			value := program.evaluateHTMLBlock(node, scope)
			stack = append(stack, value)
		case *ast.Token:
			switch node.Kind {
			case token.String:
				value := &data.String{Value: node.String()}
				stack = append(stack, value)
			case token.Identifier:
				name := node.String()
				value, exists := scope.Get(name)
				if !exists {
					panic(fmt.Sprintf("Variable isn't declared '%v' on Line %d", name, node.Line))
				}
				stack = append(stack, value)
			case token.KeywordTrue:
				value := &data.Bool{Value: true}
				stack = append(stack, value)
			case token.KeywordFalse:
				value := &data.Bool{Value: false}
				stack = append(stack, value)
			default:
				if node.IsOperator() {
					rightValue := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					if len(stack) == 0 {
						panic(fmt.Sprintf("evaluateExpression(): Only got rightValue: %s, operator: %s", rightValue, node.String()))
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
			panic(fmt.Sprintf("evaluateExpression(): Unhandled type: %T", node))
		}
	}
	if len(stack) == 0 || len(stack) > 1 {
		panic("evaluateExpression(): Invalid stack. Either 0 or above 1")
	}
	result := stack[0]

	return result
}
