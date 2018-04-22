package emitter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/data"
	//"github.com/silbinarywolf/compiler-fel/data"
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
	structTypeInfo *types.Struct
}

type Scope struct {
	mapToInfo map[string]VariableInfo
	parent    *Scope
	stackPos  int
}

type Emitter struct {
	symbols           map[string]*bytecode.Block
	unresolvedSymbols map[string]*bytecode.Block
	workspaces        []*bytecode.Block
	fileOptions       FileOptions
	htmlElementStack  []string // mostly for debug purposes, can possibly be removed
	EmitterScope
}

type EmitterScope struct {
	scope *Scope
}

type FileOptions struct {
	IsTemplateFile bool // implicitHTMLFragmentReturn
}

func (emit *Emitter) IsTemplateFile() bool {
	return emit.fileOptions.IsTemplateFile
}

func (emit *Emitter) PushScope() {
	oldScope := emit.scope

	newScope := new(Scope)
	newScope.mapToInfo = make(map[string]VariableInfo)
	if oldScope != nil {
		newScope.parent = oldScope
		newScope.stackPos = oldScope.stackPos
	}

	emit.scope = newScope
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
	emit.unresolvedSymbols = make(map[string]*bytecode.Block)
	emit.workspaces = make([]*bytecode.Block, 0, 3)
	emit.PushScope()
	return emit
}

func (emit *Emitter) Workspaces() []*bytecode.Block {
	return emit.workspaces
}

func (emit *Emitter) EmitGlobalScope(nodes []*ast.File) {
	for _, astFile := range nodes {
		for _, node := range astFile.Nodes() {
			emit.emitGlobalScope(node)
		}
	}
	if len(emit.unresolvedSymbols) > 0 {
		panic("todo(Jake): Handle unresolved symbols.")
	}
}

func (emit *Emitter) EmitBytecode(node *ast.File, fileOptions FileOptions) *bytecode.Block {
	oldOptions := emit.fileOptions
	emit.fileOptions = fileOptions
	defer func() {
		emit.fileOptions = oldOptions
	}()
	isTemplateFile := emit.IsTemplateFile()

	// Emit bytecode
	opcodes := make([]bytecode.Code, 0, 50)
	codeBlockType := bytecode.BlockDefault
	{
		if isTemplateFile {
			// NOTE(Jake): 2018-02-17
			//
			// For template files we need this for the return value
			//
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.PushAllocHTMLFragment,
			})
			codeBlockType = bytecode.BlockTemplate
		}
		for _, node := range node.Nodes() {
			opcodes = emit.emitStatement(opcodes, node)
		}
		if isTemplateFile {
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.Return,
			})
		}
	}

	codeBlock := bytecode.NewBlock(node.Filepath, codeBlockType)
	codeBlock.Opcodes = opcodes
	codeBlock.StackSize = emit.scope.stackPos
	codeBlock.HasReturnValue = codeBlockType == bytecode.BlockTemplate
	//debugOpcodes(opcodes)
	//fmt.Printf("Final bytecode output above\nStack Size: %d\n", codeBlock.StackSize)
	return codeBlock
}

func (emit *Emitter) EmitCSSDefinition(def *ast.CSSDefinition) *bytecode.Block {
	name := def.Name.String()

	// Emit bytecode
	opcodes := make([]bytecode.Code, 0, 50)
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.PushAllocCSSDefinition,
		Value: data.NewCSSDefinition(name),
	})

	for _, node := range def.Nodes() {
		opcodes = emit.emitStatement(opcodes, node)
	}

	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.Return,
	})

	debugOpcodes(opcodes)

	// Create code block
	codeBlock := bytecode.NewBlock(name, bytecode.BlockCSSDefinition)
	codeBlock.Opcodes = opcodes
	codeBlock.StackSize = emit.scope.stackPos
	codeBlock.HasReturnValue = true

	//debugOpcodes(codeBlock.Opcodes)
	//fmt.Printf("Final bytecode output above\nStack Size: %d\n", codeBlock.StackSize)
	//panic("todo(Jake): EmitCSSDefinition")
	return codeBlock
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
		if symbol, ok := emit.unresolvedSymbols[name]; ok {
			// NOTE(Jake): 2018-04-13
			//
			// If a symbol isn't found, we create it and use it in
			// an uninitialized state. Later to resolve the symbol, we
			// simply copy the data over the "placeholder"
			//
			*symbol = *block
			delete(emit.unresolvedSymbols, name)
			return true
		}
		return false
	}
	emit.symbols[name] = block
	return true
}

