package parser

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

type Scope struct {
	identifiers map[string]data.Kind

	cssDefinitions       map[string]*ast.CSSDefinition
	cssConfigDefinitions map[string]*ast.CSSConfigDefinition
	htmlDefinitions      map[string]*ast.HTMLComponentDefinition

	parent *Scope
}

func NewScope(parent *Scope) *Scope {
	result := new(Scope)

	result.identifiers = make(map[string]data.Kind)
	result.cssDefinitions = make(map[string]*ast.CSSDefinition)
	result.cssConfigDefinitions = make(map[string]*ast.CSSConfigDefinition)
	result.htmlDefinitions = make(map[string]*ast.HTMLComponentDefinition)

	result.parent = parent
	return result
}

func (scope *Scope) Set(name string, value data.Kind) {
	scope.identifiers[name] = value
}

func (scope *Scope) Get(name string) (data.Kind, bool) {
	value, ok := scope.identifiers[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.Get(name)
	}
	return value, ok
}

func (scope *Scope) GetHTMLDefinition(name string) (*ast.HTMLComponentDefinition, bool) {
	value, ok := scope.htmlDefinitions[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.GetHTMLDefinition(name)
	}
	return value, ok
}

func (scope *Scope) GetCSSDefinition(name string) (*ast.CSSDefinition, bool) {
	value, ok := scope.cssDefinitions[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.GetCSSDefinition(name)
	}
	return value, ok
}

func (scope *Scope) GetFromThisScope(name string) (data.Kind, bool) {
	value, ok := scope.identifiers[name]
	return value, ok
}
