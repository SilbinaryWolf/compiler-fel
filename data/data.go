package data

import (
	"bytes"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/types"
)

type Struct struct {
	fields   []interface{}
	typeinfo *types.Struct
}

func NewStruct(fieldCount int, typeInfo *types.Struct) *Struct {
	structData := new(Struct)
	structData.typeinfo = typeInfo
	structData.fields = make([]interface{}, fieldCount)
	return structData
}

func (structData *Struct) SetField(index int, value interface{}) {
	structData.fields[index] = value
}

func (structData *Struct) GetField(index int) interface{} {
	return structData.fields[index]
}

func (structData *Struct) GetFieldByName(name string) interface{} {
	field := structData.typeinfo.GetFieldByName(name)
	if field == nil {
		return nil
	}
	return structData.fields[field.Index()]
}

func (structData *Struct) FieldCount() int {
	return len(structData.fields)
}

type StructInterface interface {
	GetField(index int) interface{}
	SetField(index int, value interface{})
}

type HTMLKind int

const (
	HTMLKindUnknown HTMLKind = 0 + iota
	HTMLKindElement
	HTMLKindText
	HTMLKindFragment
)

type HTMLElement struct {
	// NOTE(Jake): Currently "Name" also stores HTMLText data.
	kind       HTMLKind
	nameOrText string
	attributes []HTMLAttribute

	parentNode *HTMLElement
	//previousNode *HTMLElement
	//nextNode     *HTMLElement

	childNodes []*HTMLElement
}

func (node *HTMLElement) ImplementsHTMLInterface() {}

type HTMLAttribute struct {
	Name  string
	Value string
}

func NewHTMLElement(tagName string) *HTMLElement {
	node := new(HTMLElement)
	node.kind = HTMLKindElement
	node.nameOrText = tagName
	return node
}

func NewHTMLText(text string) *HTMLElement {
	node := new(HTMLElement)
	node.kind = HTMLKindText
	node.nameOrText = text
	return node
}

func NewHTMLFragment() *HTMLElement {
	node := new(HTMLElement)
	node.kind = HTMLKindFragment
	return node
}

func (node *HTMLElement) Name() string {
	return node.nameOrText
}

func (node *HTMLElement) Text() string {
	return node.nameOrText
}

func (node *HTMLElement) Kind() HTMLKind {
	return node.kind
}

func (node *HTMLElement) SetParent(parent *HTMLElement) {
	if node.parentNode == parent {
		return
	}
	if node.parentNode != nil {
		// Remove from old parent
		for i, childNode := range node.parentNode.childNodes {
			/*childNode, ok := childNode.(*HTMLElement)
			if !ok {
				continue
			}*/
			if childNode == node {
				node.parentNode.childNodes = append(node.parentNode.childNodes[:i], node.parentNode.childNodes[i+1:]...)
				break
			}
		}
	}
	node.parentNode = parent
	parent.childNodes = append(parent.childNodes, node)
}

func (node *HTMLElement) GetAttributes() []HTMLAttribute {
	return node.attributes
}

func (node *HTMLElement) SetAttribute(name string, value string) {
	for i := 0; i < len(node.attributes); i++ {
		attr := &node.attributes[i]
		if attr.Name == name {
			attr.Value = value
			return
		}
	}
	node.attributes = append(node.attributes, HTMLAttribute{
		Name:  name,
		Value: value,
	})
}

func (node *HTMLElement) GetAttribute(name string) (string, bool) {
	for i := 0; i < len(node.attributes); i++ {
		attr := &node.attributes[i]
		if attr.Name == name {
			return attr.Value, true
		}
	}
	return "", false
}

func (node *HTMLElement) debugIndent(indent int) string {
	var buffer bytes.Buffer
	if tag := node.String(); len(tag) > 0 {
		for i := 0; i < indent; i++ {
			buffer.WriteByte('\t')
		}
		buffer.WriteString(tag)
		buffer.WriteByte('\n')
	}

	if childNodes := node.childNodes; len(childNodes) > 0 {
		indent += 1
		for _, node := range childNodes {
			//buffer.WriteString(node.debugIndent(indent))
			switch node.Kind() {
			case HTMLKindElement:
				buffer.WriteString(node.debugIndent(indent))
			case HTMLKindFragment:
				for _, node := range node.childNodes {
					buffer.WriteString(node.debugIndent(indent))
				}
			case HTMLKindText:
				for i := 0; i < indent; i++ {
					buffer.WriteByte('\t')
				}
				buffer.WriteString(node.Text())
				buffer.WriteByte('\n')
			default:
				panic(fmt.Sprintf("HTMLElement::Debug: Unhandled type %v", node.Kind()))
			}
		}
		indent -= 1
		for i := 0; i < indent; i++ {
			buffer.WriteByte('\t')
		}
		if name := node.Name(); len(name) > 0 {
			buffer.WriteByte('<')
			buffer.WriteByte('/')
			buffer.WriteString(name)
			buffer.WriteByte('>')
			buffer.WriteByte('\n')
		}
	}
	return buffer.String()
}

func (node *HTMLElement) Debug() string {
	return node.debugIndent(0)
}

