package evaluator

type Scope struct {
	variables map[string]DataType
	parent    *Scope
}

func (scope *Scope) Set(name string, value DataType) {
	scope.variables[name] = value
}

func (scope *Scope) GetAllScopes(name string) (DataType, bool) {
	value, ok := scope.variables[name]
	if !ok && scope.parent != nil {
		value, ok = scope.parent.GetCurrentScope(name)
	}
	return value, ok
}

func (scope *Scope) GetCurrentScope(name string) (DataType, bool) {
	value, ok := scope.variables[name]
	return value, ok
}
