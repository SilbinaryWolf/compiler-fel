package typer

import (
	"bytes"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/types"
)

/*type SymbolKind int

const (
	SymbolUnknown SymbolKind = 0 + iota
	SymbolVariable
	SymbolHTMLComponent
	SymbolStructDefinition
)*/

type Scope struct {
	identifiers map[string]*Symbol
	parent      *Scope
}

type Symbol struct {
	name string

	// For variables
	variable types.TypeInfo

	// For combined symbol (Component-pieces)
	cssDefinition       *ast.CSSDefinition
	cssConfigDefinition *ast.CSSConfigDefinition
	htmlDefinition      *ast.HTMLComponentDefinition
	structDefinition    *ast.StructDefinition
}

func (symbol *Symbol) GetType() string {
	if symbol.htmlDefinition != nil {
		return fmt.Sprintf("%s :: html", symbol.htmlDefinition.Name.String())
	}
	if symbol.cssDefinition != nil {
		return fmt.Sprintf("%s :: css", symbol.cssDefinition.Name.String())
	}
	if symbol.variable != nil {
		switch variable := symbol.variable.(type) {
		case *types.Procedure:
			return fmt.Sprintf("%s :: ()", variable.Name())
		default:
			return fmt.Sprintf("unknown variable type: %T", variable)
		}
	}
	return "<error calling symbol.GetType()>"
}

/*func (symbol *Symbol) expected(kind SymbolKind) error {
	name := symbol.name
	switch kind {
	case SymbolHTMLComponent:
		return fmt.Errorf("Expected identifier \"%s\" to be a HTML component.", name)
	case SymbolStructDefinition:
		return fmt.Errorf("Expected identifier \"%s\" to be a struct.", name)
	}
	return fmt.Errorf("Expected identifier \"%s\" to be a <a different type>.", name)
}*/

func NewScope(parent *Scope) *Scope {
	result := new(Scope)
	result.identifiers = make(map[string]*Symbol)
	result.parent = parent
	return result
}

func (scope *Scope) getOrCreateSymbol(name string) *Symbol {
	symbol := scope.identifiers[name]
	if symbol == nil {
		symbol = new(Symbol)
		scope.identifiers[name] = symbol
	}
	return symbol
}

func (scope *Scope) GetSymbols(name string) *Symbol {
	symbol := scope.identifiers[name]
	if symbol == nil && scope.parent != nil {
		return scope.parent.GetSymbol(name)
	}
	return symbol
}

func (scope *Scope) GetSymbol(name string) *Symbol {
	symbol := scope.identifiers[name]
	if symbol == nil && scope.parent != nil {
		return scope.parent.GetSymbol(name)
	}
	return symbol
}

func (scope *Scope) GetSymbolFromThisScope(name string) *Symbol {
	symbol := scope.identifiers[name]
	return symbol
}

//

func (scope *Scope) SetVariable(name string, typeinfo types.TypeInfo) {
	symbol := scope.getOrCreateSymbol(name)
	if symbol.variable != nil {
		panic(fmt.Sprintf("Cannot set \"variable\" of symbol \"%s\" more than once.", name))
	}
	symbol.variable = typeinfo
}

