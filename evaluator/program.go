package evaluator

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

type Program struct {
	Filepath                    string
	globalScope                 *Scope
	htmlDefinitionUsed          map[string]*HTMLNodeSet
	anonymousCSSDefinitionsUsed []*ast.CSSDefinition
	//debugLevel                  int
}

type HTMLNodeSet struct {
	HTMLDefinition *ast.HTMLComponentDefinition
	items          []*data.HTMLComponentNode
}

func New() *Program {
	program := new(Program)
	program.globalScope = NewScope(nil)
	program.htmlDefinitionUsed = make(map[string]*HTMLNodeSet)
	return program
}

func (program *Program) AddHTMLDefinitionUsed(name string, htmlDefinition *ast.HTMLComponentDefinition, node *data.HTMLComponentNode) {
	nodeSet, ok := program.htmlDefinitionUsed[name]
	if !ok {
		nodeSet = new(HTMLNodeSet)
		nodeSet.HTMLDefinition = htmlDefinition
		nodeSet.items = make([]*data.HTMLComponentNode, 0, 5)
		program.htmlDefinitionUsed[name] = nodeSet
	}
	nodeSet.items = append(nodeSet.items, node)
}
