package types

import (
	"github.com/silbinarywolf/compiler-fel/ast"
)

type TypeInfo interface {
	String() string
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

func (info *Int) String() string { return "int" }

//
// Float
//

type Float struct{}

func (info *Float) String() string { return "float" }

//
// String
//

type String struct{}

func (info *String) String() string { return "string" }

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

func (info *Array) String() string { return "[]" + info.underlying.String() }

func (info *Array) Underlying() TypeInfo { return info.underlying }

//
// Procedure
//
type Procedure struct {
	name       string
	definition *ast.ProcedureDefinition
}

func (info *Procedure) String() string                       { return "procedure " + info.name + "()" }
func (info *Procedure) Definition() *ast.ProcedureDefinition { return info.definition }

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

func (info *Bool) String() string { return "bool" }

//
// HTML Node
//
type HTMLNode struct{}

func (info *HTMLNode) String() string { return "html node" }

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

//func (info *Struct) Definition() *ast.StructDefinition { return info.definition }

func NewStruct(definiton *ast.StructDefinition) *Struct {
	result := new(Struct)
	result.name = definiton.Name.String()
	result.fields = make([]StructField, 0, len(definiton.Fields))
	for i, field := range definiton.Fields {
		result.fields = append(result.fields, StructField{
			index: i,
			Name:  field.Name.String(),
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