/*func (scope *Scope) GetVariable(name string) (types.TypeInfo, bool) {
	symbol := scope.identifiers[name]
	if symbol == nil && scope.parent != nil {
		return scope.parent.GetVariable(name)
	}
	if symbol == nil {
		return nil, false
	}
	return symbol.variable, true
}

func (scope *Scope) GetVariableThisScope(name string) (types.TypeInfo, bool) {
	symbol := scope.identifiers[name]
	if symbol == nil {
		return nil, false
	}
	return symbol.variable, true
}

func (scope *Scope) SetVariable(name string, typeinfo types.TypeInfo) {
	symbol := scope.getOrCreateSymbol(name)
	if symbol.variable != nil {
		panic(fmt.Sprintf("Cannot set \"variable\" of symbol \"%s\" more than once.", name))
	}
	symbol.variable = typeinfo
}

//

func (scope *Scope) GetHTMLDefinition(name string) *ast.HTMLComponentDefinition {
	symbol := scope.identifiers[name]
	if symbol == nil && scope.parent != nil {
		return scope.parent.GetHTMLDefinition(name)
	}
	if symbol == nil {
		return nil
	}
	return symbol.htmlDefinition
}

func (scope *Scope) GetHTMLDefinitionFromThisScope(name string) *ast.HTMLComponentDefinition {
	symbol := scope.identifiers[name]
	if symbol == nil {
		return nil
	}
	return symbol.htmlDefinition
}

func (scope *Scope) SetHTMLDefinition(name string, definition *ast.HTMLComponentDefinition) {
	symbol := scope.getOrCreateSymbol(name)
	if symbol.htmlDefinition != nil {
		panic(fmt.Sprintf("Cannot set \"HTML definition\" of symbol \"%s\" more than once.", name))
	}
	symbol.htmlDefinition = definition
}

//

func (scope *Scope) GetStructDefinition(name string) (*ast.StructDefinition, bool) {
	symbol := scope.identifiers[name]
	if symbol == nil && scope.parent != nil {
		return scope.parent.GetStructDefinition(name)
	}
	if symbol == nil {
		return nil, false
	}
	return symbol.structDefinition, true
}

func (scope *Scope) GetStructDefinitionFromThisScope(name string) *ast.StructDefinition {
	symbol := scope.identifiers[name]
	if symbol == nil {
		return nil
	}
	return symbol.structDefinition
}

//

func (scope *Scope) GetCSSDefinition(name string) (*ast.CSSDefinition, bool) {
	symbol := scope.identifiers[name]
	if symbol == nil && scope.parent != nil {
		return scope.parent.GetCSSDefinition(name)
	}
	if symbol == nil {
		return nil, false
	}
	return symbol.cssDefinition, true
}

func (scope *Scope) GetCSSDefinitionFromThisScope(name string) *ast.CSSDefinition {
	symbol := scope.identifiers[name]
	if symbol == nil {
		return nil
	}
	return symbol.cssDefinition
}

func (scope *Scope) SetCSSDefinition(name string, definition *ast.CSSDefinition) {
	symbol := scope.getOrCreateSymbol(name)
	if symbol.cssDefinition != nil {
		panic(fmt.Sprintf("Cannot set \"css definition\" of symbol \"%s\" more than once.", name))
	}
	symbol.cssDefinition = definition
}

//

func (scope *Scope) getCSSConfigDefinition(name string) (*ast.CSSConfigDefinition, bool) {
	symbol := scope.identifiers[name]
	if symbol == nil && scope.parent != nil {
		return scope.parent.getCSSConfigDefinition(name)
	}
	if symbol == nil {
		return nil, false
	}
	return symbol.cssConfigDefinition, true
}

func (scope *Scope) getCSSConfigDefinitionFromThisScope(name string) *ast.CSSConfigDefinition {
	symbol := scope.identifiers[name]
	if symbol == nil {
		return nil
	}
	return symbol.cssConfigDefinition
}

func (scope *Scope) SetCSSConfigDefinition(name string, definition *ast.CSSConfigDefinition) {
	symbol := scope.getOrCreateSymbol(name)
	if symbol.cssDefinition != nil {
		panic(fmt.Sprintf("Cannot set \"css config definition\" of symbol \"%s\" more than once.", name))
	}
	symbol.cssConfigDefinition = definition
}*/

func (scope *Scope) debug() string {
	var buffer bytes.Buffer
	if scope.parent != nil {
		buffer.WriteString(scope.parent.debug())
	}
	buffer.WriteString("\nScope:\n")
	if len(scope.identifiers) > 0 {
		for name, _ := range scope.identifiers {
			buffer.WriteString(fmt.Sprintf("- %s\n", name))
		}
	} else {
		buffer.WriteString(fmt.Sprintf("<no symbols in this scope>\n"))
	}
	return buffer.String()
}

/*
func (scope *Scope) GetCSSDefinition(name string) (*ast.CSSDefinition, bool) {
	value, ok := scope.cssDefinitions[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.GetCSSDefinition(name)
	}
	return value, ok
}

func (scope *Scope) GetCSSConfigDefinition(name string) (*ast.CSSConfigDefinition, bool) {
	value, ok := scope.cssConfigDefinitions[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.GetCSSConfigDefinition(name)
	}
	return value, ok
}*/
