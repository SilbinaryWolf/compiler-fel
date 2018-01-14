package errors

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/silbinarywolf/compiler-fel/token"
)

const fatalErrorMessage = "Fatal parsing error occurred. Please notify the developer(s)."

type ErrorHandler struct {
	errors  map[string][]error
	devMode bool
}

func (e *ErrorHandler) Init() {
	if e.errors != nil {
		panic("Cannot initialize error handler more than once.")
	}
	e.errors = make(map[string][]error)
	e.devMode = false
}

func (e *ErrorHandler) SetDeveloperMode(value bool) {
	e.devMode = value
}

type ParserError struct {
	Message error
	Token   token.Token
}

func (parserError *ParserError) Error() string {
	return parserError.Message.Error()
}

func unexpected(thisToken token.Token, context string) error {
	if thisToken.IsKeyword() {
		return fmt.Errorf("Unexpected keyword \"%s\" in %s.", thisToken.String(), context)
	}
	if thisToken.IsOperator() {
		return fmt.Errorf("Unexpected operator \"%s\" in %s.", thisToken.String(), context)
	}
	switch thisToken.Kind {
	case token.Identifier:
		return fmt.Errorf("Unexpected identifier \"%s\" in %s.", thisToken.String(), context)
	case token.EOF:
		return fmt.Errorf("Unexpectedly reached end of file in %s.", context)
	}
	return fmt.Errorf("Unexpected %s in %s", thisToken.Kind, context)
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

func (e *ErrorHandler) HasErrors() bool {
	return len(e.errors) > 0
}

func (e *ErrorHandler) AddUnexpectedErrorWithContext(t token.Token, context string) {
	e.AddError(t, unexpected(t, context))
}

func (e *ErrorHandler) AddExpectError(t token.Token, expectedList ...interface{}) {
	e.AddError(t, expect(t, expectedList))
}

func (e *ErrorHandler) AddError(token token.Token, message error) {
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
	if e.devMode {
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

func (e *ErrorHandler) PanicMessage(message error) {
	fmt.Printf("%s\n", message)
	e.PrintErrors()
	panic(fatalErrorMessage)
}

func (e *ErrorHandler) PanicError(t token.Token, message error) {
	e.AddError(t, fmt.Errorf("%s %s", "**FATAL ERROR**", message))
	e.PrintErrors()
	panic(fatalErrorMessage)
}

func (e *ErrorHandler) PrintErrors() {
	errors := e.errors
	errorCount := 0
	for _, errorList := range errors {
		errorCount += len(errorList)
	}
	//if p.Scanner.Error != nil {
	//	errorCount += 1
	//}
	if errorCount > 0 {
		errorOrErrors := "errors"
		if errorCount == 1 {
			errorOrErrors = "error"
		}
		fmt.Printf("Found %d %s...\n", errorCount, errorOrErrors)
		//if p.Scanner.Error != nil {
		//	fmt.Printf("File: %s\n", p.Scanner.Filepath)
		//	fmt.Printf("- %s \n", p.Scanner.Error)
		//}

		isFirst := true
		for filepath, errorList := range errors {
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
