package scanner

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/silbinarywolf/compiler-fel/token"
)

type Mode int

const (
	ModeDefault Mode = 0 + iota
	ModeCSS
)

type scannerState struct {
	index                    int
	indexAtLastLineIncrement int // Helps calculate column on token
	lineNumber               int
}

type Scanner struct {
	scannerState
	scanmode     Mode
	filecontents []byte
	Filepath     string
	Error        error

	//peekToken token.Token
}

const BYTE_ORDER_MARK = 0xFEFF // byte order mark, only permitted as very first character
const END_OF_FILE = 0

func New(filecontents []byte, filepath string) *Scanner {
	scanner := new(Scanner)
	scanner.lineNumber = 1
	scanner.filecontents = filecontents
	scanner.Filepath = filepath
	return scanner
}

func (scanner *Scanner) PeekNextToken() token.Token {
	// NOTE: Try to speedup PeekNextToken()
	//if scanner.peekToken.Kind != token.Unknown {
	//	return scanner.peekToken
	//}

	state := scanner.ScannerState()
	result := scanner._getNextToken()
	//scanner.peekToken = result
	scanner.SetScannerState(state)
	return result
}

func (scanner *scannerState) Column() int {
	return scanner.index - scanner.indexAtLastLineIncrement
}

func (scanner *scannerState) Line() int {
	return scanner.lineNumber
}

func (scanner *Scanner) ScannerState() scannerState {
	return scanner.scannerState
}

func (scanner *Scanner) SetScannerState(state scannerState) {
	scanner.scannerState = state
	//scanner.peekToken.Kind = token.Unknown
}

func (scanner *Scanner) GetNextToken() token.Token {
	// NOTE: Try to speedup PeekNextToken()
	//if scanner.peekToken.Kind != token.Unknown {
	//	result := scanner.peekToken
	//	scanner.peekToken.Kind = token.Unknown
	//	return result
	//}

	//fmt.Printf("Getting next token...")
	result := scanner._getNextToken()
	//token.Debug()
	return result
}

func (scanner *Scanner) SetScanMode(scanmode Mode) {
	scanner.scanmode = scanmode
	//scanner.peekToken.Kind = token.Unknown
}

func (scanner *Scanner) incrementLineNumber() {
	scanner.lineNumber += 1
	scanner.indexAtLastLineIncrement = scanner.index
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
	return C != '\n' && unicode.IsSpace(C)
}

func isAlpha(C rune) bool {
	return (C >= 'a' && C <= 'z') || (C >= 'A' && C <= 'Z') || (C >= utf8.RuneSelf && unicode.IsLetter(C))
}

func isNumber(C rune) bool {
	return (C >= '0' && C <= '9') || (C >= utf8.RuneSelf && unicode.IsDigit(C))
}

func (scanner *Scanner) eatEndOfLine() bool {
	lastIndex := scanner.index
	C := scanner.nextRune()
	if C == '\n' {
		// Unix line-endings
		scanner.incrementLineNumber()
		return true
	}
	C2 := scanner.nextRune()
	if C == '\r' && C2 == '\n' {
		// Windows line-endings
		scanner.incrementLineNumber()
		return true
	}
	scanner.index = lastIndex
	return false
}

func (scanner *Scanner) eatAllWhitespace() {
	for {
		lastIndex := scanner.index
		C := scanner.nextRune()
		if isWhitespace(C) {
			continue
		}
		// If no matches, rewind and break
		scanner.index = lastIndex
		break
	}
}

func (scanner *Scanner) eatAllComments() {
	commentBlockDepth := 0

	for {
		lastIndex := scanner.index
		C := scanner.nextRune()
		C2 := scanner.nextRune()
		if C == '/' && C2 == '/' {
			for {
				C := scanner.nextRune()
				if scanner.eatEndOfLine() {
					break
				}
				if C == 0 {
					break
				}
			}
			continue
		}
		if C == '/' && C2 == '*' {
			commentBlockDepth += 1
			for {
				if scanner.eatEndOfLine() {
					continue
				}
				C := scanner.nextRune()
				if C == 0 {
					scanner.setError(fmt.Errorf("NUL character found in comment block."))
					break
				}
				lastIndex := scanner.index
				C2 := scanner.nextRune()
				//fmt.Printf("-- COMMENT CHECK: %s%s\n", string(C), string(C2))
				if C == '/' && C2 == '*' {
					commentBlockDepth += 1
					continue
				}
				if C == '*' && C2 == '/' {
					commentBlockDepth -= 1
					if commentBlockDepth <= 0 {
						break
					}
					continue
				}
				scanner.index = lastIndex
			}
			continue
		}

		// If no matches, rewind and break
		scanner.index = lastIndex
		break
	}
}

