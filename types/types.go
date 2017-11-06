package types

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
)

var registeredTypes = make(map[string]TypeInfo)

func Equal(a TypeInfo, b TypeInfo) bool {
	if a == b {
		return true
	}
	return false
}

// Built-ins
var typeString = new(String_)
var typeInt = new(int_)
var typeFloat = new(float_)
var typeBool = new(bool_)
var typeHTML = new(html)

type TypeInfo interface {
	Name() string
	Create() data.Type
}

func HasNoType(a TypeInfo) bool {
	return a == nil
}

func Equals(a TypeInfo, b TypeInfo) bool {
	if a == b {
		return true
	}
	return false
}

// Int
type int_ struct{}

func (info *int_) Name() string      { return "int" }
func (info *int_) Create() data.Type { return new(data.Integer64) }
func Int() *int_                     { return typeInt }

// Float
type float_ struct{}

func (info *float_) Name() string      { return "float" }
func (info *float_) Create() data.Type { return new(data.Float64) }
func Float() *float_                   { return typeFloat }

// Bool
type bool_ struct{}

func (info *bool_) Name() string      { return "bool" }
func (info *bool_) Create() data.Type { return new(data.Bool) }
func Bool() *bool_                    { return typeBool }

// String
type String_ struct{}

func (info *String_) Name() string      { return "string" }
func (info *String_) Create() data.Type { return new(data.String) }
func String() *String_                  { return typeString }

// HTML
type html struct{}

func (info *html) Name() string      { return "html node" }
func (info *html) Create() data.Type { return new(data.HTMLNode) }
func HTML() *html                    { return typeHTML }

//
// Parser
//
func GetTypeFromString(name string) TypeInfo {
	return registeredTypes[name]
}

//
// Initialization
//
func RegisterType(name string, info TypeInfo) {
	_, ok := registeredTypes[name]
	if ok {
		panic(fmt.Sprintf("Already registered \"%s\" type."))
	}
	registeredTypes[name] = info
}

func init() {
	RegisterType("string", typeString)
	RegisterType("int", typeInt)
	RegisterType("float", typeFloat)
	RegisterType("bool", typeBool)

	RegisterType("html", typeHTML)
}