func (node *HTMLElement) String() string {
	name := node.Name()
	if len(name) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteByte('<')
	buffer.WriteString(node.Name())
	for _, attribute := range node.GetAttributes() {
		buffer.WriteByte(' ')
		buffer.WriteString(attribute.Name)
		buffer.WriteString("=\"")
		buffer.WriteString(attribute.Value)
		buffer.WriteString("\"")
	}
	if len(node.childNodes) == 0 {
		buffer.WriteByte('/')
	}
	buffer.WriteByte('>')
	return buffer.String()
}

type CSSSelectorPartKind int

const (
	SelectorPartKindUnknown   CSSSelectorPartKind = 0 + iota
	SelectorPartKindAttribute                     // [type="text"]

	css_selector_identifier_begin
	SelectorPartKindClass     // .a-class
	SelectorPartKindTag       // button
	SelectorPartKindID        // #main
	SelectorPartKindAtKeyword // @media
	SelectorPartKindNumber    // 3.3
	css_selector_identifier_end

	css_selector_operator_begin
	SelectorPartKindColon       // :
	SelectorPartKindDoubleColon // ::
	SelectorPartKindChild       // >
	SelectorPartKindAncestor    // ' '
	SelectorPartKindSibling     // ~
	SelectorPartKindAdjacent    // +
	//SelectorKindAncestorExplicit    // >>
	css_selector_operator_end
)

var selectorKindToString = []string{
	SelectorPartKindUnknown:   "unknown part",
	SelectorPartKindAttribute: "attribute",

	//SelectorKindIdentifier: "identifier",
	SelectorPartKindClass:     "class",
	SelectorPartKindTag:       "tag",
	SelectorPartKindID:        "id",
	SelectorPartKindAtKeyword: "at-keyword",
	SelectorPartKindNumber:    "number",

	SelectorPartKindColon:       ":",
	SelectorPartKindDoubleColon: "::",
	SelectorPartKindChild:       ">",
	SelectorPartKindAncestor:    " ",
	SelectorPartKindSibling:     "~",
	SelectorPartKindAdjacent:    "+",
	//SelectorKindAncestorExplicit:           ">>",
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

type CSSDefinition struct {
	name  string
	rules []*CSSRule
}

func (def *CSSDefinition) Name() string          { return def.name }
func (def *CSSDefinition) Rules() []*CSSRule     { return def.rules }
func (def *CSSDefinition) AddRule(node *CSSRule) { def.rules = append(def.rules, node) }

func NewCSSDefinition(name string) *CSSDefinition {
	def := new(CSSDefinition)
	def.name = name
	return def
}

type CSSRule struct {
	selectors  []CSSSelector
	properties []CSSProperty
	rules      []*CSSRule
}

func (rule *CSSRule) Selectors() []CSSSelector  { return rule.selectors }
func (rule *CSSRule) Properties() []CSSProperty { return rule.properties }
func (rule *CSSRule) Rules() []*CSSRule         { return rule.rules }
func (rule *CSSRule) AddRule(node *CSSRule)     { rule.rules = append(rule.rules, rule) }

func NewCSSRule(selectors []CSSSelector) *CSSRule {
	rule := new(CSSRule)
	rule.selectors = selectors
	return rule
}

type CSSSelectorPart struct {
	kind CSSSelectorPartKind
	// CSSSelectorIdentifier / CSSSelectorAttribute
	name string
	// CSSSelectorAttribute
	operator string
	value    string
}

func (node *CSSSelectorPart) Kind() CSSSelectorPartKind { return node.kind }
func (node *CSSSelectorPart) Name() string              { return node.name }
func (node *CSSSelectorPart) Operator() string          { return node.operator }
func (node *CSSSelectorPart) Value() string             { return node.value }
func (node *CSSSelectorPart) String() string {
	kind := node.Kind()
	if kind.IsIdentifier() {
		return node.Name()
	}
	return kind.String()
}

func NewCSSSelectorPart(kind CSSSelectorPartKind, name string) *CSSSelectorPart {
	node := new(CSSSelectorPart)
	node.kind = kind
	node.name = name
	return node
}

func NewCSSSelectorAttributePart(name string, operator string, value string) *CSSSelectorPart {
	node := new(CSSSelectorPart)
	node.kind = SelectorPartKindAttribute
	node.name = name
	node.operator = operator
	node.value = value
	return node
}

type CSSSelector []*CSSSelectorPart

func (selector *CSSSelector) AddPart(selectorPart *CSSSelectorPart) {
	*selector = append(*selector, selectorPart)
}

func NewCSSSelector(size int) CSSSelector {
	return make(CSSSelector, 0, size)
}

func (nodes CSSSelector) String() string {
	result := ""
	for _, node := range nodes {
		result += node.String() + " "
	}
	result = result[:len(result)-1]
	return result
}

type CSSProperty struct {
	Name  string
	Value string
}

/*func NewCSSProperty() *CSSProperty {
	prop := new(CSSProperty)
	prop.name = name
	prop.value = value
	return prop
}*/

func (property *CSSProperty) String() string {
	return fmt.Sprintf("%s: %s;", property.Name, property.Value)
}