func (scanner *Scanner) HasErrors() bool {
	return scanner.Error != nil
}

func (scanner *Scanner) setError(message error) {
	scanner.Error = fmt.Errorf("Line %d - %v", scanner.lineNumber, message)
}

func (scanner *Scanner) nextRune() rune {
	index := scanner.index
	if index < 0 || index >= len(scanner.filecontents) {
		return END_OF_FILE
	}
	r, size := rune(scanner.filecontents[index]), 1
	if r == 0 {
		scanner.setError(fmt.Errorf("Illegal character NUL."))
		return END_OF_FILE
	}
	if r >= utf8.RuneSelf {
		r, size = utf8.DecodeRune(scanner.filecontents[index:])
		if r == utf8.RuneError && size == 1 {
			scanner.setError(fmt.Errorf("Illegal UTF-8 encoding."))
			return END_OF_FILE
		} else if r == BYTE_ORDER_MARK && scanner.index > 0 {
			scanner.setError(fmt.Errorf("Illegal byte order mark."))
			return END_OF_FILE
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

func (scanner *Scanner) scanIdentifier() {
	for {
		lastIndex := scanner.index
		C := scanner.nextRune()
		if scanner.index < len(scanner.filecontents) &&
			(isAlpha(C) || isNumber(C) || C == '-' || C == '_' || (scanner.scanmode == ModeCSS && C == '.')) {
			continue
		}
		scanner.index = lastIndex
		break
	}
}

func (scanner *Scanner) _getNextToken() token.Token {
	t := token.Token{}
	t.Kind = token.Unknown
	defer func() {
		if t.Kind == token.Unknown {
			scannerDeveloperError("Token kind not set properly by developer")
		}
	}()

	// NOTE: Eating whitespace before *and* after comments, avoids
	//	     bug where scanner falls over on "\t" or similar
	if scanner.scanmode != ModeCSS {
		scanner.eatAllWhitespace()
	}
	scanner.eatAllComments()
	if scanner.scanmode != ModeCSS {
		scanner.eatAllWhitespace()
	}

	t.Start = scanner.index
	C := scanner.nextRune()
	switch C {
	case 0:
		t.Kind = token.EOF
	case '@':
		t.Kind = token.At

		// Scan at-keyword
		if scanner.scanmode == ModeCSS {
			t.Kind = token.AtKeyword
			scanner.scanIdentifier()
		}
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
	case '~':
		t.Kind = token.Tilde
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
	case '"', '\'':
		if scanner.scanmode == ModeDefault && C == '\'' {
			// NOTE(Jake): Enforce that ' cannot be used anywhere eventually. Ideally a 'felfmt' tool would fix
			//			   those types of strings for you.
			scanner.setError(fmt.Errorf("Cannot use ' character for strings outside of \":: css\" definitions."))
		}

		// Handle HereDoc (triple quote """)
		if scanner.scanmode == ModeDefault && C == '"' {
			lastIndex := scanner.index
			C2 := scanner.nextRune()
			C3 := scanner.nextRune()
			if C2 == '"' && C3 == '"' {
				t.Kind = token.String
				t.Start = scanner.index
				for {
					t.End = scanner.index
					C := scanner.nextRune()
					C2 := scanner.nextRune()
					C3 := scanner.nextRune()
					if C == '"' && C2 == '"' && C3 == '"' {
						break
					}
					if C == END_OF_FILE || C2 == END_OF_FILE || C3 == END_OF_FILE {
						panic("Expected end of heredoc string but instead got end of file.")
					}
				}
				//panic(string(scanner.filecontents[t.Start:t.End]))
				break
			}
			scanner.index = lastIndex
		}

		t.Kind = token.String
		t.Start = scanner.index

		for {
			t.End = scanner.index
			subC := scanner.nextRune()
			if subC == C {
				break
			}
			if subC == END_OF_FILE {
				scanner.setError(fmt.Errorf("Expected end of string but instead got end of file."))
			}
		}
	case ':':
		t.Kind = token.Colon
		switch lastIndex := scanner.index; scanner.nextRune() {
		case C:
			t.Kind = token.DoubleColon
		case '=':
			t.Kind = token.DeclareSet
		default:
			scanner.index = lastIndex
		}
	// Operators
	case '+':
		t.Kind = token.Add
		switch lastIndex := scanner.index; scanner.nextRune() {
		case '=':
			t.Kind = token.AddEqual
		default:
			scanner.index = lastIndex
		}
	// todo(Jake): Handle subtract again (identifiers have '-' so needs extra work)
	//case '-':
	///./.;'>">">'	t.Kind = token.Subtract
	case '/':
		t.Kind = token.Divide
	case '*':
		t.Kind = token.Multiply
	case '!':
		t.Kind = token.Not
		switch lastIndex := scanner.index; scanner.nextRune() {
		case '=':
			t.Kind = token.ConditionalNotEqual
		default:
			scanner.index = lastIndex
		}
	case '^':
		t.Kind = token.Power
		switch lastIndex := scanner.index; scanner.nextRune() {
		case '=':
			t.Kind = token.PowerEqual
		default:
			scanner.index = lastIndex
		}
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
		if scanner.scanmode == ModeCSS && isWhitespace(C) {
			t.Kind = token.Whitespace
			for {
				lastIndex := scanner.index
				C := scanner.nextRune()
				if isWhitespace(C) {
					continue
				}
				scanner.index = lastIndex
				break
			}
		} else if isEndOfLine(C) {
			t.Kind = token.Newline
			// Consume \n after \r for Windows line-endings
			if C == '\r' {
				if lastIndex := scanner.index; scanner.nextRune() != '\n' {
					scanner.index = lastIndex
				}
			}
			scanner.incrementLineNumber()
		} else if C == '_' || C == '-' || isAlpha(C) ||
			(scanner.scanmode == ModeCSS && (C == '.' || C == '#')) {
			t.Kind = token.Identifier
			scanner.scanIdentifier()
			identifierOrKeyword := string(scanner.filecontents[t.Start:scanner.index])
			keywordKind := token.GetKeywordKindFromString(identifierOrKeyword)
			if keywordKind != token.Unknown {
				t.Kind = keywordKind
				t.Data = identifierOrKeyword
			}
		} else if C == '.' || isNumber(C) {
			lastIndex := scanner.index
			nextIsNotNumber := !isNumber(scanner.nextRune())
			scanner.index = lastIndex

			if C == '.' && nextIsNotNumber {
				t.Kind = token.Dot
			} else {
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

				// Handle %, rem, px, additional non-whitespace text in CSS
				if scanner.scanmode == ModeCSS {
					lastIndex := scanner.index
					C := scanner.nextRune()
					if !isWhitespace(C) && !isEndOfLine(C) && (isAlpha(C) || C == '%') {
						for {
							lastIndex := scanner.index
							C := scanner.nextRune()
							if !isWhitespace(C) && !isEndOfLine(C) && (isAlpha(C) || C == '%') {
								continue
							}
							scanner.index = lastIndex
							break
						}
					} else {
						scanner.index = lastIndex
					}
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
	t.Line = scanner.Line()
	t.Column = scanner.Column()
	if len(t.Data) == 0 && t.HasUniqueData() {
		t.Data = string(scanner.filecontents[t.Start:t.End])
	}
	//if t.Kind == token.Identifier && len(t.Data) > 0 && t.Data[len(t.Data)-1] == '-' {
	//	scanner.setError(fmt.Errorf("Cannot end identifier in -."))
	//}
	if scanner.Error != nil {
		t.Kind = token.Illegal
	}
	t.Filepath = scanner.Filepath
	return t
}
