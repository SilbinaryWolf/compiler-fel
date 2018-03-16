package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/types"
)

type TypeInfoManager struct {
	registeredTypes map[TypeInfo_Identifier]TypeInfo

	// built-in
	intInfo      TypeInfo_Int
	floatInfo    TypeInfo_Float
	stringInfo   TypeInfo_String
	boolInfo     TypeInfo_Bool
	htmlNodeInfo TypeInfo_HTMLNode

	// built-in structs
	workspaceInfo *TypeInfo_Struct
}

func (manager *TypeInfoManager) Init() {
	if manager.registeredTypes != nil {
		panic("Cannot initialize TypeInfoManager twice.")
	}
	manager.registeredTypes = make(map[TypeInfo_Identifier]TypeInfo)

	// Primitives
	manager.register("int", manager.NewTypeInfoInt())
	manager.register("string", manager.NewTypeInfoString())
	manager.register("float", manager.NewTypeInfoFloat())
	manager.register("bool", manager.NewTypeInfoBool())

	// Internal types
	manager.workspaceInfo = manager.NewInternalStructInfo(
		"Workspace",
		[]TypeInfo_StructField{
			manager.NewInternalStructField("template_input_directory", "string"),
			manager.NewInternalStructField("template_output_directory", "string"),
			manager.NewInternalStructField("css_output_directory", "string"),
			manager.NewInternalStructField("css_files", "[]string"),
		},
	)
}

func (manager *TypeInfoManager) Clear() {
	if manager.registeredTypes == nil {
		panic("Cannot clear TypeInfoManager if it's already cleared..")
	}
	manager.registeredTypes = nil
}

func (manager *TypeInfoManager) register(name string, typeInfo TypeInfo) {
	key := TypeInfo_Identifier{
		Name:       name,
		ArrayDepth: 0,
	}
	_, ok := manager.registeredTypes[key]
	if ok {
		panic(fmt.Sprintf("Already registered \"%s\" type.", name))
	}
	manager.registeredTypes[key] = typeInfo
}

func (manager *TypeInfoManager) getByName(name string) TypeInfo {
	return manager.registeredTypes[TypeInfo_Identifier{
		Name:       name,
		ArrayDepth: 0,
	}]
}

func (manager *TypeInfoManager) get(identifier TypeInfo_Identifier) TypeInfo {
	name := identifier.Name
	arrayDepth := identifier.ArrayDepth
	resultType := manager.getByName(name)
	if resultType == nil {
		return nil
	}
	if arrayDepth > 0 {
		for i := 0; i < arrayDepth; i++ {
			underlyingType := resultType
			resultType = manager.NewTypeInfoArray(underlyingType)
		}
	}
	return resultType
}

type TypeInfo interface {
	String() string
}

type TypeInfo_Identifier struct {
	Name       string
	ArrayDepth int
}

// Int
type TypeInfo_Int struct{}

func (info *TypeInfo_Int) String() string { return "int" }

func (manager *TypeInfoManager) NewTypeInfoInt() *TypeInfo_Int {
	return &manager.intInfo
}

// Float
type TypeInfo_Float struct{}

func (info *TypeInfo_Float) String() string { return "float" }

func (manager *TypeInfoManager) NewTypeInfoFloat() *TypeInfo_Float {
	return &manager.floatInfo
}

// String
type TypeInfo_String struct{}

func (info *TypeInfo_String) String() string { return "string" }

func (manager *TypeInfoManager) NewTypeInfoString() *TypeInfo_String {
	return &manager.stringInfo
}

// Array
type TypeInfo_Array struct {
	underlying TypeInfo
}

func (info *TypeInfo_Array) String() string       { return "[]" + info.underlying.String() }
func (info *TypeInfo_Array) Underlying() TypeInfo { return info.underlying }

func (manager *TypeInfoManager) NewTypeInfoArray(underlying TypeInfo) *TypeInfo_Array {
	result := new(TypeInfo_Array)
	result.underlying = underlying
	return result
}

// Procedure
type TypeInfo_Procedure struct {
	name       string
	definition *ast.ProcedureDefinition
}

func (info *TypeInfo_Procedure) String() string                       { return "procedure " + info.name + "()" }
func (info *TypeInfo_Procedure) Definition() *ast.ProcedureDefinition { return info.definition }

func (manager *TypeInfoManager) NewProcedureInfo(definiton *ast.ProcedureDefinition) *TypeInfo_Procedure {
	result := new(TypeInfo_Procedure)
	result.name = definiton.Name.String()
	result.definition = definiton
	return result
}

