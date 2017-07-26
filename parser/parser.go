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
	errors []error
}

func New() *Parser {
	p := new(Parser)
	return p
}

func (p *Parser) ParseFile(filepath string) *ast.File {
	filecontentsAsBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil
	}
	return p.Parse(filecontentsAsBytes, filepath)
}

func (p *Parser) Parse(filecontentsAsBytes []byte, filepath string) *ast.File {
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
					p.addError(p.expect(t, token.BraceOpen))
					return nil
				}
				p.parseBlock()
				panic("todo: Finish Parse() func")
			default:
				p.addError(p.expect(t, "config"))
				return nil
			}
		default:
			panic(fmt.Sprintf("Parse(): Unhandled token: %s on Line %d", t.Kind.String(), t.Line))
		}
	}
	return resultNode
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

func (p *Parser) HasError() bool {
	return len(p.errors) > 0
}

func (p *Parser) GetErrors() []error {
	return p.errors
}

func (p *Parser) addError(message error) {
	p.errors = append(p.errors, message)
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
