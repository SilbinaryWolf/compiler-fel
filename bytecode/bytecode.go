package bytecode

import (
	"bytes"
	"fmt"
)

type Kind int

const (
	Unknown Kind = 0 + iota
	Label
	Pop
	PopN
	Store
	StorePopStructField
	StorePopHTMLAttribute
	AppendPopHTMLNode
	StoreInternalStructField
	StoreAppendToArray
	StoreAppendToHTMLElement
	Push
	PushAllocArrayString
	PushAllocArrayInt
	PushAllocArrayFloat
	PushAllocArrayStruct
	PushStackVar
	PushStructFieldVar
	ReplaceStructFieldVar
	PushAllocStruct
	PushAllocInternalStruct
	PushAllocHTMLNode
	PopHTMLNode
	ConditionalEqual
	Add
	AddString
	Jump
	JumpIfFalse
	Call
	Return
)

/*type NodeContextType int

const (
	NodeUnknown NodeContextType = 0 + iota
	NodeCSSDefinition
)*/

var kindToString = []string{
	Unknown: "unknown/unset bytecode",
	Label:   "Label",
	Pop:     "Pop",
	PopN:    "PopN", // Pop [N] number of times
	Store:   "Store",
	StorePopHTMLAttribute:    "StorePopHTMLAttribute",
	StorePopStructField:      "StorePopStructField",
	AppendPopHTMLNode:        "AppendPopHTMLNode",
	StoreInternalStructField: "StoreInternalStructField",
	StoreAppendToArray:       "StoreAppendToArray",
	StoreAppendToHTMLElement: "StoreAppendToHTMLElement",
	Push:                    "Push",
	PushAllocArrayString:    "PushAllocArrayString",
	PushAllocArrayInt:       "PushAllocArrayInt",
	PushAllocArrayFloat:     "PushAllocArrayFloat",
	PushAllocArrayStruct:    "PushAllocArrayStruct",
	PushStackVar:            "PushStackVar",
	PushStructFieldVar:      "PushStructFieldVar",
	ReplaceStructFieldVar:   "ReplaceStructFieldVar",
	PushAllocStruct:         "PushAllocStruct",
	PushAllocInternalStruct: "PushAllocInternalStruct",
	PushAllocHTMLNode:       "PushAllocHTMLNode",
	PopHTMLNode:             "PopHTMLNode",
	ConditionalEqual:        "ConditionalEqual",
	Add:                     "Add",
	AddString:               "AddString",
	Jump:                    "Jump",
	JumpIfFalse:             "JumpIfFalse",
	Call:                    "Call",
	Return:                  "Return",
}

type BlockKind int

const (
	BlockUnknown BlockKind = 0 + iota
	BlockProcedure
	BlockHTMLComponentDefinition
)

type Code struct {
	Kind  Kind
	Value interface{}
}

func (kind Kind) String() string {
	return kindToString[kind]
}

func (code *Code) String() string {
	result := code.Kind.String()
	if code.Value != nil {
		switch code.Value.(type) {
		case string:
			result += fmt.Sprintf(" \"%v\"", code.Value)
		default:
			result += fmt.Sprintf(" %v", code.Value)
		}
	}
	return result
}

// ie. a function, block-scope, HTMLComponent
type Block struct {
	kind      BlockKind
	Opcodes   []Code
	StackSize int
}

func NewBlock(kind BlockKind) *Block {
	block := new(Block)
	block.kind = kind
	return block
}

func (block *Block) Kind() BlockKind {
	return block.kind
}

type StructInterface interface {
	GetField(index int) interface{}
	SetField(index int, value interface{})
}

type Struct struct {
	fields []interface{}
}

func NewStruct(fieldCount int) *Struct {
	structData := new(Struct)
	structData.fields = make([]interface{}, fieldCount)
	return structData
}

func (structData *Struct) SetField(index int, value interface{}) {
	structData.fields[index] = value
}

func (structData *Struct) GetField(index int) interface{} {
	return structData.fields[index]
}

func (structData *Struct) FieldCount() int {
	return len(structData.fields)
}

type HTMLElement struct {
	Name       string
	attributes []HTMLAttribute

	childNodes []interface{}

	parentNode   *HTMLElement
	previousNode *HTMLElement
	nextNode     *HTMLElement
}

type HTMLAttribute struct {
	Name  string
	Value string
}

func NewHTMLElement(tagName string) *HTMLElement {
	node := new(HTMLElement)
	node.Name = tagName
	return node
}

func (node *HTMLElement) SetParent(parent *HTMLElement) {
	if node.parentNode == parent {
		return
	}
	if node.parentNode != nil {
		// Remove from old parent
		for i, childNode := range node.parentNode.childNodes {
			childNode, ok := childNode.(*HTMLElement)
			if !ok {
				continue
			}
			if childNode == node {
				node.parentNode.childNodes = append(node.parentNode.childNodes[:i], node.parentNode.childNodes[i+1:]...)
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
	for i := 0; i < indent; i++ {
		buffer.WriteByte('\t')
	}
	buffer.WriteString(node.String())
	buffer.WriteByte('\n')
	if len(node.childNodes) > 0 {
		indent += 1
		for _, node := range node.childNodes {
			switch node := node.(type) {
			case *HTMLElement:
				buffer.WriteString(node.debugIndent(indent))
			default:
				panic(fmt.Sprintf("HTMLElement::Debug: Unhandled type %T", node))
			}
		}
		indent -= 1
		for i := 0; i < indent; i++ {
			buffer.WriteByte('\t')
		}
		buffer.WriteByte('<')
		buffer.WriteByte('/')
		buffer.WriteString(node.Name)
		buffer.WriteByte('>')
		buffer.WriteByte('\n')
	}
	return buffer.String()
}

func (node *HTMLElement) Debug() string {
	return node.debugIndent(0)
}

func (node *HTMLElement) String() string {
	var buffer bytes.Buffer
	buffer.WriteByte('<')
	buffer.WriteString(node.Name)
	buffer.WriteByte(' ')
	for i, attribute := range node.GetAttributes() {
		if i != 0 {
			buffer.WriteByte(' ')
		}
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