func (emit *Emitter) registerWorkspace(block *bytecode.Block) {
	emit.workspaces = append(emit.workspaces, block)
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
	case *types.Int:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: int(0),
		})
	case *types.Float:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: float64(0),
		})
	case *types.String:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: "",
		})
	case *types.Bool:
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Push,
			Value: false,
		})
	case *types.Array:
		underlyingType := typeInfo.Underlying()
		switch underlyingType.(type) {
		case *types.String:
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.PushAllocArrayString,
				Value: 0,
			})
		default:
			panic(fmt.Sprintf("emitNewFromType:Array: Unhandled type %T", underlyingType))
		}
	case *types.Struct:
		name := typeInfo.Name()
		fields := typeInfo.Fields()
		if name == "" || len(fields) == 0 {
			panic("emitExpression: TypeInfo_Struct: Missing field data, this should be handled in the type checker.")
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.PushAllocStruct,
			Value: typeInfo,
		})
		for offset, structField := range fields {
			// todo(Jake): Continue decoupling Struct typeinfo from struct ast
			exprNode := &structField.DefaultValue
			fieldTypeInfo := exprNode.TypeInfo
			if fieldTypeInfo == nil {
				panic(fmt.Sprintf("emitExpression: Missing type info on property for \"%s :: struct { %s }\"", name, structField.Name))
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
		panic(fmt.Sprintf("Missing declaration for \"%s\", this should be caught in the type checker.", name))
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
	var lastPropertyField *types.StructField
	if len(leftHandSide) <= 1 {
		return opcodes, varInfo.stackPos
	}
	structTypeInfo := varInfo.structTypeInfo
	if structTypeInfo == nil {
		panic(fmt.Sprintf("emitStatement: Expected parameter %s to be a struct, this should be set when declaring a new variable (if applicable)", name))
	}
	for i := 1; i < len(leftHandSide)-1; i++ {
		if structTypeInfo == nil {
			panic("emitStatement: Non-struct cannot have properties. This should be caught in the typechecker.")
		}
		fieldName := leftHandSide[i].String()
		field := structTypeInfo.GetFieldByName(fieldName)
		if field == nil {
			panic(fmt.Sprintf("emitStatement: \"%s :: struct\" does not have property \"%s\". This should be caught in the typechecker.", structTypeInfo.Name, fieldName))
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.ReplaceStructFieldVar,
			Value: field.Index,
		})
		if typeInfo, ok := field.TypeInfo.(*types.Struct); ok {
			structTypeInfo = typeInfo
		}
	}
	fieldName := leftHandSide[len(leftHandSide)-1].String()
	lastPropertyField = structTypeInfo.GetFieldByName(fieldName)
	if lastPropertyField == nil {
		panic(fmt.Sprintf("emitStatement: \"%s :: struct\" does not have property \"%s\". This should be caught in the typechecker.", structTypeInfo.Name, fieldName))
	}
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.ReplaceStructFieldVar,
		Value: lastPropertyField.Index,
	})
	return opcodes, lastPropertyField.Index()
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
			block = bytecode.NewUnresolvedBlock(name, bytecode.BlockHTMLComponentDefinition)
			emit.symbols[name] = block
			emit.unresolvedSymbols[name] = block
			//fmt.Printf("Unresovled Symbol added: %s", name)
			//panic(fmt.Sprintf("Missing HTML component \"%s\" symbol. Either this is:\n- Uncaught in the typechecker.\nor\n- A component that hasnt been emitted.", name))
		}

		// If definition has used the "children" keyword
		//
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
		// NOTE(Jake): 2018-02-17
		//
		// Assumption is that we'll have a HTMLFragment on the stack
		// to pop from this call.
		//
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.AppendPopHTMLElementToHTMLElement,
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

	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.AppendPopHTMLElementToHTMLElement,
	})
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
				case *types.Int:
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
				case *types.Float:
					panic("todo(Jake): Add support for floating point numbers")
				default:
					panic(fmt.Sprintf("emitExpression: Type %T cannot push number (\"%s\"), this should be caught by typechecker.", typeInfo, t.String()))
				}
			case token.Add:
				switch typeInfo := topNode.TypeInfo.(type) {
				case *types.Int:
					//*types.Float:
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.Add,
					})
				case *types.String:
					opcodes = append(opcodes, bytecode.Code{
						Kind: bytecode.AddString,
					})
				default:
					panic(fmt.Sprintf("emitExpression: Type %T does not support \"%s\", this should be caught by typechecker.", typeInfo, t.Kind.String()))
				}
			case token.String:
				_, ok := topNode.TypeInfo.(*types.String)
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
			typeInfo := node.TypeInfo.(*types.Array)
			underlyingTypeInfo := typeInfo.Underlying()
			nodes := node.Nodes()
			if len(nodes) == 0 {
				panic(fmt.Sprintf("emitExpression:ArrayLiteral: Must have at least one item / node. This should be caught by typechecker"))
			}

			// Get bytecode to append per item in array literal
			var appendPopArray bytecode.Code
			switch underlyingTypeInfo := underlyingTypeInfo.(type) {
			case *types.String:
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
			structTypeInfo, ok := topNode.TypeInfo.(*types.Struct)
			if !ok {
				panic(fmt.Sprintf("emitExpression: Type %T cannot push struct literal (\"%s\"), this should be caught by typechecker.", typeInfo, structLiteral.Name))
			}
			if len(structLiteral.Fields) == 0 {
				// NOTE(Jake) 2017-12-28
				// If using struct literal syntax "MyStruct{}" without fields, assume all fields
				// use default values.
				opcodes = emit.emitNewFromType(opcodes, typeInfo)
			} else {
				structTypeInfoFields := structTypeInfo.Fields()
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.PushAllocStruct,
					Value: structTypeInfo,
				})

				for offset, structField := range structTypeInfoFields {
					name := structField.Name

					exprNode := &structField.DefaultValue
					for _, literalField := range structLiteral.Fields {
						if name == literalField.Name.String() {
							exprNode = &literalField.Expression
							break
						}
					}
					if fieldTypeInfo := exprNode.TypeInfo; fieldTypeInfo == nil {
						panic(fmt.Sprintf("emitExpression: Missing type info on property for \"%s :: struct { %s }\"", structTypeInfo.Name, structField.Name))
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

func (emit *Emitter) emitHTMLComponentDefinition(node *ast.HTMLComponentDefinition) *bytecode.Block {
	// Reset scope / html nest variables
	oldEmitterScope := emit.EmitterScope
	emit.EmitterScope = EmitterScope{}
	emit.PushScope()
	defer func() {
		emit.EmitterScope = oldEmitterScope
	}()

	opcodes := make([]bytecode.Code, 0, 15)
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.Label,
		Value: "htmldefinition:" + node.Name.String(),
	})

	// Struct size + "children" keyword
	{
		// todo(Jake): 2018-02-13
		//
		// Only add support for "children" if the "children"
		// keyword is used (and not taken by the props)
		//
		hasChildren := true
		parameterCount := 0
		if hasChildren {
			parameterCount++
		}
		if structDef := node.Struct; structDef != nil {
			parameterCount += len(structDef.Fields)
			for i := len(structDef.Fields) - 1; i >= 0; i-- {
				structField := structDef.Fields[i]
				exprNode := &structField.Expression
				opcodes = emit.emitParameter(opcodes, structField.Name.String(), exprNode.TypeInfo, (parameterCount-1)-emit.scope.stackPos)
				emit.scope.stackPos++
			}
		}
		if hasChildren {
			// Add special optional "children" parameter as first parameter
			opcodes = emit.emitParameter(opcodes, "children", nil, (parameterCount-1)-emit.scope.stackPos)
			emit.scope.stackPos++
		}
	}

	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.PushAllocHTMLFragment,
	})
	emit.scope.stackPos++

	for _, node := range node.Nodes() {
		opcodes = emit.emitStatement(opcodes, node)
	}

	// Implicit 'return' for top-level HTML nodes
	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.Return,
	})

	block := bytecode.NewBlock(node.Name.String(), bytecode.BlockHTMLComponentDefinition)
	block.Opcodes = opcodes
	block.StackSize = emit.scope.stackPos
	block.HasReturnValue = true
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
	structTypeInfo, ok := typeInfo.(*types.Struct)
	if !ok {
		emit.scope.DeclareSet(name, VariableInfo{
			stackPos: stackPos,
		})
		return opcodes
	}
	emit.scope.DeclareSet(name, VariableInfo{
		kind:           VariableStruct,
		structTypeInfo: structTypeInfo,
		stackPos:       stackPos,
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
	emit.scope.stackPos = stackSize
	for i := len(node.Parameters) - 1; i >= 0; i-- {
		parameter := node.Parameters[i]
		opcodes = emit.emitParameter(opcodes, parameter.Name.String(), parameter.TypeInfo, i)
	}

	for _, node := range node.Nodes() {
		opcodes = emit.emitStatement(opcodes, node)
	}

	// todo(Jake): If last opcode is not a return statement, add a return statement.
	lastOpcode := &opcodes[len(opcodes)-1]
	if lastOpcode.Kind != bytecode.Return {
		panic("emitProcedureDefinition: Automatically add return")
	}

	block := bytecode.NewBlock(node.Name.String(), bytecode.BlockProcedure)
	block.Opcodes = opcodes
	block.StackSize = emit.scope.stackPos
	block.HasReturnValue = node.TypeInfo != nil
	return block
}

func (emit *Emitter) emitCSSSelectors(selectors []ast.CSSSelector) []data.CSSSelector {
	// NOTE(Jake): 2018-04-19
	//
	// Selector data is all built during this emitter step
	// as I'm fairly certain I will not want interpolation
	// in CSS classes or any other fancy features.
	//
	// The reasoning is that it'd make static analysis of
	// used CSS rules harder.
	//
	// So we'll build the data structures here and pass them off
	// to some bytecode.
	//
	resultSelectors := make([]data.CSSSelector, 0, 10)
	for _, selector := range selectors {
		selectorPartNodes := selector.Nodes()
		selector := data.NewCSSSelector(len(selectorPartNodes))
		for _, selectorPartNode := range selectorPartNodes {
			switch selectorPartNode := selectorPartNode.(type) {
			case *ast.Token:
				value := selectorPartNode.String()
				switch selectorPartNode.Kind {
				case token.Identifier:
					switch value[0] {
					case '.':
						selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindClass, value))
					case '#':
						selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindID, value))
					default:
						selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindTag, value))
					}
				case token.AtKeyword:
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindAtKeyword, value))
				case token.Number:
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindNumber, value))
				case token.Colon: // :
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindColon, value))
				case token.DoubleColon: // ::
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindDoubleColon, value))
				case token.Whitespace: // ` `
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindAncestor, value))
				case token.GreaterThan: // >
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindChild, value))
				case token.Add: // +
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindAdjacent, value))
				case token.Tilde: // ~
					selector.AddPart(data.NewCSSSelectorPart(data.SelectorPartKindSibling, value))
				default:
					if selectorPartNode.IsOperator() {
						panic("todo(Jake): Fixme (or add support for operator in above `switch`)")
						// 	selectorPartString := selectorPartNode.String()
						// 	selectorList = append(selectorList, data.CSSSelectorOperator{
						// 		Operator: selectorPartString,
						// 	})
						// 	continue
					}
					panic(fmt.Sprintf("emitCSSRule(): Unhandled selector part kind: %s", selectorPartNode.Kind.String()))
				}
			case *ast.CSSAttributeSelector:
				if hasValueSet := selectorPartNode.Operator.Kind != 0; hasValueSet {
					// Handle "input[name="Name"]" selector part
					selector.AddPart(data.NewCSSSelectorAttributePart(
						selectorPartNode.Name.String(),
						selectorPartNode.Operator.String(),
						selectorPartNode.Value.String(),
					))
					break
				}
				// Handle "input[name]" selector part
				selector.AddPart(data.NewCSSSelectorAttributePart(
					selectorPartNode.Name.String(),
					"",
					"",
				))
			case *ast.CSSSelector:
				// todo(Jake)
				panic(fmt.Sprintf("todo(Jake): Fix this, %v", selectorPartNode.Nodes()))
			//subSelectorList := program.evaluateSelector(selectorPartNode.Nodes())
			//selectorList = append(selectorList, subSelectorList)

			//for _, token := range selectorPartNode.ChildNodes {
			//	value += token.String() + " "
			//}
			//value = value[:len(value)-1]
			default:
				panic(fmt.Sprintf("emitCSSRule(): Unhandled selector type: %T", selectorPartNode))
			}
		}
		resultSelectors = append(resultSelectors, selector)
	}
	return resultSelectors
}

