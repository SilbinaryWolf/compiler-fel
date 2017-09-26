package data

import (
	//"byt
	"fmt"
	//"strconv"
)

type CSSSelectorPartKind int

const (
	SelectorKindUnknown    CSSSelectorPartKind = 0 + iota
	SelectorKindIdentifier                     // .a-class, button
	SelectorKindAttribute                      // [type="text"]

	css_selector_operator_begin
	SelectorKindColon       // :
	SelectorKindDoubleColon // ::
	SelectorKindChild       // >
	SelectorKindAncestor    // ' '
	//SelectorKindAncestorExplicit    // >>
	css_selector_operator_end
)

var selectorKindToString = []string{
	SelectorKindUnknown:    "unknown part",
	SelectorKindIdentifier: "identifier",
	SelectorKindAttribute:  "attribute",

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

func (node *CSSSelectorPart) String() string {
	if node.Kind == SelectorKindIdentifier {
		return node.Name
	}
	return selectorKindToString[node.Kind]
}

type CSSProperty struct {
	Name  string
	Value string
}

func (property *CSSProperty) String() string {
	return fmt.Sprintf("%s: %s;", property.Name, property.Value)
}
