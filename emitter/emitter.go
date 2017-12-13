package emitter

import (
	"fmt"
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/parser"
	"github.com/silbinarywolf/compiler-fel/token"
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

func debugOpcodes(opcodes []bytecode.Code) {
	fmt.Printf("Opcode Debug:\n-----------\n")
	for i, code := range opcodes {
		fmt.Printf("%d - %s\n", i, code.String())
	}
	fmt.Printf("-----------\n")
}

func (emit *Emitter) emitExpression(opcodes []bytecode.Code, node *ast.Expression) []bytecode.Code {
	nodes := node.Nodes()
	if len(nodes) == 0 {
		panic("Cannot provide an empty expression to emitExpression.")
	}

	switch typeInfo := node.TypeInfo.(type) {
	case *parser.TypeInfo_Int:
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
						tokenFloat, err := strconv.ParseFloat(node.String(), 10)
						if err != nil {
							panic(fmt.Errorf("emitExpression:DeclareStatement: Cannot convert token string to float, error: %s", err))
						}
						code := bytecode.Init(bytecode.Push)
						code.Value = tokenFloat
						opcodes = append(opcodes, code)
					} else {
						tokenInt, err := strconv.ParseInt(tokenString, 10, 0)
						if err != nil {
							panic(fmt.Sprintf("emitExpression:DeclareStatement: Cannot convert token string to int, error: %s", err))
						}
						code := bytecode.Init(bytecode.Push)
						code.Value = tokenInt
						opcodes = append(opcodes, code)
					}
				case token.Add:
					code := bytecode.Init(bytecode.Add)
					opcodes = append(opcodes, code)
				default:
					panic(fmt.Sprintf("emitExpression:DeclareStatement:Token: Unhandled token kind \"%s\"", t.Kind.String()))
				}
			default:
				panic(fmt.Sprintf("emitExpression:DeclareStatement: Unhandled type %T", node))
			}
		}
	case *parser.TypeInfo_Struct:
		typeInfo := node.TypeInfo
		structDef := node.TypeInfo.Definition()
		if structDef == nil {
			panic("emitExpression: TypeInfo_Struct: Missing Definition() data, this should be handled in the type checker.")
		}

		structData := new(bytecode.Struct)
		structData.StructDefinition = structDef
		structData.Fields = make([]interface{}, 0, len(structDef.Fields))

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
				panic(fmt.Sprintf("emitExpression: Missing typeinfo on property for \"%s :: struct { %s }\"", structDef.Name, structField.Name))
			}
		}

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
