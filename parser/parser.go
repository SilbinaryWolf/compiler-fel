package parser

import (
	"fmt"
	"io/ioutil"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

type Parser struct {
	*scanner.Scanner
}

func ParseFile(filepath string) (*ast.File, error) {
	filecontentsAsBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return Parse(filecontentsAsBytes, filepath)
}

func Parse(filecontentsAsBytes []byte, filepath string) (*ast.File, error) {
	p := new(Parser)
	p.Scanner = scanner.New(filecontentsAsBytes, filepath)

	resultNode := &ast.File{
		Filepath: filepath,
	}

	//Loop:
	for {
		t := p.GetNextToken()
		switch t.Kind {
		case token.Identifier:
			ident := t.String()
			switch ident {
			case "config":
				t := p.GetNextToken()
				if t.Kind != token.BraceOpen {
					return nil, p.expect(t, token.BraceOpen)
				}
				p.parseBlock()
				panic("todo: Finish Parse() func")
			default:
				return nil, p.expect(t, "config")
			}
		default:
			panic(fmt.Sprintf("Parse(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}
	return resultNode, nil
}

func (p *Parser) expect(thisToken token.Token, expectedList ...interface{}) error {
	expectedItemsString := ""
	lengthMinusOne := len(expectedList) - 1
	for i, expectedItem := range expectedList {
		switch value := expectedItem.(type) {
		case token.Kind:
			switch value {
			case token.Identifier:
				expectedItemsString += "identifier"
			case token.InteropVariable:
				expectedItemsString += "interop variable"
			default:
				panic(fmt.Sprintf("Unhandled token kind: %s", value.String()))
			}
		case string:
			expectedItemsString += fmt.Sprintf("keyword \"%s\"", value)
		default:
			panic("unhandled type")
		}
		if i != 0 {
			if i < lengthMinusOne {
				expectedItemsString += ", "
			} else {
				expectedItemsString += " or "
			}
		}
		/*switch expectTokenKind {
		case token.Identifier:
			expectedItemsString += "identifier"
		case token.InteropVariable:
			expectedItemsString += "interop variable"
		default:
			panic(fmt.Sprintf("Unhandled token kind: %s", expectTokenKind.String()))
		}*/
	}

	return fmt.Errorf("Line %d, Expected %s instead got \"%s\".", p.GetLine(), expectedItemsString, thisToken.String())
}

/*
func (p *Parser) expectKeywords(thisToken token.Token, expectedKeywordList ...string) error {
	expectedItemsString := ""
	for i, expectTokenKind := range expectedTokenKindList {
		if i != 0 {
			if i < len(expectedTokenList)-1 {
				expectedItemsString += ", "
			} else {
				expectedItemsString += " or "
			}
		}
		switch expectTokenKind {
		case token.Identifier:
			expectedItemsString += "identifier"
		case token.InteropVariable:
			expectedItemsString += "interop variable"
		default:
			panic(fmt.Sprintf("Unhandled token kind: %s", expectTokenKind.Kind.String()))
			// no-op
		}
	}

	return fmt.Errorf("Line %d, Expected %s instead got \"%s\".", p.GetLine(), expectedItemsString, thisToken.String())
}
*/
