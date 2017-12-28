package emitter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	//"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/parser"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
)

type VariableInfo struct {
	stackPos  int
	structDef *ast.StructDefinition // TypeInfo_Struct only
}

type Scope struct {
	mapToInfo map[string]VariableInfo
	parent    *Scope
}

type Emitter struct {
	//symbols []bytecode.Block
	scope    *Scope
	stackPos int
}

func (emit *Emitter) PushScope() {
	scope := new(Scope)
	scope.mapToInfo = make(map[string]VariableInfo)
	scope.parent = emit.scope

	emit.scope = scope
}

func (emit *Emitter) PopScope() {
	parentScope := emit.scope.parent
	if parentScope == nil {
		panic("Cannot pop last scope item.")
	}
	emit.scope = parentScope
}

func (scope *Scope) DeclareSet(name string, varInfo VariableInfo) {
	_, ok := scope.mapToInfo[name]
	if ok {
		panic(fmt.Sprintf("Cannot redeclare variable \"%s\" in same scope. This should be caught in type checker.", name))
	}
	scope.mapToInfo[name] = varInfo
}

func (scope *Scope) GetThisScope(name string) (VariableInfo, bool) {
	result, ok := scope.mapToInfo[name]
	return result, ok
}

func (scope *Scope) Get(name string) (VariableInfo, bool) {
	result, ok := scope.mapToInfo[name]
	if !ok {
		if scope.parent == nil {
			return VariableInfo{}, false
		}
		result, ok = scope.parent.Get(name)
	}
	return result, ok
}

func New() *Emitter {
	emit := new(Emitter)
	emit.PushScope()
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
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.Label,
		Value: "DEBUG LABEL: " + text,
	})
	return opcodes
}

func debugOpcodes(opcodes []bytecode.Code) {
	fmt.Printf("Opcode Debug:\n-----------\n")
	for i, code := range opcodes {
		fmt.Printf("%d - %s\n", i, code.String())
	}
	fmt.Printf("-----------\n")
}

//
// This is used to emit bytecode for zeroing out a type that wasn't given an
// explicit value
//
// ie. "test: string"
//
func (emit *Emitter) emitNewFromType(opcodes []bytecode.Code, typeInfo types.TypeInfo) []bytecode.Code {
	//opcodes = addDebugString(opcodes, "emitNewFromType")
	switch typeInfo := typeInfo.(type) {
	case *parser.TypeInfo_Int:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: int(0),
		})
	case *parser.TypeInfo_Float:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: float64(0),
		})
	case *parser.TypeInfo_String:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: "",
		})
	case *parser.TypeInfo_Array:
		underlyingType := typeInfo.Underlying()
		var code bytecode.Code
		switch underlyingType.(type) {
		case *parser.TypeInfo_String:
			code.Kind = bytecode.PushArrayString
		default:
			panic(fmt.Sprintf("emitNewFromType:Array: Unhandled type %T", underlyingType))
		}
		if code.Kind == bytecode.Unknown {
			panic("Unhandled.")
		}
		opcodes = append(opcodes, code)
	case *parser.TypeInfo_Struct:
		structDef := typeInfo.Definition()
		if structDef == nil {
			panic("emitExpression: TypeInfo_Struct: Missing Definition() data, this should be handled in the type checker.")
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.PushAllocStruct,
			Value: len(structDef.Fields),
		})
		for offset, structField := range structDef.Fields {
			exprNode := &structField.Expression
			fieldTypeInfo := exprNode.TypeInfo
			if fieldTypeInfo == nil {
				panic(fmt.Sprintf("emitExpression: Missing type info on property for \"%s :: struct { %s }\"", structDef.Name, structField.Name))
			}
			if len(exprNode.Nodes()) == 0 {
				opcodes = emit.emitNewFromType(opcodes, fieldTypeInfo)
			} else {
				opcodes = emit.emitExpression(opcodes, exprNode)
			}

			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.StorePopStructField,
				Value: offset,
			})
		}
	default:
		panic(fmt.Sprintf("emitNewFromType: Unhandled type %T", typeInfo))
	}
	return opcodes
}

