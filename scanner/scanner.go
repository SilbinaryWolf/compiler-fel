package scanner

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/token"
)

type ScannerState struct {
	index         int
	lastLineIndex int
	lineNumber    int
}

type Scanner struct {
	ScannerState
	filecontents []byte
	pathname     string
	inPHPMode    bool
}

func New(filecontents []byte, filepath string) *Scanner {
	scanner := new(Scanner)
	scanner.lineNumber = 1
	scanner.filecontents = filecontents
	// NOTE(Jake): Pad the ending of the file
	scanner.filecontents = append(scanner.filecontents, 0, 0, 0, 0, 0, 0, 0, 0)
	scanner.pathname = filepath
	return scanner
}

func (scanner *Scanner) PeekNextToken() token.Token {
	state := scanner.ScannerState
	result := scanner._getNextToken(true)
	scanner.ScannerState = state
	return result
}

func (scanner *Scanner) GetNextToken() token.Token {
	//fmt.Printf("Getting next token...")
	token := scanner._getNextToken(true)
	//token.Debug()
	return token
}

func (scanner *Scanner) GetPosition() int {
	return scanner.index
}

func (scanner *Scanner) GetLine() int {
	return scanner.lineNumber
}

func (scanner *Scanner) incrementLineNumber() {
	scanner.lineNumber += 1
	scanner.lastLineIndex = scanner.index
}

func scannerDeveloperError(message string, arguments ...interface{}) {
	panic(fmt.Sprintf("Developer scanner error: "+message, arguments...))
}

func isEndOfLine(C byte) bool {
	// NOTE: \r technically isn't a newline character, but for simplicity
	//		 we'll treat it as so for Windows line-endings.
	return C == '\r' || C == '\n'
}

func isWhitespace(C byte) bool {
	return C == ' ' || C == '\t' || C == '\v' || C == '\f'
}

func isAlpha(C byte) bool {
	return (C >= 'a' && C <= 'z') || (C >= 'A' && C <= 'Z')
}

func isNumber(C byte) bool {
	return C >= '0' && C <= '9'
}

func eatEndOfLine(scanner *Scanner) bool {
	C := scanner.getChar(0)
	C2 := scanner.getChar(1)

	if C == '\r' && C2 == '\n' {
		// Windows line-endings
		scanner.incrementLineNumber()
		scanner.index += 2
		return true
	} else if C == '\n' {
		// Unix line-endings
		scanner.incrementLineNumber()
		scanner.index++
		return true
	}
	return false
}

func eatAllWhitespaceAndComments(scanner *Scanner, eatNewline bool, eatWhitespace bool) {
	commentBlockDepth := 0

	for {
		if eatNewline && eatEndOfLine(scanner) {
			continue
		} else if eatWhitespace && isWhitespace(scanner.getChar(0)) {
			scanner.index++
		} else if scanner.getChar(0) == '#' || (scanner.getChar(0) == '/' && scanner.getChar(1) == '/') {
			scanner.index += 2
			for scanner.getChar(0) != 0 && !isEndOfLine(scanner.getChar(0)) {
				scanner.index++
			}
			eatEndOfLine(scanner)
		} else if scanner.getChar(0) == '/' && scanner.getChar(1) == '*' {
			commentBlockDepth += 1
			scanner.index += 2
			for scanner.getChar(0) != 0 && commentBlockDepth > 0 {
				if scanner.getChar(0) == '/' && scanner.getChar(1) == '*' {
					commentBlockDepth += 1
					scanner.index += 2
				} else if scanner.getChar(0) == '*' && scanner.getChar(1) == '/' {
					commentBlockDepth -= 1
					scanner.index += 2
				} else if !eatEndOfLine(scanner) {
					scanner.index++
				}
			}
		} else {
			break
		}
	}
}

func (scanner *Scanner) getChar(lookAhead int) byte {
	index := scanner.index + lookAhead
	if index >= 0 && index < len(scanner.filecontents) {
		r := scanner.filecontents[index]
		return r
	}
	return 0
}

/*func (scanner *Scanner) peekNextTokenIncludeNewline() token.Token {
	state := scanner.ScannerState
	token := scanner._getNextToken(false)
	scanner.ScannerState = state
	return token
}

func (scanner *Scanner) getNextTokenIncludeNewline() token.Token {
	//fmt.Printf("Getting next token...")
	token := scanner._getNextToken(false)
	//token.Debug()
	return token
}*/

