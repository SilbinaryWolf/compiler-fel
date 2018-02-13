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

type VariableKind int

const (
	VariableDefault = iota + 0
	VariableStruct
	//VariableStructField
)

type VariableInfo struct {
	kind     VariableKind
	stackPos int

	// VariableStruct
	structDef *ast.StructDefinition

	// VariableStructField
	//structFieldPos int
}

type Scope struct {
	mapToInfo map[string]VariableInfo
	parent    *Scope
}

type Emitter struct {
	symbols          map[string]*bytecode.Block
	scope            *Scope
	stackPos         int
	htmlElementStack []string
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
	emit.symbols = make(map[string]*bytecode.Block)
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

func (emit *Emitter) registerSymbol(name string, block *bytecode.Block) bool {
	_, ok := emit.symbols[name]
	if ok {
		return false
	}
	emit.symbols[name] = block
	return true
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
	case *parser.TypeInfo_Bool:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: false,
		})
	case *parser.TypeInfo_Array:
		underlyingType := typeInfo.Underlying()
		switch underlyingType.(type) {
		case *parser.TypeInfo_String:
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.PushAllocArrayString,
				Value: 0,
			})
		default:
			panic(fmt.Sprintf("emitNewFromType:Array: Unhandled type %T", underlyingType))
		}
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

func (emit *Emitter) emitVariableIdent(opcodes []bytecode.Code, ident token.Token) []bytecode.Code {
	name := ident.String()
	varInfo, ok := emit.scope.Get(name)
	if !ok {
		panic(fmt.Sprintf("Missing declaration for %s, this should be caught in the type checker.", name))
	}
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.PushStackVar,
		Value: varInfo.stackPos,
	})
	return opcodes
}

func (emit *Emitter) emitVariableIdentWithProperty(
	opcodes []bytecode.Code,
	leftHandSide []token.Token,
) ([]bytecode.Code, int) {
	// NOTE(Jake): 2017-12-29
	//
	// `varInfo` is required so we aren't using `emitVariableIdent`
	//
	name := leftHandSide[0].String()
	varInfo, ok := emit.scope.Get(name)
	if !ok {
		panic(fmt.Sprintf("Missing declaration for %s, this should be caught in the type checker.", name))
	}
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.PushStackVar,
		Value: varInfo.stackPos,
	})
	var lastPropertyField *ast.StructField
	if len(leftHandSide) <= 1 {
		return opcodes, varInfo.stackPos
	}
	structDef := varInfo.structDef
	if structDef == nil {
		panic(fmt.Sprintf("emitStatement: Expected parameter %s to be a struct, this should be set when declaring a new variable (if applicable)", name))
	}
	for i := 1; i < len(leftHandSide)-1; i++ {
		if structDef == nil {
			panic("emitStatement: Non-struct cannot have properties. This should be caught in the typechecker.")
		}
		fieldName := leftHandSide[i].String()
		field := structDef.GetFieldByName(fieldName)
		if field == nil {
			panic(fmt.Sprintf("emitStatement: \"%s :: struct\" does not have property \"%s\". This should be caught in the typechecker.", structDef.Name, fieldName))
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.ReplaceStructFieldVar,
			Value: field.Index,
		})
		if typeInfo, ok := field.TypeInfo.(*parser.TypeInfo_Struct); ok {
			structDef = typeInfo.Definition()
		}
	}
	fieldName := leftHandSide[len(leftHandSide)-1].String()
	lastPropertyField = structDef.GetFieldByName(fieldName)
	if lastPropertyField == nil {
		panic(fmt.Sprintf("emitStatement: \"%s :: struct\" does not have property \"%s\". This should be caught in the typechecker.", structDef.Name, fieldName))
	}
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.ReplaceStructFieldVar,
		Value: lastPropertyField.Index,
	})
	return opcodes, lastPropertyField.Index
}

func (emit *Emitter) emitProcedureCall(opcodes []bytecode.Code, node *ast.Call) []bytecode.Code {
	name := node.Name.String()
	block, ok := emit.symbols[name]
	if !ok {
		panic(fmt.Sprintf("Missing procedure %s, this should be caught in the typechecker", name))
	}
	for i := 0; i < len(node.Parameters); i++ {
		expr := node.Parameters[i]
		opcodes = emit.emitExpression(opcodes, &expr.Expression)
		//opcodes = emit.emitNewFromType(opcodes, parameter.TypeInfo)
	}
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.Call,
		Value: block,
	})
	return opcodes
}

