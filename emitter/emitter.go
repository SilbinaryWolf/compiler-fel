package emitter

import (
	"fmt"
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/token"
	"strconv"
	"strings"
)

type Emitter struct {
	nameToSymbolIndex map[string]int
	symbols           []CodeBlock

	stackPos int
}

// ie. a function, block-scope, HTMLComponent
type CodeBlock struct {
	opCodes   []bytecode.Code
	stackSize int
}

func New() *Emitter {
	emit := new(Emitter)
	emit.nameToSymbolIndex = make(map[string]int)
	return emit
}

func appendReverse(nodes []ast.Node, nodesToPrepend []ast.Node) []ast.Node {
	for i := len(nodesToPrepend) - 1; i >= 0; i-- {
		node := nodesToPrepend[i]
		nodes = append(nodes, node)
	}
	return nodes
}

func (emit *Emitter) emitExpression(node *ast.Expression) {

}

func (emit *Emitter) EmitBytecode(node ast.Node) *CodeBlock {
	emit.stackPos = 0

	opcodes := make([]bytecode.Code, 0, 10)

	topNodes := make([]ast.Node, 0, 10)
	topNodes = appendReverse(topNodes, node.Nodes())
	if topNodes == nil {
		panic("EmitBytecode: Top-level node shouldnt have no nodes.")
	}
	for len(topNodes) > 0 {
		node := topNodes[len(topNodes)-1]
		topNodes = topNodes[:len(topNodes)-1]

		switch node := node.(type) {
		case *ast.DeclareStatement:
			for _, node := range node.Nodes() {
				switch node := node.(type) {
				case *ast.Token:
					t := node.Token
					switch t.Kind {
					case token.Number:
						tokenString := t.String()

						if strings.Contains(tokenString, ".") {
							tokenFloat, err := strconv.ParseFloat(node.String(), 10)
							if err != nil {
								panic(fmt.Errorf("EmitBytecode:DeclareStatement: Cannot convert token string to float, error: %s", err))
							}
							code := bytecode.Init(bytecode.Push)
							code.Value = tokenFloat
							opcodes = append(opcodes, code)
						} else {
							tokenInt, err := strconv.ParseInt(tokenString, 10, 0)
							if err != nil {
								panic(fmt.Sprintf("EmitBytecode:DeclareStatement: Cannot convert token string to int, error: %s", err))
							}
							code := bytecode.Init(bytecode.Push)
							code.Value = tokenInt
							opcodes = append(opcodes, code)
						}
					case token.Add:
						code := bytecode.Init(bytecode.Add)
						opcodes = append(opcodes, code)
					default:
						panic(fmt.Sprintf("EmitBytecode:DeclareStatement:Token: Unhandled token kind \"%s\"", t.Kind.String()))
					}
				default:
					panic(fmt.Sprintf("EmitBytecode:DeclareStatement: Unhandled type %T", node))
				}
			}
			code := bytecode.Init(bytecode.Store)
			code.StackPos = emit.stackPos
			emit.stackPos++
			opcodes = append(opcodes, code)

			// Debug opcodes
			fmt.Printf("Opcode Debug:\n-----------\n")
			for _, code := range opcodes {
				fmt.Printf("%s\n", code.String())
			}
			fmt.Printf("-----------\n")

			panic("todo(Jake): Finish DeclareStatement opcode")
		case *ast.CSSRule:
			topNodes = appendReverse(topNodes, node.Nodes())
		case *ast.CSSProperty:
			panic("todo(Jake): CSSProperty")
		case *ast.HTMLComponentDefinition:
			panic(fmt.Sprintf("EmitBytecode: Todo HTMLComponentDef"))
		case *ast.CSSDefinition:
			emit.EmitBytecode(node)
			panic(fmt.Sprintf("EmitBytecode: Todo CSSDefinition"))
		case *ast.StructDefinition,
			*ast.CSSConfigDefinition:
			continue
		default:
			panic(fmt.Sprintf("EmitBytecode: Unhandled type %T", node))
		}
	}
	codeBlock := new(CodeBlock)
	codeBlock.stackSize = emit.stackPos
	panic("todo(Jake):Finish emitBytecode")
	return codeBlock
}
