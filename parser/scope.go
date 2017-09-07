package parser

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

type Scope struct {
	variables map[string]data.Kind

	cssDefinitions  map[string]*ast.CSSDefinition
	htmlDefinitions map[string]*ast.HTMLComponentDefinition

	parent *Scope
}

func NewScope(parent *Scope) *Scope {
	result := new(Scope)

	result.variables = make(map[string]data.Kind)
	result.cssDefinitions = make(map[string]*ast.CSSDefinition)
	result.htmlDefinitions = make(map[string]*ast.HTMLComponentDefinition)

	result.parent = parent
	return result
}

func (scope *Scope) Set(name string, value data.Kind) {
	scope.variables[name] = value
}

func (scope *Scope) Get(name string) (data.Kind, bool) {
	value, ok := scope.variables[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.Get(name)
	}
	return value, ok
}

func (scope *Scope) GetFromThisScope(name string) (data.Kind, bool) {
	value, ok := scope.variables[name]
	return value, ok
}