func (emit *Emitter) emitHTMLNode(opcodes []bytecode.Code, node *ast.Call) []bytecode.Code {
	emit.PushScope()
	defer emit.PopScope()

	definition := node.HTMLDefinition
	if definition != nil {
		name := node.Name.String()
		block, ok := emit.symbols[name]
		if !ok {
			panic(fmt.Sprintf("Missing HTML component %s, this should be caught in the typechecker", name))
		}

		// If definition has used the "children" keyword
		// todo(Jake): 2018-02-09
		//
		// Store whether definition can have children or not.
		// If it cant, then dont push "children" value as first parameter.
		//
		hasChildren := true
		if hasChildren {
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.PushAllocHTMLFragment,
			})
			for _, node := range node.Nodes() {
				opcodes = emit.emitStatement(opcodes, node)
				// NOTE(Jake): 2018-02-08
				//
				// We remove the final bytecode here so that we can push the resulting
				//
				//
				lastOpcode := &opcodes[len(opcodes)-1]
				switch kind := lastOpcode.Kind; kind {
				case bytecode.AppendPopHTMLElementToHTMLElement,
					bytecode.AppendPopHTMLNodeReturn:
					opcodes = opcodes[:len(opcodes)-1]
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.AppendPopHTMLElementToHTMLElement,
					})
				default:
					panic(fmt.Sprintf("emitHTMLNode:Component: Unhandled kind %v", kind))
				}
			}
		}

		if structDef := definition.Struct; structDef != nil {
			for i := 0; i < len(structDef.Fields); i++ {
				structField := structDef.Fields[i]
				name := structField.Name.String()
				exprNode := &structField.Expression
				for _, parameterField := range node.Parameters {
					if name == parameterField.Name.String() {
						exprNode = &parameterField.Expression
						break
					}
				}
				if len(exprNode.Nodes()) == 0 {
					opcodes = emit.emitNewFromType(opcodes, exprNode.TypeInfo)
				} else {
					opcodes = emit.emitExpression(opcodes, exprNode)
				}
			}
		}

		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.CallHTML,
			Value: block,
		})
		return opcodes
	}

	tagName := node.Name.String()
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.PushAllocHTMLNode,
		Value: tagName,
	})
	// todo(Jake): 2018-01-15
	//
	// Add bytecode called StoreHTMLNodeAttribute which
	// acts like a store, but instead takes parameter.Name
	//
	// This will store data on the HTMLElement structs map[string]interface{}
	//
	for i := 0; i < len(node.Parameters); i++ {
		parameter := node.Parameters[i]
		exprNode := &parameter.Expression
		if len(exprNode.Nodes()) == 0 {
			opcodes = emit.emitNewFromType(opcodes, exprNode.TypeInfo)
		} else {
			opcodes = emit.emitExpression(opcodes, exprNode)
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.StorePopHTMLAttribute,
			Value: parameter.Name.String(),
		})
	}

	{
		// NOTE(Jake): 2017-01-17
		//
		// This htmlElementStack can be replaced with a simple `int`
		// counter. The reason its using an array for now is to make
		// debugging potentially easier.
		//
		emit.htmlElementStack = append(emit.htmlElementStack, tagName)
		for _, node := range node.Nodes() {
			opcodes = emit.emitStatement(opcodes, node)
		}
		emit.htmlElementStack = emit.htmlElementStack[:len(emit.htmlElementStack)-1]
	}

	if isRootHTMLElement := len(emit.htmlElementStack) == 0; isRootHTMLElement {
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.AppendPopHTMLNodeReturn,
		})
	} else {
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.AppendPopHTMLElementToHTMLElement,
		})
	}
	return opcodes
}

