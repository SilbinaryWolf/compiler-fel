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

	typeInfo := expressionNode.TypeInfo
	//isStringExpr := types.Equals(typeInfo, types.String())

	// todo(Jake): Rewrite string concat to use `var stringBuffer bytes.Buffer` and see if
	//			   there is a speedup
	for _, itNode := range nodes {

		switch node := itNode.(type) {
		case *ast.ArrayLiteral:
			resultValue := data.NewArray(typeInfo.Create())
			for _, itNode := range node.ChildNodes {
				switch node := itNode.(type) {
				case *ast.Expression:
					value := program.evaluateExpression(node, scope)
					resultValue.Push(value)
					continue
				}
				panic(fmt.Sprintf("evaluateExpression:arrayLiteral: Unhandled type %T", itNode))
			}
			stack = append(stack, resultValue)
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

					// Handle string
					rightTypeString, rValueIsString := rightValue.(*data.String)
					leftTypeString, lValueIsString := leftValue.(*data.String)
					if lValueIsString && rValueIsString {
						/*rightTypeString, ok := rightValue.(*data.String)
							if !ok {
								panic(fmt.Sprintf("evaluateExpression(): Unexpected error, expression type was string but right-value inside weren't. Left: %s, Right: %s", leftValue.String(), rightValue.String()))
							}
							leftTypeString, ok := leftValue.(*data.String)
							if !ok {
								panic(fmt.Sprintf("evaluateExpression(): Unexpected error, expression type was string but left-value inside weren't."))
						}*/
						switch node.Kind {
						case token.Add:
							result := &data.String{
								Value: leftTypeString.Value + rightTypeString.Value,
							}
							stack = append(stack, result)
							continue
						case token.ConditionalEqual:
							result := &data.Bool{
								Value: leftTypeString.Value == rightTypeString.Value,
							}
							stack = append(stack, result)
							continue
						}
						panic(fmt.Sprintf("evaluateExpression(): Invalid operation %s with string data types.", node.Kind.String()))
					}

					//
					rightTypeBool, rValueIsBool := rightValue.(*data.Bool)
					leftTypeBool, lValueIsBool := leftValue.(*data.Bool)
					if lValueIsBool && rValueIsBool {
						switch node.Kind {
						case token.ConditionalEqual:
							stack = append(stack, &data.Bool{
								Value: leftTypeBool.Value == rightTypeBool.Value,
							})
							continue
						case token.ConditionalAnd:
							stack = append(stack, &data.Bool{
								Value: leftTypeBool.Value && rightTypeBool.Value,
							})
							continue
						case token.ConditionalOr:
							stack = append(stack, &data.Bool{
								Value: leftTypeBool.Value || rightTypeBool.Value,
							})
							continue
						}
						panic(fmt.Sprintf("evaluateExpression(): Invalid operation %s with bool data types.", node.Kind.String()))
					}

					panic(fmt.Sprintf("todo(Jake): Handle %s numbers together. (Only strings supported currently)", node.Kind.String()))
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
