package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/token"
)

func (p *Parser) parseDefinition(name token.Token) ast.Node {
	keywordToken := p.GetNextToken()
	keyword := keywordToken.String()
	switch keyword {
	case "css":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		node := p.parseCSS(name)
		return node
	case "css_config":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		node := p.parseCSSConfigRuleDefinition(name)
		return node
	case "struct":
		if t := p.GetNextToken(); t.Kind != token.BraceOpen {
			p.addError(p.expect(t, token.BraceOpen))
			return nil
		}
		//
		//
		childNodes := p.parseStatements()
		fields := make([]ast.StructField, 0, len(childNodes))
		// NOTE(Jake): A bit of a hack, we should have a 'parseStruct' function
		for _, itNode := range childNodes {
			switch node := itNode.(type) {
			case *ast.DeclareStatement:
				field := ast.StructField{}
				field.Name = node.Name
				//field.Expression.TypeIdentifier = node.Expression.TypeIdentifier
				field.Expression = node.Expression
				fields = append(fields, field)
			default:
				p.addErrorToken(fmt.Errorf("Expected statement, instead got %T.", itNode), name)
				return nil
			}
		}
		node := new(ast.Struct)
		node.Name = name
		node.Fields = fields
		return node
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
			// NOTE: No longer applicable.
			//if htmlNodeCount > 1 {
			//	p.addErrorToken(fmt.Errorf("\"%s:: html\" cannot have multiple HTML nodes at the top-level.", nameString), name)
			//}
		}

		if name.Kind != token.Unknown {
			// Retrieve properties block
			var cssDef *ast.CSSDefinition
			var structure *ast.Struct
		RetrievePropertyDefinitionLoop:
			for _, itNode := range childNodes {
				switch node := itNode.(type) {
				case *ast.Struct:
					if structure != nil {
						p.addError(fmt.Errorf("Cannot declare \":: struct\" twice in the same HTML component."))
						break RetrievePropertyDefinitionLoop
					}
					structure = node
				case *ast.CSSDefinition:
					if cssDef != nil {
						p.addError(fmt.Errorf("Cannot declare \":: css\" twice in the same HTML component."))
						break RetrievePropertyDefinitionLoop
					}
					cssDef = node
				}
			}

			// Component
			node := new(ast.HTMLComponentDefinition)
			node.Name = name
			node.Properties = structure
			node.CSSDefinition = cssDef
			node.ChildNodes = childNodes

			return node
		}

		// TODO(Jake): Disallow ":: properties" block in un-named HTML
		node := new(ast.HTMLBlock)
		node.HTMLKeyword = keywordToken
		node.ChildNodes = childNodes
		return node
	}
	p.addError(fmt.Errorf("Unexpected keyword '%s' for definition (::) type. Expected 'css', 'html' or 'properties' on Line %d", keyword, keywordToken.Line))
	return nil
}