func (emit *Emitter) emitExpression(opcodes []bytecode.Code, topNode *ast.Expression) []bytecode.Code {
	nodes := topNode.Nodes()
	if len(nodes) == 0 {
		panic("emitExpression: Cannot provide an empty expression to emitExpression.")
	}
	typeInfo := topNode.TypeInfo

	for _, node := range nodes {
		switch node := node.(type) {
		case *ast.TokenList:
			opcodes, _ = emit.emitVariableIdentWithProperty(opcodes, node.Tokens())
		case *ast.Call:
			switch node.Kind() {
			case ast.CallProcedure:
				opcodes = emit.emitProcedureCall(opcodes, node)
			case ast.CallHTMLNode:
				if len(node.Nodes()) > 0 {
					panic("emitExpression: Cannot have sub-elements for HTMLNode in an expression.")
				}
				panic("emitExpression: Cannot use HTMLNode in expression currently.")
				//opcodes = emit.emitHTMLNode(opcodes, node)
			default:
				panic(fmt.Errorf("emitExpression: Unhandled *ast.Call kind: %s", node.Name))
			}
		case *ast.Token:
			switch t := node.Token; t.Kind {
			case token.Identifier:
				opcodes = emit.emitVariableIdent(opcodes, t)
			case token.ConditionalEqual:
				opcodes = append(opcodes, bytecode.Code{
					Kind: bytecode.ConditionalEqual,
				})
			case token.Number:
				switch typeInfo.(type) {
				case *parser.TypeInfo_Int:
					tokenString := t.String()
					if strings.Contains(tokenString, ".") {
						panic("Cannot add float to int, this should be caught in the type checker.")
					}
					tokenInt, err := strconv.ParseInt(tokenString, 10, 0)
					if err != nil {
						panic(fmt.Sprintf("emitExpression:Int:Token: Cannot convert token string to int, error: %s", err))
					}
					opcodes = append(opcodes, bytecode.Code{
						Kind:  bytecode.Push,
						Value: tokenInt,
					})
				case *parser.TypeInfo_Float:
					panic("todo(Jake): Add support for floating point numbers")
				default:
					panic(fmt.Sprintf("emitExpression: Type %T cannot push number (\"%s\"), this should be caught by typechecker.", typeInfo, t.String()))
				}
			case token.Add:
				switch typeInfo := topNode.TypeInfo.(type) {
				case *parser.TypeInfo_Int:
					//*parser.TypeInfo_Float:
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.Add,
					})
				case *parser.TypeInfo_String:
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.AddString,
					})
				default:
					panic(fmt.Sprintf("emitExpression: Type %T does not support \"%s\", this should be caught by typechecker.", typeInfo, t.Kind.String()))
				}
			case token.String:
				_, ok := topNode.TypeInfo.(*parser.TypeInfo_String)
				if !ok {
					panic(fmt.Sprintf("emitExpression: Type %T cannot push string (\"%s\"), this should be caught by typechecker.", typeInfo, t.String()))
				}
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.Push,
					Value: t.String(),
				})
			case token.KeywordFalse:
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.Push,
					Value: false,
				})
			case token.KeywordTrue:
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.Push,
					Value: true,
				})
			default:
				panic(fmt.Sprintf("emitExpression:Token: Unhandled token kind: \"%s\", this should be caught by typechecker.", t.Kind.String()))
			}
		case *ast.ArrayLiteral:
			typeInfo := node.TypeInfo.(*parser.TypeInfo_Array)
			underlyingTypeInfo := typeInfo.Underlying()
			nodes := node.Nodes()
			if len(nodes) == 0 {
				panic(fmt.Sprintf("emitExpression:ArrayLiteral: Must have at least one item / node. This should be caught by typechecker"))
			}

			// Get bytecode to append per item in array literal
			var appendPopArray bytecode.Code
			switch underlyingTypeInfo := underlyingTypeInfo.(type) {
			case *parser.TypeInfo_String:
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.PushAllocArrayString,
					Value: len(nodes),
				})
				appendPopArray = bytecode.Code{
					Kind: bytecode.AppendPopArrayString,
				}
			default:
				panic(fmt.Sprintf("emitExpression:ArrayLiteral: Unhandled type %T", underlyingTypeInfo))
			}

			for _, node := range nodes {
				node := node.(*ast.Expression)
				if len(node.Nodes()) == 0 {
					panic("emitExpression:ArrayLiteral: No nodes in expression, this should be caught by typechecker")
				}
				opcodes = emit.emitExpression(opcodes, node)
				opcodes = append(opcodes, appendPopArray)
			}
		case *ast.StructLiteral:
			structLiteral := node
			typeInfo, ok := topNode.TypeInfo.(*parser.TypeInfo_Struct)
			if !ok {
				panic(fmt.Sprintf("emitExpression: Type %T cannot push struct literal (\"%s\"), this should be caught by typechecker.", typeInfo, structLiteral.Name))
			}
			structDef := typeInfo.Definition()
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
						opcodes = emit.emitNewFromType(opcodes, exprNode.TypeInfo)
						//panic(fmt.Sprintf("emitExpression:TypeInfo_Struct: Missing value for field \"%s\" on \"%s :: struct\", type checker should enforce that you need all fields.", structField.Name, structDef.Name))
					} else {
						opcodes = emit.emitExpression(opcodes, exprNode)
					}
					opcodes = append(opcodes, bytecode.Code{
						Kind:  bytecode.StorePopStructField,
						Value: offset,
					})
				}
			}
		default:
			panic(fmt.Sprintf("emitExpression:%T: Unhandled type %T", typeInfo, node))
		}
	}
	return opcodes
}

