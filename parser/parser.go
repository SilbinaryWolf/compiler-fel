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

func (p *Parser) ParseFile(filepath string) (*ast.File, error) {
	filecontentsAsBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	result := p.Parse(filecontentsAsBytes, filepath)
	return result, nil
}

func (p *Parser) Parse(filecontentsAsBytes []byte, filepath string) *ast.File {
	p.Scanner = scanner.New(filecontentsAsBytes, filepath)
	resultNode := &ast.File{
		Filepath: filepath,
	}
	resultNode.ChildNodes = p.parseStatements()
	return resultNode
}

func (p *Parser) expect(thisToken token.Token, expectedList ...interface{}) error {

	// todo(Jake): switch to using a buffer as that uses less allocations
	//			   ie. increase speed from 6500ns to 15ns
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
				expectedItemsString += value.String()
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

	line := p.GetLine()
	if thisToken.Kind == token.Newline {
		// Reading the newline token will offset to the next line causing a mistake in the
		// error message
		line--
	}

	// NOTE(Jake): Line 1, Expected { instead got "newline"
	result := fmt.Errorf("Expected %s instead got \"%s\".", line, expectedItemsString, thisToken.String())

	// For debugging
	panic(result)

	return result
}

func (p *Parser) GetErrors() []error {
	return p.errors
}

func (p *Parser) HasErrors() bool {
	return p.Scanner.Error != nil || len(p.errors) > 0
}

func (p *Parser) addError(message error) {
	// todo(Jake): Expose this function to AST/token/etc data to retrieve line number
	//message = fmt.Errorf("Line %d - %s", -99, message)
	p.errors = append(p.errors, message)
}

func (p *Parser) addErrorLine(message error, line int) {
	message = fmt.Errorf("Line %d | %s", line, message)
	p.errors = append(p.errors, message)
}

func (p *Parser) PrintErrors() {
	errors := p.GetErrors()
	errorCount := len(errors)
	if p.Scanner.Error != nil {
		errorCount += 1
	}
	if errorCount > 0 {
		errorOrErrors := "errors"
		if len(errors) == 1 {
			errorOrErrors = "error"
		}
		fmt.Printf("Error parsing file: %v\n", p.Filepath)
		fmt.Printf("Found %d %s in \"%s\"\n", errorCount, errorOrErrors, p.Filepath)
		for _, err := range errors {
			fmt.Printf("- %v \n\n", err)
		}
		if p.Scanner.Error != nil {
			fmt.Printf("- %v \n\n", p.Scanner.Error)
		}
	}
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