func (emit *Emitter) emitCSSRule(opcodes []bytecode.Code, node *ast.CSSRule) []bytecode.Code {
	emit.PushScope()
	defer emit.PopScope()

	// todo(Jake): Track cssRule depth in emitter and handle nested cases below
	/*switch node.Kind() {
	case ast.CSSKindRule:
		// Nested selectors
		panic("Nested selectors are not allowed. This should be caught by the typechecker.")
	case ast.CSSKindAtKeyword:
		panic("todo(Jake): Handle @media nested selector")
		// Setup rule node
		selectors := emit.emitCSSSelectors(node.Selectors())
		mediaRuleNode := data.NewCSSRule(selectors)
		mediaRuleNode.AddRule(resultRule)

		// Get parent selector
		//for _, parentSelectorListNode := range parentCSSRule.Selectors {
		//	selectorList := make(data.CSSSelector, 0, len(parentSelectorListNode))
		//	selectorList = append(selectorList, parentSelectorListNode...)
		//	ruleNode.Selectors = append(ruleNode.Selectors, selectorList)
		//}
		//mediaRuleNode.Rules = append(mediaRuleNode.Rules, ruleNode)

		// Become the wrapping @media query
		resultRule = mediaRuleNode
	default:
		panic("emitCSSRule(): Unhandled CSSType.")
	}*/

	resultRule := data.NewCSSRule(emit.emitCSSSelectors(node.Selectors()))
	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.PushAllocCSSRule,
		Value: resultRule,
	})

	// Attach CSS properties
	for _, node := range node.Nodes() {
		opcodes = emit.emitStatement(opcodes, node)
	}

	opcodes = append(opcodes, bytecode.Code{
		Kind: bytecode.Pop,
	})

	//panic("todo(Jake): handle css rule emit")
	return opcodes
}