func (emit *Emitter) emitLeftHandSide(opcodes []bytecode.Code, leftHandSide []ast.Token) []bytecode.Code {
	name := leftHandSide[0].String()
	varInfo, ok := emit.scope.Get(name)
	if !ok {
		return nil
	}
	if len(leftHandSide) > 1 {
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.PushStackVar,
			Value: varInfo.stackPos,
		})
	}
	return opcodes
}

func emitHTMLComponentDefinition(node *ast.HTMLComponentDefinition) *bytecode.Block {
	emit := New()

	opcodes := make([]bytecode.Code, 0, 15)
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.Label,
		Value: "htmldefinition:" + node.Name.String(),
	})

	if structDef := node.Struct; structDef != nil {
		// Struct size + "children" keyword
		parameterCount := len(structDef.Fields) + 1

		for i := len(structDef.Fields) - 1; i >= 0; i-- {
			structField := structDef.Fields[i]
			exprNode := &structField.Expression
			opcodes = emit.emitParameter(opcodes, structField.Name.String(), exprNode.TypeInfo, (parameterCount-1)-emit.stackPos)
			emit.stackPos++
		}
		// Add special optional "children" parameter as first parameter
		opcodes = emit.emitParameter(opcodes, "children", nil, (parameterCount-1)-emit.stackPos)
		emit.stackPos++
	}

	for _, node := range node.Nodes() {
		opcodes = emit.emitStatement(opcodes, node)
	}

	// Implicit 'return' for top-level HTML nodes
	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.PushReturnHTMLNodeArray,
	})
	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.Return,
	})

	block := bytecode.NewBlock(bytecode.BlockHTMLComponentDefinition)
	block.Opcodes = opcodes
	block.StackSize = emit.stackPos
	return block
}

func (emit *Emitter) emitParameter(opcodes []bytecode.Code, name string, typeInfo types.TypeInfo, stackPos int) []bytecode.Code {
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.Store,
		Value: stackPos,
	})
	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.Pop,
	})
	structTypeInfo, ok := typeInfo.(*parser.TypeInfo_Struct)
	if !ok {
		emit.scope.DeclareSet(name, VariableInfo{
			stackPos: stackPos,
		})
		return opcodes
	}
	emit.scope.DeclareSet(name, VariableInfo{
		kind:      VariableStruct,
		structDef: structTypeInfo.Definition(),
		stackPos:  stackPos,
	})
	return opcodes
}

func emitProcedureDefinition(node *ast.ProcedureDefinition) *bytecode.Block {
	emit := New()

	opcodes := make([]bytecode.Code, 0, 35)
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.Label,
		Value: "procedure:" + node.Name.String(),
	})

	stackSize := len(node.Parameters)
	emit.stackPos = stackSize
	for i := len(node.Parameters) - 1; i >= 0; i-- {
		parameter := node.Parameters[i]
		opcodes = emit.emitParameter(opcodes, parameter.Name.String(), parameter.TypeInfo, i)
	}

	for _, node := range node.Nodes() {
		opcodes = emit.emitStatement(opcodes, node)
	}

	block := bytecode.NewBlock(bytecode.BlockProcedure)
	block.Opcodes = opcodes
	block.StackSize = emit.stackPos
	return block
}