type TypeInfo_Bool struct{}

func (info *TypeInfo_Bool) String() string { return "bool" }

func (manager *TypeInfoManager) NewTypeInfoBool() *TypeInfo_Bool {
	return &manager.boolInfo
}

// HTMLNode
type TypeInfo_HTMLNode struct{}

func (info *TypeInfo_HTMLNode) String() string { return "html node" }

func (manager *TypeInfoManager) NewHTMLNode() *TypeInfo_HTMLNode {
	return &manager.htmlNodeInfo
}

// HTMLComponentNode
/*type TypeInfo_HTMLComponentNode struct {
	name       string
	definition *ast.HTMLComponentDefinition
}

func NewHTMLComponentNode(name string) TypeInfo {
	result := new(TypeInfo_HTMLComponentNode)
	result.name = definition.Name.String()
	//result.definition = definiton
	return result
}*/

// Struct
type TypeInfo_Struct struct {
	name   string
	fields []TypeInfo_StructField
	//definition *ast.StructDefinition
}

func (info *TypeInfo_Struct) String() string                 { return info.name }
func (info *TypeInfo_Struct) Name() string                   { return info.name }
func (info *TypeInfo_Struct) Fields() []TypeInfo_StructField { return info.fields }

//func (info *TypeInfo_Struct) Definition() *ast.StructDefinition { return info.definition }

func (info *TypeInfo_Struct) GetFieldByName(name string) *TypeInfo_StructField {
	fields := info.Fields()
	for i := 0; i < len(fields); i++ {
		field := &fields[i]
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (manager *TypeInfoManager) NewStructInfo(definiton *ast.StructDefinition) *TypeInfo_Struct {
	result := new(TypeInfo_Struct)
	result.name = definiton.Name.String()
	result.fields = make([]TypeInfo_StructField, 0, len(definiton.Fields))
	for i, field := range definiton.Fields {
		result.fields = append(result.fields, TypeInfo_StructField{
			Index: i,
			Name:  field.Name.String(),
			TypeIdentifier: TypeInfo_Identifier{
				Name:       field.TypeIdentifier.Name.String(),
				ArrayDepth: field.TypeIdentifier.ArrayDepth,
			},
			TypeInfo:     field.TypeInfo,
			DefaultValue: field.Expression,
		})
	}
	return result
}

func (manager *TypeInfoManager) NewInternalStructInfo(name string, fields []TypeInfo_StructField) *TypeInfo_Struct {
	for i := range fields {
		field := &fields[i]
		field.Index = i
	}
	result := new(TypeInfo_Struct)
	result.name = name
	result.fields = fields
	return result
}

type TypeInfo_StructField struct {
	Name           string
	Index          int
	TypeIdentifier TypeInfo_Identifier
	TypeInfo       types.TypeInfo
	DefaultValue   ast.Expression
}

func (manager *TypeInfoManager) NewInternalStructField(name string, typeIdentName string) TypeInfo_StructField {
	arrayDepth := 0
	for typeIdentName[0] == '[' {
		if typeIdentName[1] == ']' {
			typeIdentName = typeIdentName[2:]
			arrayDepth++
		}
	}
	//arrayDepth := strings.Count(typeIdentName, "[]")
	typeIdent := TypeInfo_Identifier{
		Name:       typeIdentName,
		ArrayDepth: arrayDepth,
	}
	typeInfo := manager.get(typeIdent)
	if typeInfo == nil {
		panic(fmt.Sprintf("NewInternalStructField: Cannot find type info for %s on property %s", typeIdentName, name))
	}
	result := TypeInfo_StructField{
		Name:           name,
		TypeIdentifier: typeIdent,
		TypeInfo:       typeInfo,
	}
	result.DefaultValue.TypeInfo = typeInfo
	return result
}

// Internal Structs
func (manager *TypeInfoManager) InternalWorkspaceStruct() *TypeInfo_Struct {
	return manager.workspaceInfo
}

// Functions
func (p *Parser) DetermineType(node *ast.Type) types.TypeInfo {
	return p.typeinfo.get(TypeInfo_Identifier{
		Name:       node.Name.String(),
		ArrayDepth: node.ArrayDepth,
	})
}

func TypeEquals(a TypeInfo, b TypeInfo) bool {
	aAsArray, aOk := a.(*TypeInfo_Array)
	bAsArray, bOk := b.(*TypeInfo_Array)
	if aOk && bOk {
		return TypeEquals(aAsArray.underlying, bAsArray.underlying)
	}
	if a == b {
		return true
	}
	return false
}
