package scanner

import (
	"fmt"
	"unicode/utf8"

	"github.com/silbinarywolf/compiler-fel/token"
)

type ScannerState struct {
	index         int
	lastLineIndex int // Helps calculate column on token
	lineNumber    int
}

type Scanner struct {
	ScannerState
	filecontents []byte
	Filepath     string
}

const BYTE_ORDER_MARK = 0xFEFF // byte order mark, only permitted as very first character
const END_OF_FILE = 0

func New(filecontents []byte, filepath string) *Scanner {
	scanner := new(Scanner)
	scanner.lineNumber = 1
	scanner.filecontents = filecontents
	// NOTE(Jake): Pad the ending of the file
	scanner.filecontents = append(scanner.filecontents, 0, 0, 0, 0, 0, 0, 0, 0)
	scanner.Filepath = filepath
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

func isEndOfLine(C rune) bool {
	// NOTE: \r technically isn't a newline character, but for simplicity
	//		 we'll treat it as so for Windows line-endings.
	return C == '\r' || C == '\n'
}

func isWhitespace(C rune) bool {
	return C == ' ' || C == '\t' || C == '\v' || C == '\f'
}

func isAlpha(C rune) bool {
	return (C >= 'a' && C <= 'z') || (C >= 'A' && C <= 'Z')
}

func isNumber(C rune) bool {
	return C >= '0' && C <= '9'
}

func eatEndOfLine(scanner *Scanner) bool {
	lastIndex := scanner.index
	C := scanner.nextRune()
	C2 := scanner.nextRune()

	if C == '\r' && C2 == '\n' {
		// Windows line-endings
		scanner.incrementLineNumber()
		return true
	}
	if C == '\n' {
		// Unix line-endings
		scanner.incrementLineNumber()
		return true
	}
	scanner.index = lastIndex
	return false
}

func eatAllWhitespaceAndComments(scanner *Scanner, eatNewline bool, eatWhitespace bool) {
	commentBlockDepth := 0

	for {
		if eatNewline && eatEndOfLine(scanner) {
			continue
		}
		lastIndex := scanner.index
		C := scanner.nextRune()
		C2 := scanner.nextRune()
		if eatWhitespace && isWhitespace(C) {
			continue
		}
		if C == '/' && C2 == '/' {
			for {
				C := scanner.nextRune()
				if C == 0 || isEndOfLine(C) {
					break
				}
			}
			//eatEndOfLine(scanner)
			continue
		}
		if C == '/' && C2 == '*' {
			commentBlockDepth += 1
			for {
				lastIndex := scanner.index
				C := scanner.nextRune()
				if C == 0 || commentBlockDepth == 0 {
					break
				}
				C2 := scanner.nextRune()
				if C == '/' && C2 == '*' {
					commentBlockDepth += 1
					continue
				}
				if C == '*' && C2 == '/' {
					commentBlockDepth -= 1
					continue
				}
				//if !eatEndOfLine(scanner) {
				//	scanner.index++
				//}
				scanner.index = lastIndex
				break
			}
			continue
		}

		// If no matches, rewind and break
		scanner.index = lastIndex
		break
	}
}

/*func (scanner *Scanner) getByte() byte {
	if index >= 0 && index < len(scanner.filecontents) {
		return scanner.filecontents[index]
	}
	return 0
}*/

func (scanner *Scanner) nextRune() rune {
	index := scanner.index
	if index < 0 && index >= len(scanner.filecontents) {
		return 0
	}
	r, size := rune(scanner.filecontents[index]), 1
	switch {
	case r == 0:
		panic("Illegal character NUL")
	case r >= utf8.RuneSelf:
		// not ASCII
		r, size = utf8.DecodeRune(scanner.filecontents[index:])
		if r == utf8.RuneError && size == 1 {
			panic("illegal UTF-8 encoding")
		} else if r == BYTE_ORDER_MARK && scanner.index > 0 {
			panic("illegal byte order mark")
		}
	}
	scanner.index += size
	return r
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
	defer func() {
		if t.Kind == token.Unknown {
			scannerDeveloperError("Token kind not set properly by developer")
		}
	}()

	eatAllWhitespaceAndComments(scanner, false, true)

	t.Start = scanner.index
	C := scanner.nextRune()
	t.Start = scanner.index
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
		for {
			C := scanner.nextRune()
			if C != END_OF_FILE &&
				(isAlpha(C) || isNumber(C) || C == '_') {
				continue
			}
			break
		}
	case '\'':
		panic("Character string unsupported.")
		/*t.Kind = token.Character
		t.Start++
		for {
			nextC := scanner.nextRune(0)
			if C != END_OF_FILE && C != nextC {
				if nextC == '\\' {
					// Skip command code
					scanner.index++
				}
				scanner.index++
				continue
			}
			break
		}
		t.End = scanner.index
		scanner.index++*/
	case '"', '`':
		t.Kind = token.Character
		t.Start++
		for {
			nextC := scanner.nextRune()
			if C != END_OF_FILE && C != nextC {
				if nextC == '\\' {
					// Skip command code
					scanner.index++
				}
				scanner.index++
				continue
			}
			break
		}
		t.End = scanner.index
		scanner.index++
	case ':':
		t.Kind = token.Declare
		switch lastIndex := scanner.index; scanner.nextRune() {
		case C:
			t.Kind = token.Define
		case '=':
			t.Kind = token.DeclareSet
		default:
			scanner.index = lastIndex
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
		switch lastIndex := scanner.index; scanner.nextRune() {
		case C:
			t.Kind = token.ConditionalAnd
		default:
			scanner.index = lastIndex
		}
	case '|':
		t.Kind = token.Or
		switch lastIndex := scanner.index; scanner.nextRune() {
		case C:
			t.Kind = token.ConditionalOr
		default:
			scanner.index = lastIndex
		}
	case '=':
		t.Kind = token.Equal
		switch lastIndex := scanner.index; scanner.nextRune() {
		case C:
			t.Kind = token.ConditionalEqual
		default:
			scanner.index = lastIndex
		}
	// Other
	default:
		if isEndOfLine(C) {
			t.Kind = token.Newline
			// Check for \r for Windows line endings
			lastIndex := scanner.index
			if C := scanner.nextRune(); !isEndOfLine(C) {
				scanner.index = lastIndex
			}
			scanner.incrementLineNumber()
		} else if C == '\\' || C == '_' || isAlpha(C) {
			t.Kind = token.Identifier
			for {
				lastIndex := scanner.index
				C := scanner.nextRune()
				if scanner.index < len(scanner.filecontents) &&
					(isAlpha(C) || isNumber(C) || C == '\\' || C == '-' || C == '_' || C == '.') {
					continue
				}
				scanner.index = lastIndex
				break
			}
			identifierOrKeyword := string(scanner.filecontents[t.Start:scanner.index])
			keywordKind := token.GetKeywordKindFromString(identifierOrKeyword)
			if keywordKind != token.Unknown {
				t.Kind = keywordKind
				t.Data = identifierOrKeyword
			}
		} else if C == '.' || isNumber(C) {
			lastIndex := scanner.index
			C2 := scanner.nextRune()
			if C == '.' && !isNumber(C2) {
				// Detect .567
				t.Kind = token.Dot
			} else {
				// Regular number
				scanner.index = lastIndex

				// Regular number
				t.Kind = token.Number
				for {
					lastIndex := scanner.index
					C := scanner.nextRune()
					if isNumber(C) || C == '.' {
						continue
					}
					scanner.index = lastIndex
					break
				}
			}
		} else {
			panic(fmt.Sprintf("Unknown token type found in getToken(): %q (%v), at Line %d (%s)", C, C, scanner.lineNumber, scanner.Filepath))
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
