package token

import (
	"fmt"
	"runtime"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	EOF
	Newline      // \r, \n
	Whitespace   // \t
	ParenOpen    // (
	ParenClose   // )
	BraceOpen    // {
	BraceClose   // }
	BracketOpen  // [
	BracketClose // ]
	Comma        // ,
	Declare      // :
	DeclareSet   // :=
	Define       // ::
	Semicolon    // ;
	Dot          // .
	Hash         // #
	At           // @

	unique_begin

	Identifier      // ABunchOfUnquotedLetters
	InteropVariable // $var
	Number          // 30, 1.462
	Character       // 'C'
	String          // "ABunchOfQuotedLetters"

	unique_end

	keyword_begin

	KeywordIf
	KeywordElse
	KeywordFor
	//KeywordConfig
	//KeywordHTML

	keyword_end

	operator_begin
	// Operators and delimiters
	Add              // +
	Subtract         // -
	Divide           // /
	Multiply         // *
	Modulo           // %
	Ternary          // ?
	Equal            // =
	And              // &
	Or               // |
	ConditionalEqual // ==
	ConditionalAnd   // &&
	ConditionalOr    // ||
	Not              // !
	GreaterThan      // >
	LessThan         // <
	Power            // ^
	operator_end
)

var precedence = []int{
	Unknown:          0,
	ParenOpen:        1,
	ParenClose:       1,
	ConditionalOr:    2,
	ConditionalAnd:   2,
	ConditionalEqual: 3,
	Add:              4,
	Subtract:         4,
	Divide:           4,
	Multiply:         4,
}

var kindToString = []string{
	Unknown:         "unknown token",
	EOF:             "eof",
	Newline:         "newline",
	Whitespace:      " ",
	InteropVariable: "interop variable",

	Identifier: "identifier",

	//Number       // 30, 1.462
	ParenOpen:    "(",
	ParenClose:   ")",
	BraceOpen:    "{",
	BraceClose:   "}",
	BracketOpen:  "[",
	BracketClose: "]",
	Comma:        ",",
	Declare:      ":",
	DeclareSet:   ":=",
	Define:       "::",
	Semicolon:    ";",
	Character:    "character",
	String:       "string",
	Not:          "!",
	Dot:          ".",
	Hash:         "#",
	Ternary:      "?",

	KeywordIf:   "if",
	KeywordElse: "else",
	KeywordFor:  "for",
	//KeywordConfig: "config",
	//KeywordHTML:   "html",

	// Operators and delimiters
	Add:              "+",
	Subtract:         "-",
	Divide:           "/",
	Multiply:         "*",
	Modulo:           "%",
	Equal:            "=",
	And:              "&",
	Or:               "|",
	ConditionalEqual: "==",
	ConditionalAnd:   "&&",
	ConditionalOr:    "||",
	GreaterThan:      ">",
	LessThan:         "<",
	Power:            "^",
}

type Token struct {
	Kind   Kind
	Data   string
	Line   int
	Column int
	Start  int
	End    int
}

func GetKeywordKindFromString(keyword string) Kind {
	for kindIndex := keyword_begin + 1; kindIndex < keyword_end; kindIndex++ {
		if keyword == kindToString[kindIndex] {
			return kindIndex
		}
	}
	return Unknown
}

func (token Token) IsOperator() bool {
	return token.Kind > operator_begin && token.Kind < operator_end
}

func (token Token) IsKeyword() bool {
	return token.Kind > keyword_begin && token.Kind < keyword_end
}

func (token Token) HasUniqueData() bool {
	return token.Kind > unique_begin && token.Kind < unique_end
}

func (kind Kind) String() string {
	return kindToString[kind]
}

func (token Token) String() string {
	switch token.Kind {
	case InteropVariable:
		return token.Data
	case Identifier, Number:
		return token.Data
	case String:
		// NOTE(Jake): Change this so string outputs properly in Evaluator
		//return fmt.Sprintf("\"%s\"", token.Data)
		return token.Data
	case Whitespace:
		return " "
	}
	return kindToString[token.Kind]
}

func (token Token) Precedence() int {
	return precedence[token.Kind]
}

func (token Token) Debug() {
	// Get callee function stuff
	fpcs := make([]uintptr, 1)
	runtime.Callers(3, fpcs)
	fn := runtime.FuncForPC(fpcs[0] - 1)
	_, fnLine := fn.FileLine(fpcs[0] - 1)

	fmt.Printf("Debug Token: %v (Func. Line: %-4v, Func: %v)\n", token.DebugString(), fnLine, fn.Name())
}

func (token Token) DebugString() string {
	var result string
	result = fmt.Sprintf("%-10v (Line: %v)", token.String(), token.Line)
	return result
}
