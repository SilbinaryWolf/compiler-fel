package parser

import (
	"fmt"
	"io/ioutil"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/scanner"
	"github.com/silbinarywolf/compiler-fel/token"
)

type GlobalDefinitions struct{}

type Parser struct {
	*scanner.Scanner
	errors                        map[string][]error
	typecheckHtmlNodeDependencies map[string]*ast.HTMLNode
}

func New() *Parser {
	p := new(Parser)
	p.errors = make(map[string][]error)
	//p.typecheckHtmlDefinitionDependencies = make(map[string]*ast.HTMLComponentDefinition)
	//p.typecheckHtmlDefinitionStack = make([]*ast.HTMLComponentDefinition, 0, 20)
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

	line := p.Line()
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

//func (p *Parser) GetErrors() []error {
//	return p.errors
//}

func (p *Parser) HasErrors() bool {
	return p.Scanner.Error != nil || len(p.errors) > 0
}

func (p *Parser) addError(message error) {
	// todo(Jake): Expose this function to AST/token/etc data to retrieve line number
	//message = fmt.Errorf("Line %d - %s", -99, message)
	filepath := p.Filepath
	_, ok := p.errors[filepath]
	if !ok {
		p.errors[filepath] = make([]error, 0, 10)
	}
	p.errors[filepath] = append(p.errors[filepath], message)
}

func (p *Parser) addErrorToken(message error, token token.Token) {
	filepath := token.Filepath
	_, ok := p.errors[filepath]
	if !ok {
		p.errors[filepath] = make([]error, 0, 10)
	}
	message = fmt.Errorf("Line %d | %s", token.Line, message)
	p.errors[filepath] = append(p.errors[filepath], message)
}

func (p *Parser) PrintErrors() {
	errorCount := 0
	for _, errorList := range p.errors {
		errorCount += len(errorList)
	}
	if p.Scanner.Error != nil {
		errorCount += 1
	}
	if errorCount > 0 {
		errorOrErrors := "errors"
		if errorCount == 1 {
			errorOrErrors = "error"
		}
		fmt.Printf("Found %d %s...\n", errorCount, errorOrErrors)
		if p.Scanner.Error != nil {
			fmt.Printf("File: %s\n", p.Scanner.Filepath)
			fmt.Printf("- %s \n", p.Scanner.Error)
		}

		isFirst := true
		for filepath, errorList := range p.errors {
			fmt.Printf("File: %s\n", filepath)
			for _, err := range errorList {
				fmt.Printf("- %v \n", err)
			}
			if !isFirst {
				fmt.Printf("\n")
			}
			isFirst = false
		}
		fmt.Printf("\n")
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
