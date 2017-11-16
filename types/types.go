package types

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
)

type TypeInfo interface {
	String() string
	Create() data.Type
}

func HasNoType(a TypeInfo) bool {
	return a == nil
}

func Equals(a TypeInfo, b TypeInfo) bool {
	aAsArray, aOk := a.(*Array_)
	bAsArray, bOk := b.(*Array_)
	if aOk && bOk {
		return Equals(aAsArray.underlying, bAsArray.underlying)
	}
	if a == b {
		return true
	}
	return false
}

// Float
type Float_ struct{}

func (info *Float_) String() string { return "float" }

func (info *Float_) Create() data.Type { return new(data.Float64) }

var typeFloat = new(Float_)

func Float() TypeInfo { return typeFloat }

// Bool
type Bool_ struct{}

func (info *Bool_) String() string { return "bool" }

func (info *Bool_) Create() data.Type { return new(data.Bool) }

var typeBool = new(Bool_)

func Bool() TypeInfo { return typeBool }

// String
type String_ struct{}

func (info *String_) String() string { return "string" }

func (info *String_) Create() data.Type { return new(data.String) }

var typeString = new(String_)

func String() TypeInfo { return typeString }

// HTML
type HTMLNode_ struct{}

func (info *HTMLNode_) String() string { return "html node" }

func (info *HTMLNode_) Create() data.Type { return new(data.HTMLNode) }

var typeHTMLNode_ = new(HTMLNode_)

func HTMLNode() TypeInfo { return typeHTMLNode_ }

// Array
type Array_ struct {
	underlying TypeInfo
}

func (info *Array_) String() string       { return "[]" + info.underlying.String() }
func (info *Array_) Underlying() TypeInfo { return info.underlying }
func (info *Array_) Create() data.Type    { return data.NewArray(info.underlying.Create()) }

func Array(underlying TypeInfo) TypeInfo {
	result := new(Array_)
	result.underlying = underlying
	return result
}

//
// Initialization
//
var registeredTypes = make(map[string]TypeInfo)

func RegisterType(name string, info TypeInfo) {
	_, ok := registeredTypes[name]
	if ok {
		panic(fmt.Sprintf("Already registered \"%s\" type.", name))
	}
	registeredTypes[name] = info
}

func GetRegisteredType(name string) TypeInfo {
	return registeredTypes[name]
}

func init() {
	RegisterType("string", String())
	//RegisterType("int", Int())
	RegisterType("float", Float())
	RegisterType("bool", Bool())
}
