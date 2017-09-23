package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseDefinition(name token.Token) ast.Node {
	keywordToken := p.GetNextToken()
	switch keyword := keywordToken.String(); keyword {
	case "css":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		node := p.parseCSS(name)
		return node
	case "struct":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		childNodes := p.parseStatements()
		propertiesNode := new(ast.HTMLProperties)
		propertiesNode.Statements = make([]*ast.DeclareStatement, 0, len(childNodes))
		for _, itNode := range childNodes {
			switch node := itNode.(type) {
			case *ast.DeclareStatement:
				propertiesNode.Statements = append(propertiesNode.Statements, node)
			default:
				panic(fmt.Sprintf("parseDefinition(): Unhandled node type %T in :: properties block", node))
			}
		}
		return propertiesNode
	case "html":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		childNodes := p.parseStatements()

		// Check HTML nodes
		htmlNodeCount := 0
		for _, itNode := range childNodes {
			_, ok := itNode.(*ast.HTMLNode)
			if !ok {
				continue
			}
			htmlNodeCount++
		}
		if htmlNodeCount == 0 || htmlNodeCount > 1 {
			var nameString string
			if name.Kind != token.Unknown {
				nameString = name.String() + " "
			}
			if htmlNodeCount == 0 {
				p.addErrorToken(fmt.Errorf("\"%s:: html\" must contain one HTML node at the top-level.", nameString), name)
			}
			if htmlNodeCount > 1 {
				p.addErrorToken(fmt.Errorf("\"%s:: html\" cannot have multiple HTML nodes at the top-level.", nameString), name)
			}
		}

		if name.Kind != token.Unknown {
			// Retrieve properties block
			var cssDef *ast.CSSDefinition
			var properties *ast.HTMLProperties
		RetrievePropertyDefinitionLoop:
			for _, itNode := range childNodes {
				switch node := itNode.(type) {
				case *ast.HTMLProperties:
					if properties != nil {
						p.addError(fmt.Errorf("Cannot declare ':: properties' twice in the same HTML component."))
						break RetrievePropertyDefinitionLoop
					}
					properties = node
				case *ast.CSSDefinition:
					if cssDef != nil {
						p.addError(fmt.Errorf("Cannot declare ':: css' twice in the same HTML component."))
						break RetrievePropertyDefinitionLoop
					}
					cssDef = node
				}
			}

			// Component
			node := new(ast.HTMLComponentDefinition)
			node.Name = name
			node.Properties = properties
			node.CSSDefinition = cssDef
			node.ChildNodes = childNodes

			return node
		}

		// TODO(Jake): Disallow ":: properties" block in un-named HTML
		node := new(ast.HTMLBlock)
		node.HTMLKeyword = keywordToken
		node.ChildNodes = childNodes
		return node
	default:
		p.addError(fmt.Errorf("Unexpected keyword '%s' for definition (::) type. Expected 'css', 'html' or 'properties' on Line %d", keyword, keywordToken.Line))
		return nil
	}
	return nil
}
