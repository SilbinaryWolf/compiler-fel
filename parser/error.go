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

type ErrorHandler struct {
	errors map[string][]error
}

func (e *ErrorHandler) Init() {
	e.errors = make(map[string][]error)
}

type ParserError struct {
	Message error
	Token   token.Token
}

func (parserError *ParserError) Error() string {
	return parserError.Message.Error()
}

func unexpected(thisToken token.Token) error {
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

func expect(thisToken token.Token, expectedList ...interface{}) *ParserError {

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

	var message error
	if thisToken.Kind == token.String {
		message = fmt.Errorf("Expected %s instead got \"%s\".", expectedItemsString, thisToken.String())
	} else {
		message = fmt.Errorf("Expected %s instead got %s.", expectedItemsString, thisToken.String())
	}
	return &ParserError{
		Message: message,
		Token:   thisToken,
	}
}

//func (p *Parser) GetErrors() []error {
//	return p.errors
//}

func (e *ErrorHandler) HasErrors() bool {
	return len(e.errors) > 0
}

func (p *Parser) HasErrors() bool {
	return p.Scanner.Error != nil || p.ErrorHandler.HasErrors()
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

func (e *ErrorHandler) addErrorToken(message error, token token.Token) {
	filepath := token.Filepath
	_, ok := e.errors[filepath]
	if !ok {
		e.errors[filepath] = make([]error, 0, 10)
	}
	lineMessage := fmt.Sprintf("Line %d | ", token.Line)
	indentString := "\n  "
	for i := 0; i < len(lineMessage); i++ {
		indentString += " "
	}
	message = fmt.Errorf(lineMessage + strings.Replace(message.Error(), "\n", indentString, -1))

	// Get where the error message was added from to help
	// track where error messages are raised.
	if DEVELOPER_MODE {
		// NOTE(Jake): 2018-01-09
		//
		// Added code to get the last 3 layers of the call stack for improved
		// debugging in the typechecker.
		//
		callIndex := 1
		pc, file, line, ok := runtime.Caller(callIndex)
		details := runtime.FuncForPC(pc)
		for i := 0; i < 3; i++ {
			// Go up one call stack higher if one of these functions
			if name := details.Name(); strings.Contains(name, "fatalErrorToken") ||
				strings.Contains(name, "fatalError") {
				callIndex++
			}
			pc, file, line, ok = runtime.Caller(callIndex)

			if ok {
				// Reduce full filepath to just the scope of the `compiler-fel` repo
				fileParts := strings.Split(file, "/")
				if len(fileParts) >= 3 {
					file = fileParts[len(fileParts)-3] + "/" + fileParts[len(fileParts)-2] + "/" + fileParts[len(fileParts)-1]
				}
				message = fmt.Errorf("%s\n-- Line: %d | %s", message, line, file)
			}
			callIndex++
		}
	}
	e.errors[filepath] = append(e.errors[filepath], message)
}

func (p *Parser) fatalError(message error) {
	fmt.Printf("%s\n", message)
	p.PrintErrors()
	panic(FatalErrorMessage)
}

func (p *Parser) fatalErrorToken(message error, token token.Token) {
	p.addErrorToken(message, token)
	p.PrintErrors()
	panic(FatalErrorMessage)
}

// todo(Jake): 2018-01-14
//
// Decouple PrintErrors() to use ErrorHandler.
// Perhaps moving forward, the error handler can belong in
// it's own package and be applied to the "scanner.Scanner" class.
//
// This way it's functionality will be available across scanning tokens
// and parsing.
//
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
