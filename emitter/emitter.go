package emitter

import (
	"fmt"
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/parser"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
	"strconv"
	"strings"
)

type Emitter struct {
	//nameToSymbolIndex map[string]int
	symbols []bytecode.Block

	nameToStackPos map[string]int
	stackPos       int
}

func New() *Emitter {
	emit := new(Emitter)
	emit.nameToStackPos = make(map[string]int)
	//emit.nameToSymbolIndex = make(map[string]int)
	return emit
}

func appendReverse(nodes []ast.Node, nodesToPrepend []ast.Node) []ast.Node {
	for i := len(nodesToPrepend) - 1; i >= 0; i-- {
		node := nodesToPrepend[i]
		nodes = append(nodes, node)
	}
	return nodes
}

func addDebugString(opcodes []bytecode.Code, text string) []bytecode.Code {
	code := bytecode.Init(bytecode.DebugString)
	code.Value = text
	opcodes = append(opcodes, code)
	return opcodes
}

func debugOpcodes(opcodes []bytecode.Code) {
	fmt.Printf("Opcode Debug:\n-----------\n")
	for i, code := range opcodes {
		fmt.Printf("%d - %s\n", i, code.String())
	}
	fmt.Printf("-----------\n")
}

func (emit *Emitter) emitNewFromType(opcodes []bytecode.Code, typeInfo types.TypeInfo) []bytecode.Code {
	opcodes = addDebugString(opcodes, "emitNewFromType")
	switch typeInfo.(type) {
	case *parser.TypeInfo_Int:
		code := bytecode.Init(bytecode.Push)
		code.Value = int(0)
		opcodes = append(opcodes, code)
	case *parser.TypeInfo_Float:
		code := bytecode.Init(bytecode.Push)
		code.Value = float64(0.0)
		opcodes = append(opcodes, code)
	case *parser.TypeInfo_String:
		code := bytecode.Init(bytecode.Push)
		code.Value = ""
		opcodes = append(opcodes, code)
	default:
		panic(fmt.Sprintf("emitNewFromType: Unhandled type %T", typeInfo))
	}
	return opcodes
}

func (emit *Emitter) emitExpression(opcodes []bytecode.Code, node *ast.Expression) []bytecode.Code {
	nodes := node.Nodes()
	if len(nodes) == 0 {
		panic("Cannot provide an empty expression to emitExpression.")
	}

	switch typeInfo := node.TypeInfo.(type) {
	case *parser.TypeInfo_Int:
		//*parser.TypeInfo_Float:
		for _, node := range nodes {
			switch node := node.(type) {
			case *ast.Token:
				t := node.Token
				switch t.Kind {
				case token.Identifier:
					nameString := t.String()
					stackPos, ok := emit.nameToStackPos[nameString]
					if !ok {
						panic("Undeclared variable \"%s\", this should be caught in the type checker.")
					}
					code := bytecode.Init(bytecode.PushStackVar)
					code.Value = stackPos
					opcodes = append(opcodes, code)
				case token.ConditionalEqual:
					code := bytecode.Init(bytecode.ConditionalEqual)
					opcodes = append(opcodes, code)
				case token.Number:
					tokenString := t.String()
					if strings.Contains(tokenString, ".") {
						//if typeInfo.(type) == *parser.TypeInfo_Int {
						//	panic("This should not happen as the type is int.")
						//}
						tokenFloat, err := strconv.ParseFloat(node.String(), 10)
						if err != nil {
							panic(fmt.Errorf("emitExpression: Cannot convert token string to float, error: %s", err))
						}
						code := bytecode.Init(bytecode.Push)
						code.Value = tokenFloat
						opcodes = append(opcodes, code)
					} else {
						tokenInt, err := strconv.ParseInt(tokenString, 10, 0)
						if err != nil {
							panic(fmt.Sprintf("emitExpression: Cannot convert token string to int, error: %s", err))
						}
						code := bytecode.Init(bytecode.Push)
						code.Value = tokenInt
						opcodes = append(opcodes, code)
					}
				case token.Add:
					code := bytecode.Init(bytecode.Add)
					opcodes = append(opcodes, code)
				default:
					panic(fmt.Sprintf("emitExpression:Token: Unhandled token kind \"%s\"", t.Kind.String()))
				}
			default:
				panic(fmt.Sprintf("emitExpression: Unhandled type %T", node))
			}
		}
	case *parser.TypeInfo_String:
		/*for _, node := range node.Nodes() {
			switch node := node.(type) {
			case *ast.Token:
			}
			code := bytecode.Init(bytecode.Push)
			code.Value =
			opcodes = append(opcodes, code)
		}*/
		panic("todo(Jake): add support for string expressions")
	case *parser.TypeInfo_Struct:
		structDef := typeInfo.Definition()
		if structDef == nil {
			panic("emitExpression: TypeInfo_Struct: Missing Definition() data, this should be handled in the type checker.")
		}

		var structLiteral *ast.StructLiteral
		if len(nodes) > 0 {
			var ok bool
			structLiteral, ok = nodes[0].(*ast.StructLiteral)
			if !ok {
				panic("emitExpression: Should only have ast.StructLiteral in TypeInfo_Struct expression, this should be handled in type checker.")

			}
			if len(nodes) > 1 {
				panic("emitExpression: Should only have one node in TypeInfo_Struct, this should be handled in type checker.")
			}
		}

		// NOTE(Jake): This belongs in "vm"
		//structData := new(bytecode.Struct)
		//structData.StructDefinition = structDef
		//structData.Fields = make([]interface{}, 0, len(structDef.Fields))

		code := bytecode.Init(bytecode.AllocStruct)
		code.Value = len(structDef.Fields)
		opcodes = append(opcodes, code)

		for offset, structField := range structDef.Fields {
			name := structField.Name.String()

			exprNode := &structField.Expression
			hasField := false
			for _, literalField := range structLiteral.Fields {
				if name == literalField.Name.String() {
					exprNode = &literalField.Expression
					hasField = true
					break
				}
			}
			fieldTypeInfo := exprNode.TypeInfo
			if fieldTypeInfo == nil {
				panic(fmt.Sprintf("emitExpression: Missing typeinfo on property for \"%s :: struct { %s }\"", structDef.Name, structField.Name))
			}
			if hasField {
				opcodes = emit.emitExpression(opcodes, exprNode)
			} else {
				opcodes = emit.emitNewFromType(opcodes, fieldTypeInfo)
			}

			code := bytecode.Init(bytecode.StoreStructField)
			code.Value = offset
			opcodes = append(opcodes, code)
			//structData.Fields = append(structData.Fields, value)
		}

		debugOpcodes(opcodes)
		panic("todo(Jake): TypeInfo_Struct")
	default:
		panic(fmt.Sprintf("emitExpression: Unhandled expression with type %T", typeInfo))
	}
	return opcodes
}

