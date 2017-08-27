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
	case "properties":
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
		if name.Kind != token.Unknown {
			// Retrieve properties block
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
				}
			}

			// Component
			node := new(ast.HTMLComponentDefinition)
			node.Name = name
			node.Properties = properties
			node.ChildNodes = childNodes

			// Add component
			nameAsString := node.Name.String()
			_, ok := p.htmlComponentDefinitions[nameAsString]
			if ok {
				p.addError(fmt.Errorf("Cannot redeclare %s.", nameAsString))
			} else {
				p.htmlComponentDefinitions[nameAsString] = node
			}

			return node
		}

		// TODO(Jake): Disallow ":: properties" block in un-named HTML

		// HTML block
		node := new(ast.HTMLDefinition)
		node.ChildNodes = childNodes
		return node
	default:
		p.addError(fmt.Errorf("Unexpected keyword '%s' for definition (::) type. Expected 'css', 'html' or 'properties' on Line %d", keyword, keywordToken.Line))
		return nil
	}
	return nil
}
