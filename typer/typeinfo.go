package typer

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/types"
)

type TypeInfoManager struct {
	registeredTypes map[types.Identifier]types.TypeInfo

	// built-in
	intInfo      types.Int
	floatInfo    types.Float
	stringInfo   types.String
	boolInfo     TypeInfo_Bool
	htmlNodeInfo TypeInfo_HTMLNode

	// built-in structs
	workspaceInfo *types.Struct
}

func (manager *TypeInfoManager) Init() {
	if manager.registeredTypes != nil {
		panic("Cannot initialize TypeInfoManager twice.")
	}
	manager.registeredTypes = make(map[types.Identifier]types.TypeInfo)

	// Primitives
	manager.register("int", manager.NewTypeInfoInt())
	manager.register("string", manager.NewTypeInfoString())
	manager.register("float", manager.NewTypeInfoFloat())
	manager.register("bool", manager.NewTypeInfoBool())

	// Internal types
	manager.workspaceInfo = types.NewInternalStruct(
		"Workspace",
		[]types.StructField{
			manager.NewInternalStructField(" name", "string"),
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

func (manager *TypeInfoManager) register(name string, typeInfo types.TypeInfo) {
	key := types.Identifier{
		Name:       name,
		ArrayDepth: 0,
	}
	_, ok := manager.registeredTypes[key]
	if ok {
		panic(fmt.Sprintf("Already registered \"%s\" type.", name))
	}
	manager.registeredTypes[key] = typeInfo
}

func (manager *TypeInfoManager) getByName(name string) types.TypeInfo {
	return manager.registeredTypes[types.Identifier{
		Name:       name,
		ArrayDepth: 0,
	}]
}

func (manager *TypeInfoManager) get(identifier types.Identifier) types.TypeInfo {
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

// Int
func (manager *TypeInfoManager) NewTypeInfoInt() *types.Int {
	return &manager.intInfo
}

// Float
func (manager *TypeInfoManager) NewTypeInfoFloat() *types.Float {
	return &manager.floatInfo
}

// String
func (manager *TypeInfoManager) NewTypeInfoString() *types.String {
	return &manager.stringInfo
}

// Array
func (manager *TypeInfoManager) NewTypeInfoArray(underlying types.TypeInfo) *types.Array {
	return types.NewArray(underlying)
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

func (_ *TypeInfoManager) NewStructInfo(definiton *ast.StructDefinition) *types.Struct {
	return types.NewStruct(definiton)
}

func (manager *TypeInfoManager) NewInternalStructField(name string, typeIdentName string) types.StructField {
	arrayDepth := 0
	for typeIdentName[0] == '[' {
		if typeIdentName[1] == ']' {
			typeIdentName = typeIdentName[2:]
			arrayDepth++
		}
	}
	//arrayDepth := strings.Count(typeIdentName, "[]")
	typeIdent := types.Identifier{
		Name:       typeIdentName,
		ArrayDepth: arrayDepth,
	}
	typeInfo := manager.get(typeIdent)
	if typeInfo == nil {
		panic(fmt.Sprintf("NewInternalStructField: Cannot find type info for %s on property %s", typeIdentName, name))
	}
	result := types.StructField{
		Name:           name,
		TypeIdentifier: typeIdent,
		TypeInfo:       typeInfo,
	}
	result.DefaultValue.TypeInfo = typeInfo
	return result
}

// Internal Structs
func (manager *TypeInfoManager) InternalWorkspaceStruct() *types.Struct {
	return manager.workspaceInfo
}

// Functions
func (p *Typer) DetermineType(node *ast.TypeIdent) types.TypeInfo {
	return p.typeinfo.get(types.Identifier{
		Name:       node.Name.String(),
		ArrayDepth: node.ArrayDepth,
	})
}

func TypeEquals(a types.TypeInfo, b types.TypeInfo) bool {
	aAsArray, aOk := a.(*types.Array)
	bAsArray, bOk := b.(*types.Array)
	if aOk && bOk {
		return TypeEquals(aAsArray.Underlying(), bAsArray.Underlying())
	}
	if a == b {
		return true
	}
	return false
}
