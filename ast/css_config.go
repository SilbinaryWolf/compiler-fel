package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
)

type CSSConfigDefinition struct {
	Name token.Token
	Base
}

/*type CSSConfigRule struct {
	SelectorWildcards []CSSSelectorWildcard
	Base
}

type CSSSelectorWildcard struct {
	Base
}

type CSSConfigProperty struct {
	Name token.Token
	Base
}*/
