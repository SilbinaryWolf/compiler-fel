package data

import (
	//"byt
	"fmt"
	//"strconv"
)

type CSSSelectorPartKind int

const (
	SelectorKindUnknown   CSSSelectorPartKind = 0 + iota
	SelectorKindAttribute                     // [type="text"]

	css_selector_identifier_begin
	SelectorKindClass     // .a-class
	SelectorKindTag       // button
	SelectorKindID        // #main
	SelectorKindAtKeyword // @media
	SelectorKindNumber    // 3.3
	css_selector_identifier_end

	css_selector_operator_begin
	SelectorKindColon       // :
	SelectorKindDoubleColon // ::
	SelectorKindChild       // >
	SelectorKindAncestor    // ' '
	//SelectorKindAncestorExplicit    // >>
	css_selector_operator_end
)

var selectorKindToString = []string{
	SelectorKindUnknown:   "unknown part",
	SelectorKindAttribute: "attribute",

	//SelectorKindIdentifier: "identifier",
	SelectorKindClass: "class",
	SelectorKindID:    "id",
	SelectorKindTag:   "tag",

	SelectorKindColon:       ":",
	SelectorKindDoubleColon: "::",
	SelectorKindChild:       ">",
	SelectorKindAncestor:    " ",
	//SelectorKindAncestorExplicit:           ">>",
}

type CSSDefinition struct {
	Name       string
	ChildNodes []*CSSRule
}

type CSSRule struct {
	Selectors  []CSSSelector
	Properties []CSSProperty
	Rules      []*CSSRule
}

type CSSSelector []CSSSelectorPart

func (nodes CSSSelector) String() string {
	result := ""
	for _, node := range nodes {
		result += node.String() + " "
	}
	result = result[:len(result)-1]
	return result
}

type CSSSelectorPart struct {
	Kind CSSSelectorPartKind
	// CSSSelectorIdentifier / CSSSelectorAttribute
	Name string
	// CSSSelectorAttribute
	Operator string
	Value    string
}

func (kind CSSSelectorPartKind) IsOperator() bool {
	return kind > css_selector_operator_begin && kind < css_selector_operator_end
}

func (kind CSSSelectorPartKind) IsIdentifier() bool {
	return kind > css_selector_identifier_begin && kind < css_selector_identifier_end
}

func (kind CSSSelectorPartKind) String() string {
	return selectorKindToString[kind]
}

func (node *CSSSelectorPart) String() string {
	if node.Kind.IsIdentifier() {
		return node.Name
	}
	return node.Kind.String()
}

type CSSProperty struct {
	Name  string
	Value string
}

func (property *CSSProperty) String() string {
	return fmt.Sprintf("%s: %s;", property.Name, property.Value)
}