func (scanner *Scanner) _getNextToken(eatNewline bool) token.Token {
	t := token.Token{}
	t.Kind = token.Unknown
	t.Pathname = scanner.pathname
	defer func() {
		if t.Kind == token.Unknown {
			scannerDeveloperError("Token kind not set properly by developer")
		}
	}()

	eatAllWhitespaceAndComments(scanner, eatNewline, true)

	C := scanner.getChar(0)
	t.Start = scanner.index
	scanner.index++
	switch C {
	case 0:
		t.Kind = token.EOF
	case '@':
		t.Kind = token.At
	case '(':
		t.Kind = token.ParenOpen
	case ')':
		t.Kind = token.ParenClose
	case '[':
		t.Kind = token.BracketOpen
	case ']':
		t.Kind = token.BracketClose
	case '{':
		t.Kind = token.BraceOpen
	case '}':
		t.Kind = token.BraceClose
	case '%':
		t.Kind = token.Modulo
	case ',':
		t.Kind = token.Comma
	case ';':
		t.Kind = token.Semicolon
	case '$':
		t.Kind = token.InteropVariable
		t.Start++
		// todo(Jake): Enforce cannot have number after $, must be alpha or _
		//if isAlpha(scanner.getChar(0)) || scanner.getChar(0) == '_' {
		//	scanner.index++
		//}
		for scanner.index < len(scanner.filecontents) &&
			(isAlpha(scanner.getChar(0)) || isNumber(scanner.getChar(0)) || scanner.getChar(0) == '_') {
			scanner.index++
		}
	case '"', '\'', '`':
		t.Kind = token.String
		t.Start++
		for scanner.index < len(scanner.filecontents) &&
			scanner.getChar(0) != C {
			if scanner.getChar(0) == '\\' {
				// Skip command code
				scanner.index++
			}
			scanner.index++
		}
		t.End = scanner.index
		scanner.index++
	case ':':
		t.Kind = token.Declare
		nextC := scanner.getChar(0)
		switch nextC {
		case C:
			t.Kind = token.Define
			scanner.index++
		case '=':
			t.Kind = token.DeclareSet
			scanner.index++
		}
	// Operators
	case '+':
		t.Kind = token.Add
	case '-':
		t.Kind = token.Subtract
	case '/':
		t.Kind = token.Divide
	case '*':
		t.Kind = token.Multiply
	case '!':
		t.Kind = token.Not
	case '^':
		t.Kind = token.Power
	case '>':
		t.Kind = token.GreaterThan
	case '<':
		t.Kind = token.LessThan
	case '?':
		t.Kind = token.Ternary
	case '&':
		t.Kind = token.And
		if scanner.getChar(0) == C {
			t.Kind = token.ConditionalAnd
			scanner.index++
		}
	case '|':
		t.Kind = token.Or
		if scanner.getChar(0) == C {
			t.Kind = token.ConditionalOr
			scanner.index++
		}
	case '=':
		t.Kind = token.Equal
		if scanner.getChar(0) == C {
			t.Kind = token.ConditionalEqual
			scanner.index++
		}
	// Other
	default:
		if isEndOfLine(C) {
			t.Kind = token.Newline
			// Check for \r for Windows line endings
			if isEndOfLine(scanner.getChar(0)) {
				scanner.index++
			}
			scanner.incrementLineNumber()
		} else if C == '\\' || C == '_' || isAlpha(C) {
			t.Kind = token.Identifier
			for scanner.index < len(scanner.filecontents) &&
				(isAlpha(scanner.getChar(0)) || isNumber(scanner.getChar(0)) || scanner.getChar(0) == '\\' || scanner.getChar(0) == '_' || scanner.getChar(0) == '.') {
				scanner.index++
			}
			identifierOrKeyword := string(scanner.filecontents[t.Start:scanner.index])
			keywordKind := token.GetKeywordKindFromString(identifierOrKeyword)
			if keywordKind != token.Unknown {
				t.Kind = keywordKind
				t.Data = identifierOrKeyword
			}
		} else if C == '.' || isNumber(C) {
			if C == '.' && !isNumber(scanner.getChar(0)) {
				t.Kind = token.Dot
			} else {
				t.Kind = token.Number

				for isNumber(scanner.getChar(0)) || scanner.getChar(0) == '.' {
					scanner.index++
				}
			}
		} else {
			panic(fmt.Sprintf("Unknown token type found in getToken(): %q (%v), at Line %d (%s)", C, C, scanner.lineNumber, scanner.pathname))
		}
	}
	if t.Start > len(scanner.filecontents) {
		t.Kind = token.EOF
		return t
	}
	if t.End == 0 {
		t.End = scanner.index
	}
	t.Line = scanner.lineNumber
	t.Column = scanner.index - scanner.lastLineIndex
	if len(t.Data) == 0 && t.HasUniqueData() {
		t.Data = string(scanner.filecontents[t.Start:t.End])
	}
	return t
}
