package evaluator

import "github.com/silbinarywolf/compiler-fel/data"

type Scope struct {
	variables map[string]data.Type
	parent    *Scope
}

func NewScope(parent *Scope) *Scope {
	result := new(Scope)
	result.variables = make(map[string]data.Type)
	result.parent = parent
	return result
}

func (scope *Scope) Set(name string, value data.Type) {
	scope.variables[name] = value
}

func (scope *Scope) Get(name string) (data.Type, bool) {
	value, ok := scope.variables[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.Get(name)
	}
	return value, ok
}

func (scope *Scope) GetThisScope(name string) (data.Type, bool) {
	value, ok := scope.variables[name]
	return value, ok
}