func (emit *Emitter) getPushVariableCodeFromIdent(ident string) bytecode.Code {
	varInfo, ok := emit.scope.Get(ident)
	if !ok {
		panic("Undeclared variable \"%s\", this should be caught in the type checker.")
	}
	return bytecode.Code{
		Kind:  bytecode.PushStackVar,
		Value: varInfo.stackPos,
	}
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
				switch t := node.Token; t.Kind {
				case token.Identifier:
					opcodes = append(opcodes, emit.getPushVariableCodeFromIdent(t.String()))
				case token.ConditionalEqual:
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.ConditionalEqual,
					})
				case token.Number:
					tokenString := t.String()
					if strings.Contains(tokenString, ".") {
						panic("Cannot add float to int, this should be caught in the type checker.")
						//if typeInfo.(type) == *parser.TypeInfo_Int {
						//	panic("This should not happen as the type is int.")
						//}
						/*tokenFloat, err := strconv.ParseFloat(node.String(), 10)
						if err != nil {
							panic(fmt.Errorf("emitExpression:TypeInfo_Int:Token: Cannot convert token string to float, error: %s", err))
						}
						code := bytecode.Init(bytecode.Push)
						code.Value = tokenFloat
						opcodes = append(opcodes, code)*/
					} else {
						tokenInt, err := strconv.ParseInt(tokenString, 10, 0)
						if err != nil {
							panic(fmt.Sprintf("emitExpression:Int:Token: Cannot convert token string to int, error: %s", err))
						}
						opcodes = append(opcodes, bytecode.Code{
							Kind:  bytecode.Push,
							Value: tokenInt,
						})
					}
				case token.Add:
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.Add,
					})
				default:
					panic(fmt.Sprintf("emitExpression:Int:Token: Unhandled token kind \"%s\"", t.Kind.String()))
				}
			default:
				panic(fmt.Sprintf("emitExpression:Int: Unhandled type %T", node))
			}
		}
	case *parser.TypeInfo_String:
		for _, node := range node.Nodes() {
			switch node := node.(type) {
			case *ast.Token:
				switch t := node.Token; t.Kind {
				case token.Identifier:
					opcodes = append(opcodes, emit.getPushVariableCodeFromIdent(t.String()))
				case token.Add:
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.AddString,
					})
				case token.String:
					opcodes = append(opcodes, bytecode.Code{
						Kind:  bytecode.Push,
						Value: t.String(),
					})
				default:
					panic(fmt.Sprintf("emitExpression:String:Token: Unhandled token kind \"%s\"", t.Kind.String()))
				}
			default:
				panic(fmt.Sprintf("emitExpression:String: Unhandled type %T", node))
			}
		}
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

		if len(structLiteral.Fields) == 0 {
			// NOTE(Jake) 2017-12-28
			// If using struct literal syntax "MyStruct{}" without fields, assume all fields
			// use default values.
			opcodes = emit.emitNewFromType(opcodes, typeInfo)
		} else {
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.PushAllocStruct,
				Value: len(structDef.Fields),
			})

			// NOTE(Jake): 2017-12-28
			// If using struct literal syntax "MyStruct{FieldA: 3, Name: "Jake"}" with fields, you need
			// to provide *ALL* the field values. Reasoning is that one method implies you want explicit
			// structures (maybe for testing data) so if you add additional fields to a struct later, the compiler
			// will tell you about missing fields.
			for offset, structField := range structDef.Fields {
				name := structField.Name.String()
				exprNode := &structField.Expression
				for _, literalField := range structLiteral.Fields {
					if name == literalField.Name.String() {
						exprNode = &literalField.Expression
						break
					}
				}
				if fieldTypeInfo := exprNode.TypeInfo; fieldTypeInfo == nil {
					panic(fmt.Sprintf("emitExpression: Missing type info on property for \"%s :: struct { %s }\"", structDef.Name, structField.Name))
				}
				if len(exprNode.Nodes()) == 0 {
					panic(fmt.Sprintf("emitExpression:TypeInfo_Struct: Missing info for field \"%s\" on \"%s :: struct\", typechecker should enforce that you need all fields.", structField.Name, structDef.Name))
				}
				opcodes = emit.emitExpression(opcodes, exprNode)
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.StorePopStructField,
					Value: offset,
				})
			}
		}

		//debugOpcodes(opcodes)
		//panic("todo(Jake): TypeInfo_Struct")
	default:
		panic(fmt.Sprintf("emitExpression: Unhandled expression with type %T", typeInfo))
	}
	return opcodes
}

