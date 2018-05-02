package types

import (
	"github.com/silbinarywolf/compiler-fel/ast"
)

type TypeInfo interface {
	String() string
	ImplementsTypeInfo()
}

//
// Identifier
//

type Identifier struct {
	Name       string
	ArrayDepth int
}

//
// Int
//

type Int struct{}

func (_ *Int) String() string      { return "int" }
func (_ *Int) ImplementsTypeInfo() {}

//
// Float
//

type Float struct{}

func (_ *Float) String() string      { return "float" }
func (_ *Float) ImplementsTypeInfo() {}

//
// String
//

type String struct{}

func (_ *String) String() string      { return "string" }
func (_ *String) ImplementsTypeInfo() {}

//
// Array
//

type Array struct {
	underlying TypeInfo
}

func NewArray(underlying TypeInfo) *Array {
	info := new(Array)
	info.underlying = underlying
	return info
}

func (info *Array) String() string       { return "[]" + info.underlying.String() }
func (info *Array) Underlying() TypeInfo { return info.underlying }
func (_ *Array) ImplementsTypeInfo()     {}

//
// Procedure
//
type Procedure struct {
	name       string
	definition *ast.ProcedureDefinition
}

func (info *Procedure) String() string                       { return "procedure " + info.name + "()" }
func (info *Procedure) Name() string                         { return info.name }
func (info *Procedure) Definition() *ast.ProcedureDefinition { return info.definition }
func (_ *Procedure) ImplementsTypeInfo()                     {}

func NewProcedure(definiton *ast.ProcedureDefinition) *Procedure {
	result := new(Procedure)
	result.name = definiton.Name.String()
	result.definition = definiton
	return result
}

//
// Bool
//
type Bool struct{}

func (_ *Bool) String() string      { return "bool" }
func (_ *Bool) ImplementsTypeInfo() {}

//
// HTML Node
//
type HTMLNode struct{}

func (_ *HTMLNode) String() string      { return "html node" }
func (_ *HTMLNode) ImplementsTypeInfo() {}

//
// Struct
//
type Struct struct {
	name   string
	fields []StructField
	//definition *ast.StructDefinition
}

func (info *Struct) String() string        { return info.name }
func (info *Struct) Name() string          { return info.name }
func (info *Struct) Fields() []StructField { return info.fields }
func (_ *Struct) ImplementsTypeInfo()      {}

//func (info *Struct) Definition() *ast.StructDefinition { return info.definition }

func NewStruct(structDef *ast.StructDefinition) *Struct {
	name := structDef.Name().String()
	fields := structDef.Fields()

	result := new(Struct)
	result.name = name
	result.fields = make([]StructField, 0, len(fields))
	for i, field := range fields {
		result.fields = append(result.fields, StructField{
			index: i,
			Name:  field.Name().String(),
			TypeIdentifier: Identifier{
				Name:       field.TypeIdentifier.Name.String(),
				ArrayDepth: field.TypeIdentifier.ArrayDepth,
			},
			TypeInfo:     field.TypeInfo,
			DefaultValue: field.Expression,
		})
	}
	return result
}

func NewInternalStruct(name string, fields []StructField) *Struct {
	for i := range fields {
		field := &fields[i]
		field.index = i
	}
	result := new(Struct)
	result.name = name
	result.fields = fields
	return result
}

func (info *Struct) GetFieldByName(name string) *StructField {
	fields := info.Fields()
	for i := 0; i < len(fields); i++ {
		field := &fields[i]
		if field.Name == name {
			return field
		}
	}
	return nil
}

type StructField struct {
	Name           string
	index          int
	TypeIdentifier Identifier
	TypeInfo       TypeInfo
	DefaultValue   ast.Expression
}

func (field *StructField) Index() int { return field.index }
