package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/util"
)

func (p *Parser) checkHTMLNode(node *ast.HTMLNode) {
	name := node.Name.String()
	if len(node.ChildNodes) > 0 && util.IsSelfClosingTagName(name) {
		p.addErrorToken(fmt.Errorf("%s is a self-closing tag and cannot have child elements.", name), node.Name)
	}

	//
	// todo(Jake): Extend this to allow user configured/whitelisted tag names
	//
	//isValidHTML5TagName := util.IsValidHTML5TagName(name)
	//if !isValidHTML5TagName {
	//p.htmlComponentNodes = append(p.htmlComponentNodes, node)
	//p.addErrorLine(fmt.Errorf("\"%s\" is not a valid HTML5 element.", name), node.Name.Line)
	//}
}
