package parser

import (
	//"encoding/json"

	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/util"
)

func (p *Parser) checkHTMLNode(node *ast.HTMLNode) {
	name := node.Name.String()
	if len(node.ChildNodes) > 0 && util.IsSelfClosingTagName(name) {
		p.addErrorLine(fmt.Errorf("%s is a self-closing tag and cannot have child elements.", name), node.Name.Line)
	}

	if !util.IsValidHTML5TagName(name) {
		p.addErrorLine(fmt.Errorf("\"%s\" is not a valid HTML5 element.", name), node.Name.Line)
	}
}