func (emit *Emitter) emitGlobalScope(node ast.Node) {
	switch node := node.(type) {
	case *ast.WorkspaceDefinition:
		block := emitWorkspaceDefinition(node)
		emit.registerWorkspace(block)
	case *ast.ProcedureDefinition:
		block := emitProcedureDefinition(node)
		ok := emit.registerSymbol(node.Name.String(), block)
		if !ok {
			panic(fmt.Sprintf("Procedure name %s is used already. This should be caught in the typechecker.", node.Name.String()))
		}
	case *ast.HTMLComponentDefinition:
		block := emit.emitHTMLComponentDefinition(node)
		ok := emit.registerSymbol(node.Name.String(), block)
		if !ok {
			panic(fmt.Sprintf("HTML Component name %s is used already. This should be caught in the typechecker.", node.Name.String()))
		}
	}
}

func emitWorkspaceDefinition(node *ast.WorkspaceDefinition) *bytecode.Block {
	emit := New()
	structTypeInfo := node.WorkspaceTypeInfo.(*types.Struct)
	//emit.PushScope()
	//defer emit.PopScope()

	opcodes := make([]bytecode.Code, 0, 35)
	{
		// Push "workspace" as first parameter
		workspaceStackPos := emit.scope.stackPos
		emit.scope.stackPos++
		emit.scope.DeclareSet("workspace", VariableInfo{
			kind:           VariableStruct,
			structTypeInfo: structTypeInfo,
			stackPos:       workspaceStackPos,
		})
		opcodes = emit.emitNewFromType(opcodes, structTypeInfo)
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.Store,
			Value: workspaceStackPos,
		})
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.Pop,
		})

		// Store name of workspace from ast in inaccessible " name" field.
		{
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.PushStackVar,
				Value: workspaceStackPos,
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.Push,
				Value: node.Name.String(),
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind:  bytecode.StorePopStructField,
				Value: structTypeInfo.GetFieldByName(" name").Index(),
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.Pop,
			})
		}

		for _, node := range node.Nodes() {
			opcodes = emit.emitStatement(opcodes, node)
		}
		opcodes = append(opcodes, bytecode.Code{
			Kind:  bytecode.PushStackVar,
			Value: workspaceStackPos,
		})
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.Return,
		})
	}

	block := bytecode.NewBlock(node.Name.String(), bytecode.BlockWorkspaceDefinition)
	block.Opcodes = opcodes
	block.StackSize = emit.scope.stackPos
	block.HasReturnValue = true
	return block
}

