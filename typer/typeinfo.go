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
	boolInfo     types.Bool
	htmlNodeInfo types.HTMLNode

	// built-in structs
	workspaceInfo *types.Struct
}

func (manager *TypeInfoManager) Init() {
	if manager.registeredTypes != nil {
		panic("Cannot initialize TypeInfoManager twice.")
	}
	manager.registeredTypes = make(map[types.Identifier]types.TypeInfo)

	// Primitives
	manager.register("bool", manager.NewTypeInfoBool())
	manager.register("int", manager.NewTypeInfoInt())
	manager.register("string", manager.NewTypeInfoString())
	manager.register("float", manager.NewTypeInfoFloat())

	// Internal types
	manager.workspaceInfo = types.NewInternalStruct(
		"Workspace",
		[]types.StructField{
			// NOTE(Jake): 2018-04-10
			//
			// Using a space with " name" as a hack so that it can't
			// be accessed by user-code, but can be read/understood by
			// internal tools. (the workspace name)
			//
			manager.NewInternalStructField(" name", "string"),
			manager.NewInternalStructField("template_input_directory", "string"),
			manager.NewInternalStructField("template_output_directory", "string"),
			manager.NewInternalStructField("css_output_directory", "string"),
			manager.NewInternalStructField("css_files", "[]string"),
		},
	)
}

/*func (manager *TypeInfoManager) Clear() {
	if manager.registeredTypes == nil {
		panic("Cannot clear TypeInfoManager if it's already cleared..")
	}
	manager.registeredTypes = nil
}*/

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

func (manager *TypeInfoManager) NewTypeInfoBool() *types.Bool     { return &manager.boolInfo }
func (manager *TypeInfoManager) NewTypeInfoInt() *types.Int       { return &manager.intInfo }
func (manager *TypeInfoManager) NewTypeInfoFloat() *types.Float   { return &manager.floatInfo }
func (manager *TypeInfoManager) NewTypeInfoString() *types.String { return &manager.stringInfo }

// Internal Struct Types
func (manager *TypeInfoManager) InternalWorkspaceStruct() *types.Struct { return manager.workspaceInfo }

func (_ *TypeInfoManager) NewTypeInfoArray(underlying types.TypeInfo) *types.Array {
	return types.NewArray(underlying)
}

func (_ *TypeInfoManager) NewProcedureInfo(definiton *ast.ProcedureDefinition) *types.Procedure {
	return types.NewProcedure(definiton)
}

// HTMLNode
func (manager *TypeInfoManager) NewHTMLNode() *types.HTMLNode {
	return &manager.htmlNodeInfo
}

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
