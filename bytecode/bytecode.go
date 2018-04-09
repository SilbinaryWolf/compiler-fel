package bytecode

import (
	"bytes"
	"fmt"
	"github.com/silbinarywolf/compiler-fel/types"
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
	AppendPopHTMLNodeReturn
	StoreInternalStructField
	AppendPopArrayString
	AppendPopHTMLElementToHTMLElement
	CastToHTMLText
	Push
	PushAllocArrayString
	PushAllocArrayInt
	PushAllocArrayFloat
	PushAllocArrayStruct
	PushAllocHTMLFragment
	PushStackVar
	PushStructFieldVar
	PushReturnHTMLNodeArray
	ReplaceStructFieldVar
	PushAllocStruct
	PushAllocInternalStruct
	PushAllocHTMLNode
	ConditionalEqual
	Add
	AddString
	Jump
	JumpIfFalse
	Call
	CallHTML
	Return
)

type HTMLKind int

const (
	HTMLKindUnknown HTMLKind = 0 + iota
	HTMLKindElement
	HTMLKindText
	HTMLKindFragment
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
	StorePopHTMLAttribute:             "StorePopHTMLAttribute",
	StorePopStructField:               "StorePopStructField",
	AppendPopHTMLNodeReturn:           "AppendPopHTMLNodeReturn",
	StoreInternalStructField:          "StoreInternalStructField",
	AppendPopArrayString:              "AppendPopArrayString",
	AppendPopHTMLElementToHTMLElement: "AppendPopHTMLElementToHTMLElement",
	CastToHTMLText:                    "CastToHTMLText",
	Push:                              "Push",
	PushAllocArrayString:              "PushAllocArrayString",
	PushAllocArrayInt:                 "PushAllocArrayInt",
	PushAllocArrayFloat:               "PushAllocArrayFloat",
	PushAllocArrayStruct:              "PushAllocArrayStruct",
	PushAllocHTMLFragment:             "PushAllocHTMLFragment",
	PushReturnHTMLNodeArray:           "PushReturnHTMLNodeArray",
	PushStackVar:                      "PushStackVar",
	PushStructFieldVar:                "PushStructFieldVar",
	ReplaceStructFieldVar:             "ReplaceStructFieldVar",
	PushAllocStruct:                   "PushAllocStruct",
	PushAllocInternalStruct:           "PushAllocInternalStruct",
	PushAllocHTMLNode:                 "PushAllocHTMLNode",
	ConditionalEqual:                  "ConditionalEqual",
	Add:                               "Add",
	AddString:                         "AddString",
	Jump:                              "Jump",
	JumpIfFalse:                       "JumpIfFalse",
	Call:                              "Call",
	CallHTML:                          "CallHTML",
	Return:                            "Return",
}

type BlockKind int

const (
	BlockDefault BlockKind = 0 + iota
	BlockTemplate
	BlockProcedure
	BlockHTMLComponentDefinition
	BlockWorkspaceDefinition
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
	kind           BlockKind
	Name           string // procedure name / workspace name / etc
	Opcodes        []Code
	StackSize      int
	HasReturnValue bool
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