func (emit *Emitter) emitStatement(opcodes []bytecode.Code, node ast.Node) []bytecode.Code {
	switch node := node.(type) {
	case *ast.Block:
		emit.PushScope()
		for _, node := range node.Nodes() {
			opcodes = emit.emitStatement(opcodes, node)
		}
		emit.PopScope()
	case *ast.CSSDefinition:
		nodes := node.Nodes()
		if len(nodes) > 0 {
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.Label,
				Value: "css:" + node.Name.String(),
			})

			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.PushNewContextNode,
				Value: bytecode.NodeCSSDefinition,
			})

			for _, node := range nodes {
				opcodes = emit.emitStatement(opcodes, node)
			}
			//debugOpcodes(opcodes)
		}
	case *ast.CSSRule:
		/*var emptyTypeData *data.CSSDefinition
		internalType := reflect.TypeOf(emptyTypeData)
		code := bytecode.Init(bytecode.PushAllocInternalStruct)
		code.Value = internalType
		opcodes = append(opcodes, code)
		{
			fieldMeta, ok := internalType.FieldByName("Name")
			if !ok {
				panic("Cannot find \"Name\".")
			}
			code := bytecode.Init(bytecode.StoreInternalStructField)
			code.Value = fieldMeta
			opcodes = append(opcodes, code)
		}*/
		// no-op
		debugOpcodes(opcodes)
		panic("todo(Jake): *ast.CSSRule")
	case *ast.CSSProperty:
		//panic("todo(Jake): *ast.CSSProperty")
	case *ast.DeclareStatement:
		opcodes = emit.emitExpression(opcodes, &node.Expression)
		typeInfo := node.Expression.TypeInfo

		nameString := node.Name.String()
		_, ok := emit.scope.GetThisScope(nameString)
		if ok {
			panic(fmt.Sprintf("Redeclared \"%s\" in same scope, this should be caught in the type checker.", nameString))
		}

		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Store,
			Value: emit.stackPos,
		})
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.Pop,
		})

		{
			var varStructDef *ast.StructDefinition = nil
			if typeInfo, ok := typeInfo.(*parser.TypeInfo_Struct); ok {
				varStructDef = typeInfo.Definition()
			}
			emit.scope.DeclareSet(nameString, VariableInfo{
				stackPos:  emit.stackPos,
				structDef: varStructDef,
			})
			emit.stackPos++
		}
	case *ast.OpStatement:
		name := node.LeftHandSide[0].String()
		varInfo, ok := emit.scope.Get(name)
		if !ok {
			panic(fmt.Sprintf("Missing declaration for %s, this should be caught in the type checker.", name))
		}
		if len(node.LeftHandSide) > 1 {
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.PushStackVar,
				Value: varInfo.stackPos,
			})
		}
		switch node.Operator.Kind {
		case token.Equal:
			// no-op
		case token.AddEqual:
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.PushStackVar,
				Value: varInfo.stackPos,
			})
		default:
			panic(fmt.Sprintf("emitStatement: Unhandled operator kind: %s", node.Operator.Kind.String()))
		}

		opcodes = emit.emitExpression(opcodes, &node.Expression)
		switch node.Operator.Kind {
		case token.Equal:
			// no-op
		case token.AddEqual:
			switch node.TypeInfo.(type) {
			case *parser.TypeInfo_String:
				opcodes = append(opcodes, bytecode.Code{
					Kind: bytecode.AddString,
				})
			default:
				opcodes = append(opcodes, bytecode.Code{
					Kind: bytecode.Add,
				})
			}
		default:
			panic(fmt.Sprintf("emitStatement: Unhandled operator kind: %s", node.Operator.Kind.String()))
		}
		if len(node.LeftHandSide) > 1 {
			structDef := varInfo.structDef
			if structDef == nil {
				panic("emitStatement: Non-struct cannot have properties. This should be caught in the typechecker.")
			}
			fieldName := node.LeftHandSide[1].String()
			field := structDef.GetFieldByName(fieldName)
			if field == nil {
				panic(fmt.Sprintf("emitStatement: \"%s :: struct\" does not have property \"%s\". This should be caught in the typechecker.", structDef.Name, fieldName))
			}
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.StorePopStructField,
				Value: field.Index,
			})
			if len(node.LeftHandSide) > 2 {
				panic("todo(Jake): Support nested properties")
			}
		} else {
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.Store,
				Value: varInfo.stackPos,
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.Pop,
			})
		}
	case *ast.If:
		originalOpcodesLength := len(opcodes)

		opcodes = emit.emitExpression(opcodes, &node.Condition)

		var jumpCodeOffset int
		{
			jumpCodeOffset = len(opcodes)
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.JumpIfFalse,
			})
		}

		// Generate bytecode
		beforeIfStatementCount := len(opcodes)
		nodes := node.Nodes()
		for _, node := range nodes {
			opcodes = emit.emitStatement(opcodes, node)
		}

		if beforeIfStatementCount == len(opcodes) {
			// Dont output any bytecode for an empty `if`
			opcodes = opcodes[:originalOpcodesLength] // Remove if statement
			break
		}
		opcodes[jumpCodeOffset].Value = len(opcodes)
	case *ast.HTMLComponentDefinition:
		//panic(fmt.Sprintf("emitStatement: Todo HTMLComponentDef"))
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
	fmt.Printf("Final bytecode output above\nStack Size: %d\n", codeBlock.StackSize)
	return codeBlock
}