func (emit *Emitter) emitStatement(opcodes []bytecode.Code, node ast.Node) []bytecode.Code {
	switch node := node.(type) {
	case *ast.DeclareStatement:
		opcodes = emit.emitExpression(opcodes, &node.Expression)

		code := bytecode.Init(bytecode.Store)
		code.Value = emit.stackPos
		nameString := node.Name.String()
		_, ok := emit.nameToStackPos[nameString]
		if ok {
			panic("Redeclared \"%s\" in same scope, this should be caught in the type checker.")
		}
		emit.nameToStackPos[nameString] = emit.stackPos
		emit.stackPos++

		opcodes = append(opcodes, code)
	case *ast.OpStatement:
		stackPos, ok := emit.nameToStackPos[node.Name.String()]
		if !ok {
			panic(fmt.Sprintf("Missing declaration for %s, this should be caught in the type checker.", node.Name))
		}
		opcodes = emit.emitExpression(opcodes, &node.Expression)
		code := bytecode.Init(bytecode.Store)
		code.Value = stackPos
		opcodes = append(opcodes, code)
	case *ast.If:
		originalOpcodesLength := len(opcodes)

		opcodes = emit.emitExpression(opcodes, &node.Condition)
		jumpCodeOffset := len(opcodes)
		opcodes = append(opcodes, bytecode.Init(bytecode.JumpIfFalse))

		// Generate bytecode
		beforeIfStatementCount := len(opcodes)
		nodes := node.Nodes()
		for _, node := range nodes {
			opcodes = emit.emitStatement(opcodes, node)
		}

		if beforeIfStatementCount == len(opcodes) {
			// Dont output any bytecode for an empty if
			opcodes = opcodes[:originalOpcodesLength] // Remove if statement
			break
		}
		opcodes[jumpCodeOffset].Value = len(opcodes)
		debugOpcodes(opcodes)
	case *ast.CSSRule:
		panic("todo(Jake): CSSRule")
	case *ast.CSSProperty:
		panic("todo(Jake): CSSProperty")
	case *ast.HTMLComponentDefinition:
		panic(fmt.Sprintf("emitStatement: Todo HTMLComponentDef"))
	case *ast.CSSDefinition:
		emit.EmitBytecode(node)
		panic(fmt.Sprintf("emitStatement: Todo CSSDefinition"))
	case *ast.StructDefinition,
		*ast.CSSConfigDefinition:
		break
	default:
		panic(fmt.Sprintf("emitStatement: Unhandled type %T", node))
	}
	return opcodes
}

func (emit *Emitter) EmitBytecode(node ast.Node) *bytecode.Block {
	opcodes := make([]bytecode.Code, 0, 10)

	topNodes := make([]ast.Node, 0, 10)
	topNodes = appendReverse(topNodes, node.Nodes())
	if topNodes == nil {
		panic("EmitBytecode: Top-level node shouldnt have no nodes.")
	}
	for len(topNodes) > 0 {
		node := topNodes[len(topNodes)-1]
		topNodes = topNodes[:len(topNodes)-1]

		opcodes = emit.emitStatement(opcodes, node)
	}
	codeBlock := new(bytecode.Block)
	codeBlock.Opcodes = opcodes
	codeBlock.StackSize = emit.stackPos
	debugOpcodes(opcodes)
	fmt.Printf("Stack Size: %d\n", codeBlock.StackSize)
	return codeBlock
}
