package parser

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/silbinarywolf/compiler-fel/token"
)

// Adds additional info that tells the developer where a parser error was raised in Golang
const DEVELOPER_MODE = true

const FatalErrorMessage = "Fatal parsing error occurred. Please notify the developer(s)."

func (p *Parser) unexpected(thisToken token.Token) error {
	if thisToken.IsKeyword() {
		return fmt.Errorf("Unexpected keyword \"%s\".", thisToken.String())
	}
	if thisToken.IsOperator() {
		return fmt.Errorf("Unexpected operator \"%s\".", thisToken.String())
	}
	switch thisToken.Kind {
	case token.Identifier:
		return fmt.Errorf("Unexpected identifier \"%s\".", thisToken.String())
	case token.EOF:
		return fmt.Errorf("Unexpectedly reached end of file.")
	}
	return fmt.Errorf("Unexpected %s", thisToken.Kind)
}

func (p *Parser) expect(thisToken token.Token, expectedList ...interface{}) error {

	// todo(Jake): switch to using a buffer as that uses less allocations
	//			   ie. increase speed from 6500ns to 15ns
	expectedItemsString := ""
	for i, expectedItem := range expectedList {
		if i != 0 {
			if i == len(expectedList)-1 {
				expectedItemsString += " or "
			} else {
				expectedItemsString += ", "
			}
		}
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
		/*switch expectTokenKind {
		case token.Identifier:
			expectedItemsString += "identifier"
		case token.InteropVariable:
			expectedItemsString += "interop variable"
		default:
			panic(fmt.Sprintf("Unhandled token kind: %s", expectTokenKind.String()))
		}*/
	}
	return fmt.Errorf("Expected %s instead got \"%s\".", expectedItemsString, thisToken.String())
}

//func (p *Parser) GetErrors() []error {
//	return p.errors
//}

func (p *Parser) HasErrors() bool {
	return p.Scanner.Error != nil || len(p.errors) > 0
}

/*func (p *Parser) addError(message error) {
	// todo(Jake): Expose this function to AST/token/etc data to retrieve line number
	//message = fmt.Errorf("Line %d - %s", -99, message)
	filepath := p.Filepath
	_, ok := p.errors[filepath]
	if !ok {
		p.errors[filepath] = make([]error, 0, 10)
	}
	p.errors[filepath] = append(p.errors[filepath], message)
}*/

func (p *Parser) addErrorToken(message error, token token.Token) {
	filepath := token.Filepath
	_, ok := p.errors[filepath]
	if !ok {
		p.errors[filepath] = make([]error, 0, 10)
	}
	message = fmt.Errorf("Line %d | %s", token.Line, message)
	if DEVELOPER_MODE {
		// Get where the error message was added from to help
		// track where error messages are raised.

		pc, file, line, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)

		// Go up one call stack higher if one of these functions
		if name := details.Name(); strings.Contains(name, "fatalErrorToken") ||
			strings.Contains(name, "fatalError") {
			pc, file, line, ok = runtime.Caller(2)
		}

		if ok && details != nil {
			// Reduce full filepath to just the scope of the `compiler-fel` repo
			fileParts := strings.Split(file, "/")
			if len(fileParts) >= 3 {
				file = fileParts[len(fileParts)-3] + "/" + fileParts[len(fileParts)-2] + "/" + fileParts[len(fileParts)-1]
			}
			message = fmt.Errorf("%s\n-- Line: %d | %s", message, line, file)
		}
	}
	p.errors[filepath] = append(p.errors[filepath], message)
}

func (p *Parser) fatalError(message error) {
	//p.addError(message)
	p.PrintErrors()
	panic(FatalErrorMessage)
}

func (p *Parser) fatalErrorToken(message error, token token.Token) {
	p.addErrorToken(message, token)
	p.PrintErrors()
	panic(FatalErrorMessage)
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