func (emit *Emitter) emitGlobalScope(node ast.Node) {
	switch node := node.(type) {
	case *ast.ProcedureDefinition:
		block := emitProcedureDefinition(node)
		ok := emit.registerSymbol(node.Name.String(), block)
		if !ok {
			panic(fmt.Sprintf("Procedure name %s is used already. This should be caught in the typechecker.", node.Name.String()))
		}
	case *ast.HTMLComponentDefinition:
		block := emitHTMLComponentDefinition(node)
		ok := emit.registerSymbol(node.Name.String(), block)
		if !ok {
			panic(fmt.Sprintf("HTML Component name %s is used already. This should be caught in the typechecker.", node.Name.String()))
		}
	}
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
		fmt.Printf("todo(Jake): *ast.CSSDefinition\n")

		/*nodes := node.Nodes()
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
		}*/
	case *ast.ProcedureDefinition:
		// Handled previously.
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
		//debugOpcodes(opcodes)
		panic("todo(Jake): *ast.CSSRule")
	case *ast.CSSProperty:
		panic("todo(Jake): *ast.CSSProperty")
	case *ast.DeclareStatement:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Label,
			Value: "DeclareStatement",
		})
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
				kind:      VariableStruct,
				stackPos:  emit.stackPos,
				structDef: varStructDef,
			})
			emit.stackPos++
		}
	case *ast.Expression:
		// todo(Jake): 2018-02-01
		//
		// Disallow *ast.Expression in typechecker if the context
		// is not a ":: html" definition.
		//
		switch typeInfo := node.TypeInfo.(type) {
		case *parser.TypeInfo_String:
			opcodes = emit.emitExpression(opcodes, node)
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.CastToHTMLText,
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.AppendPopHTMLElementToHTMLElement,
			})
		case *parser.TypeInfo_HTMLNode:
			opcodes = emit.emitExpression(opcodes, node)
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.AppendPopHTMLElementToHTMLElement,
			})
		default:
			panic(fmt.Sprintf("emitStatement:Expression: Unhandled type %T", typeInfo))
		}
	case *ast.Call:
		switch node.Kind() {
		case ast.CallProcedure:
			opcodes = emit.emitProcedureCall(opcodes, node)
			resultTypeInfo := node.Definition.TypeInfo
			if resultTypeInfo != nil {
				opcodes = append(opcodes, bytecode.Code{
					Kind: bytecode.Pop,
				})
			}
		case ast.CallHTMLNode:
			opcodes = emit.emitHTMLNode(opcodes, node)
		default:
			panic(fmt.Errorf("emitExpression: Unhandled *ast.Call kind: %s", node.Name))
		}
	case *ast.ArrayAppendStatement:
		leftHandSide := node.LeftHandSide
		var storeOffset int
		opcodes, storeOffset = emit.emitVariableIdentWithProperty(opcodes, leftHandSide)
		if lastCode := &opcodes[len(opcodes)-1]; lastCode.Kind == bytecode.ReplaceStructFieldVar {
			// NOTE(Jake): 2017-12-29
			//
			// Change final bytecode to be pushed
			// so it can be stored using `StorePopStructField`
			// below.
			//
			lastCode.Kind = bytecode.PushStructFieldVar
		}
		exprNode := &node.Expression
		opcodes = emit.emitExpression(opcodes, exprNode)
		typeInfo := exprNode.TypeInfo
		switch typeInfo := typeInfo.(type) {
		case *parser.TypeInfo_String:
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.AppendPopArrayString,
			})
		default:
			panic(fmt.Errorf("emitExpression:ArrayAppend: Unhandled kind: %s", typeInfo))
		}
		if len(leftHandSide) > 1 {
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.StorePopStructField,
				Value: storeOffset,
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.Pop,
			})
			break
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Store,
			Value: storeOffset,
		})
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.Pop,
		})
	case *ast.OpStatement:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Label,
			Value: "OpStatement",
		})
		leftHandSide := node.LeftHandSide

		var storeOffset int
		opcodes, storeOffset = emit.emitVariableIdentWithProperty(opcodes, leftHandSide)
		lastCode := &opcodes[len(opcodes)-1]
		if lastCode.Kind == bytecode.ReplaceStructFieldVar {
			// NOTE(Jake): 2017-12-29
			//
			// Change final bytecode to be pushed
			// so it can be stored using `StorePopStructField`
			// below.
			//
			lastCode.Kind = bytecode.PushStructFieldVar
		}

		// NOTE(Jake): 2017-12-29
		//
		// If we're doing a simple set ("=") and not an
		// add-equal/etc ("+="), then we can remove the
		// last opcode as it's instruction pushes the variables
		// current value to the register stack.
		//
		isSettingVariable := node.Operator.Kind == token.Equal
		if isSettingVariable {
			opcodes = opcodes[:len(opcodes)-1]
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

		if len(leftHandSide) > 1 {
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.StorePopStructField,
				Value: storeOffset,
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.Pop,
			})
			break
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Store,
			Value: storeOffset,
		})
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.Pop,
		})
	case *ast.Return:
		if len(node.Expression.Nodes()) > 0 {
			opcodes = emit.emitExpression(opcodes, &node.Expression)
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.Return,
		})
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
	for _, node := range node.Nodes() {
		emit.emitGlobalScope(node)
	}

	opcodes := make([]bytecode.Code, 0, 10)
	for _, node := range node.Nodes() {
		opcodes = emit.emitStatement(opcodes, node)
	}
	codeBlock := new(bytecode.Block)
	codeBlock.Opcodes = opcodes
	codeBlock.StackSize = emit.stackPos
	debugOpcodes(opcodes)
	fmt.Printf("Final bytecode output above\nStack Size: %d\n", codeBlock.StackSize)
	return codeBlock
}
