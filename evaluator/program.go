package evaluator

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

type Program struct {
	Filepath              string
	globalScope           *Scope
	currentComponentScope []*ast.HTMLComponentDefinition

	htmlDefinitionUsed          map[string]HTMLComponentNodeInfo
	htmlTemplatesUsed           []HTMLComponentNodeInfo
	anonymousCSSDefinitionsUsed []*ast.CSSDefinition
	//debugLevel                  int
}

type HTMLComponentNodeInfo struct {
	HTMLDefinition *ast.HTMLComponentDefinition
	Nodes          []*data.HTMLComponentNode
}

func New() *Program {
	program := new(Program)
	program.globalScope = NewScope(nil)
	program.htmlDefinitionUsed = make(map[string]HTMLComponentNodeInfo)
	return program
}

func (program *Program) CurrentComponentScope() *ast.HTMLComponentDefinition {
	length := len(program.currentComponentScope)
	if length == 0 {
		return nil
	}
	return program.currentComponentScope[length-1]
}

func (program *Program) AddHTMLTemplateUsed(node *data.HTMLComponentNode) {
	var nodes []*data.HTMLComponentNode
	nodes = append(nodes, node)

	program.htmlTemplatesUsed = append(program.htmlTemplatesUsed, HTMLComponentNodeInfo{
		Nodes:          nodes,
		HTMLDefinition: nil,
	})
}

func (program *Program) AddHTMLDefinitionUsed(name string, htmlDefinition *ast.HTMLComponentDefinition, node *data.HTMLComponentNode) {
	nodeSet, ok := program.htmlDefinitionUsed[name]
	if !ok {
		nodeSet = HTMLComponentNodeInfo{}
		nodeSet.HTMLDefinition = htmlDefinition
		nodeSet.Nodes = make([]*data.HTMLComponentNode, 0, 5)
		program.htmlDefinitionUsed[name] = nodeSet
	}
	nodeSet.Nodes = append(nodeSet.Nodes, node)
}