func (emit *Emitter) emitCSSDefinition(opcodes []bytecode.Code, definition *ast.CSSDefinition) []bytecode.Code {
	for _, node := range definition.Nodes() {
		switch node := node.(type) {
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
			panic("todo(Jake): *ast.CSSProperty" + node.Name.String())
		}
	}
	return opcodes
}

func (emit *Emitter) emitCSSProperty(opcodes []bytecode.Code, property *ast.CSSProperty) []bytecode.Code {
	for _, node := range property.Nodes() {
		switch node := node.(type) {
		case *ast.TokenList:
			opcodes, _ = emit.emitVariableIdentWithProperty(opcodes, node.Tokens())
		case *ast.Call:
			switch node.Kind() {
			case ast.CallProcedure:
				opcodes = emit.emitProcedureCall(opcodes, node)
			case ast.CallHTMLNode:
				if len(node.Nodes()) > 0 {
					panic("emitCSSProperty: Cannot have sub-elements for HTMLNode in an expression.")
				}
				panic("emitCSSProperty: Cannot use HTMLNode in expression currently.")
				//opcodes = emit.emitHTMLNode(opcodes, node)
			default:
				panic(fmt.Errorf("emitCSSProperty: Unhandled *ast.Call kind: %s", node.Name))
			}
		case *ast.Token:
			t := node.Token
			switch t.Kind {
			case token.Identifier:
				// NOTE(Jake): 2018-04-22
				//
				// Use a variable if it's defined, otherwise print
				// out the raw identifier.
				//
				name := t.String()
				_, ok := emit.scope.Get(name)
				if ok {
					opcodes = emit.emitVariableIdent(opcodes, t)
					break
				}
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.Push,
					Value: t.String(),
				})
			case token.Number,
				token.String:
				opcodes = append(opcodes, bytecode.Code{
					Kind:  bytecode.Push,
					Value: t.String(),
				})
			default: // ie. number, string
				panic(fmt.Sprintf("emitCSSProperty: Unhandled token kind: %s", node.Kind.String()))
			}
		default:
			panic(fmt.Sprintf("emitCSSProperty: Unhandled type: %T", node))
		}
	}

	opcodes = append(opcodes, bytecode.Code{
		Kind:  bytecode.AppendCSSPropertyToCSSRule,
		Value: property.Name.String(),
	})

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
	case *ast.WorkspaceDefinition,
		*ast.CSSDefinition,
		*ast.ProcedureDefinition:
		// Handled in emitGlobalScope()
	case *ast.CSSRule:
		opcodes = emit.emitCSSRule(opcodes, node)
	case *ast.CSSProperty:
		opcodes = emit.emitCSSProperty(opcodes, node)
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
			Value: emit.scope.stackPos,
		})
		opcodes = append(opcodes, bytecode.Code{
			Kind: bytecode.Pop,
		})

		{
			var varStructTypeInfo *types.Struct = nil
			if typeInfo, ok := typeInfo.(*types.Struct); ok {
				varStructTypeInfo = typeInfo
			}
			emit.scope.DeclareSet(nameString, VariableInfo{
				kind:           VariableStruct,
				stackPos:       emit.scope.stackPos,
				structTypeInfo: varStructTypeInfo,
			})
			emit.scope.stackPos++
		}
	case *ast.Expression:
		// todo(Jake): 2018-02-01
		//
		// Disallow *ast.Expression in typechecker if the context
		// is not a ":: html" definition.
		//
		switch typeInfo := node.TypeInfo.(type) {
		case *types.String:
			opcodes = emit.emitExpression(opcodes, node)
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.CastToHTMLText,
			})
			opcodes = append(opcodes, bytecode.Code{
				Kind: bytecode.AppendPopHTMLElementToHTMLElement,
			})
		case *types.HTMLNode:
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
			// NOTE(Jake): 2018-02-17
			//
			// Since the context of this is simply a statement
			// we just pop the return value. (if there is one)
			//
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
		case *types.String:
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
			case *types.String:
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
