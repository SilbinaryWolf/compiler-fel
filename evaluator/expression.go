package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (program *Program) evaluateExpression(expressionNode *ast.Expression, scope *Scope) data.Type {
	var stack []data.Type

	nodes := expressionNode.Nodes()
	if len(nodes) == 0 {
		panic(fmt.Sprintf("evaluateExpression: Expression node is missing nodes."))
	}
	if expressionNode.TypeInfo == nil {
		panic(fmt.Sprintf("evaluateExpression: Expression node has not been type-checked. Type Token: %v\nExpression Node Data:\n%v", expressionNode.TypeIdentifier, expressionNode))
	}

	//typeInfo := expressionNode.TypeInfo
	//isStringExpr := types.Equals(typeInfo, types.String())

	// todo(Jake): Rewrite string concat to use `var stringBuffer bytes.Buffer` and see if
	//			   there is a speedup
	for _, itNode := range nodes {

		switch node := itNode.(type) {
		case *ast.StructLiteral:
			panic("Deprecated")
			/*typeinfo := node.TypeInfo.(*parser.TypeInfo_Struct)
			structDef := typeinfo.Definition()

			resultValue := new(data.Struct)
			resultValue.Fields = make([]data.Type, 0, len(structDef.Fields))

			for _, structField := range structDef.Fields {
				name := structField.Name.String()

				exprNode := &structField.Expression
				hasField := false
				for _, literalField := range node.Fields {
					if name == literalField.Name.String() {
						exprNode = &literalField.Expression
						hasField = true
						break
					}
				}
				typeinfo := exprNode.TypeInfo
				if typeinfo == nil {
					panic(fmt.Sprintf("evaluateExpression: Missing typeinfo on property for \"%s :: struct { %s }\"", structDef.Name, structField.Name))
				}
				var value data.Type
				if hasField {
					value = program.evaluateExpression(exprNode, scope)
				} else {
					panic("NOTE(Jake): Deprecated evaluator, Create()")
				}
				resultValue.Fields = append(resultValue.Fields, value)
			}
			stack = append(stack, resultValue)*/
			//panic(fmt.Sprintf("Debug struct literal data: %T, %s", typeinfo, resultValue))
		case *ast.ArrayLiteral:
			panic("NOTE(Jake): Deprecated evaluator, Create()")
			/*resultValue := data.NewArray()
			//resultValue := data.NewArray(typeInfo.Create())
			for _, itNode := range node.ChildNodes {
				switch node := itNode.(type) {
				case *ast.Expression:
					value := program.evaluateExpression(node, scope)
					resultValue.Push(value)
					continue
				}
				panic(fmt.Sprintf("evaluateExpression:arrayLiteral: Unhandled type %T", itNode))
			}
			stack = append(stack, resultValue)*/
		case *ast.HTMLBlock:
			value := program.evaluateHTMLBlock(node, scope)
			stack = append(stack, value)
		case *ast.Token:
			switch node.Kind {
			case token.String:
				value := &data.String{Value: node.String()}
				stack = append(stack, value)
			case token.Number:
				//
				// todo(Jake): Handle this string conversion at parser time.
				//				ie. ast.IntLiteral, ast.FloatLiteral
				//
				str := node.String()
				if strings.Contains(str, ".") {
					float, err := strconv.ParseFloat(node.String(), 10)
					if err != nil {
						panic(fmt.Errorf("Failed to parse float value from string: %s", err))
					}
					value := &data.Float64{Value: float}
					stack = append(stack, value)
					continue
				}
				intVal, err := strconv.ParseInt(node.String(), 10, 0)
				if err != nil {
					panic(fmt.Errorf("Failed to parse int value from string: %s", err))
				}
				value := &data.Integer64{Value: intVal}
				stack = append(stack, value)
			case token.Identifier:
				name := node.String()
				value, exists := scope.Get(name)
				if !exists {
					panic(fmt.Sprintf("Variable isn't declared '%v' on Line %d", name, node.Line))
				}
				stack = append(stack, value)
			case token.KeywordTrue:
				stack = append(stack, data.NewBool(true))
			case token.KeywordFalse:
				stack = append(stack, data.NewBool(false))
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
							stack = append(stack, &data.String{
								Value: leftTypeString.Value + rightTypeString.Value,
							})
							continue
						case token.ConditionalEqual:
							stack = append(stack, data.NewBool(leftTypeString.Value == rightTypeString.Value))
							continue
						case token.ConditionalNotEqual:
							stack = append(stack, data.NewBool(leftTypeString.Value != rightTypeString.Value))
							continue
						}
						panic(fmt.Sprintf("evaluateExpression(): Invalid operation %s with string data types.", node.Kind.String()))
					}

					// Handle bool
					rightTypeBool, rValueIsBool := rightValue.(*data.Bool)
					leftTypeBool, lValueIsBool := leftValue.(*data.Bool)
					if lValueIsBool && rValueIsBool {
						switch node.Kind {
						case token.ConditionalNotEqual:
							stack = append(stack, data.NewBool(leftTypeBool.Value() != rightTypeBool.Value()))
							continue
						case token.ConditionalEqual:
							stack = append(stack, data.NewBool(leftTypeBool.Value() == rightTypeBool.Value()))
							continue
						case token.ConditionalAnd:
							stack = append(stack, data.NewBool(leftTypeBool.Value() && rightTypeBool.Value()))
							continue
						case token.ConditionalOr:
							stack = append(stack, data.NewBool(leftTypeBool.Value() || rightTypeBool.Value()))
							continue
						}
						panic(fmt.Sprintf("evaluateExpression(): Invalid operation %s with bool data types.", node.Kind.String()))
					}

					{
						rightType, rOk := rightValue.(*data.Integer64)
						leftType, lOk := leftValue.(*data.Integer64)
						if lOk && rOk {
							switch node.Kind {
							case token.ConditionalNotEqual:
								stack = append(stack, data.NewBool(leftTypeString.Value != rightTypeString.Value))
								continue
							case token.Add:
								stack = append(stack, &data.Integer64{Value: leftType.Value + rightType.Value})
								continue
							case token.Equal:
								panic(fmt.Sprintf("Line %d - Invalid expression, got = inside it.", node.Line))
							}
						}
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
