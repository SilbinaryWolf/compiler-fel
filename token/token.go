package token

type Kind int

const (
	Unknown Kind = 0 + iota
	Illegal
	EOF
	Newline      // \r, \n
	Whitespace   // ' ', \t
	ParenOpen    // (
	ParenClose   // )
	BraceOpen    // {
	BraceClose   // }
	BracketOpen  // [
	BracketClose // ]
	Comma        // ,
	Colon        // :
	DeclareSet   // :=
	DoubleColon  // ::
	Semicolon    // ;
	Dot          // .
	Tilde        // ~
	Hash         // #
	At           // @

	unique_begin

	Identifier      // ABunchOfUnquotedLetters
	AtKeyword       // @import, @media
	InteropVariable // $var
	Number          // 30, 1.462
	NumberWithUnit  // 100%, 32px, 5.5em
	Character       // 'C'
	String          // "ABunchOfQuotedLetters"

	unique_end

	keyword_begin

	KeywordIf
	KeywordElse
	KeywordFor
	KeywordTrue
	KeywordFalse
	//KeywordConfig
	//KeywordHTML

	keyword_end

	operator_begin
	Operator
	// Operators and delimiters
	Add                 // +
	AddEqual            // +=
	Subtract            // -
	Divide              // /
	Multiply            // *
	Modulo              // %
	Ternary             // ?
	Equal               // =
	Power               // ^
	PowerEqual          // ^=
	And                 // &
	Or                  // |
	Not                 // !
	ConditionalNotEqual // !=
	ConditionalEqual    // ==
	ConditionalAnd      // &&
	ConditionalOr       // ||
	GreaterThan         // >
	LessThan            // <
	operator_end
)

var precedence = []int{
	Unknown:             0,
	ParenOpen:           1,
	ParenClose:          1,
	ConditionalOr:       2,
	ConditionalAnd:      2,
	ConditionalEqual:    3,
	ConditionalNotEqual: 3, // NOTE(Jake): Didn't check this against other langs
	Add:                 4,
	Subtract:            4,
	Divide:              4,
	Multiply:            4,
}

var kindToString = []string{
	Unknown:         "unset token",
	Illegal:         "illegal token",
	EOF:             "eof",
	Newline:         "newline",
	Whitespace:      " ",
	InteropVariable: "interop variable",

	Identifier: "identifier",
	AtKeyword:  "at-keyword",

	Number:         "number",
	NumberWithUnit: "number with unit",
	ParenOpen:      "(",
	ParenClose:     ")",
	BraceOpen:      "{",
	BraceClose:     "}",
	BracketOpen:    "[",
	BracketClose:   "]",
	Comma:          ",",
	Colon:          ":",
	DeclareSet:     ":=",
	DoubleColon:    "::",
	Semicolon:      ";",
	Character:      "character",
	String:         "string",
	Not:            "!",
	Dot:            ".",
	Hash:           "#",
	Tilde:          "~",
	At:             "@",
	Ternary:        "?",

	KeywordIf:    "if",
	KeywordElse:  "else",
	KeywordFor:   "for",
	KeywordTrue:  "true",
	KeywordFalse: "false",
	//KeywordConfig: "config",
	//KeywordHTML:   "html",

	// Operators and delimiters
	Operator:            "operator",
	Add:                 "+",
	AddEqual:            "+=",
	Subtract:            "-",
	Divide:              "/",
	Multiply:            "*",
	Modulo:              "%",
	Equal:               "=",
	Power:               "^",
	PowerEqual:          "^=",
	And:                 "&",
	Or:                  "|",
	ConditionalNotEqual: "!=",
	ConditionalEqual:    "==",
	ConditionalAnd:      "&&",
	ConditionalOr:       "||",
	GreaterThan:         ">",
	LessThan:            "<",
}

type Token struct {
	Kind     Kind
	Data     string
	Line     int
	Column   int
	Start    int
	End      int
	Filepath string
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
	case Illegal:
		// no-o
	case Whitespace:
		return " "
	}
	if token.HasUniqueData() {
		return token.Data
	}
	return kindToString[token.Kind]
}

func (token Token) Precedence() int {
	return precedence[token.Kind]
}

/*func (token Token) Debug() {
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
}*/
