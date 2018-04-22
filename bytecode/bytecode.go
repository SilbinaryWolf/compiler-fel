package bytecode

import (
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
	AppendPopHTMLNodeReturn
	StoreInternalStructField
	AppendPopArrayString
	AppendPopHTMLElementToHTMLElement
	AppendCSSPropertyToCSSRule
	CastToHTMLText
	Push

	// Array Structures
	PushAllocArrayString
	PushAllocArrayInt
	PushAllocArrayFloat
	PushAllocArrayStruct
	PushAllocHTMLFragment

	// CSS Structure
	PushAllocCSSDefinition
	//PushAllocCSSSelector
	//PushAllocCSSSelectorPart
	PushAllocCSSRule

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
	AppendCSSPropertyToCSSRule:        "AppendCSSPropertyToCSSRule",
	CastToHTMLText:                    "CastToHTMLText",
	Push:                              "Push",
	// Array Structures
	PushAllocArrayString:  "PushAllocArrayString",
	PushAllocArrayInt:     "PushAllocArrayInt",
	PushAllocArrayFloat:   "PushAllocArrayFloat",
	PushAllocArrayStruct:  "PushAllocArrayStruct",
	PushAllocHTMLFragment: "PushAllocHTMLFragment",
	// CSS Structures
	PushAllocCSSDefinition:  "PushAllocCSSDefinition",
	PushAllocCSSRule:        "PushAllocCSSRule",
	PushReturnHTMLNodeArray: "PushReturnHTMLNodeArray",
	PushStackVar:            "PushStackVar",
	PushStructFieldVar:      "PushStructFieldVar",
	ReplaceStructFieldVar:   "ReplaceStructFieldVar",
	PushAllocStruct:         "PushAllocStruct",
	PushAllocInternalStruct: "PushAllocInternalStruct",
	PushAllocHTMLNode:       "PushAllocHTMLNode",
	ConditionalEqual:        "ConditionalEqual",
	Add:                     "Add",
	AddString:               "AddString",
	Jump:                    "Jump",
	JumpIfFalse:             "JumpIfFalse",
	Call:                    "Call",
	CallHTML:                "CallHTML",
	Return:                  "Return",
}

type BlockKind int

const (
	BlockDefault BlockKind = 0 + iota
	BlockUnresolved
	BlockTemplate
	BlockProcedure
	BlockHTMLComponentDefinition
	BlockWorkspaceDefinition
	BlockCSSDefinition
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
	name           string // procedure name / workspace name / etc
	kind           BlockKind
	isUnresolved   bool
	Opcodes        []Code
	StackSize      int
	HasReturnValue bool
}

func (block *Block) Name() string { return block.name }

func NewBlock(name string, kind BlockKind) *Block {
	block := new(Block)
	block.name = name
	block.kind = kind
	return block
}

func NewUnresolvedBlock(name string, kind BlockKind) *Block {
	block := new(Block)
	block.name = name
	block.kind = kind
	block.isUnresolved = true
	return block
}

func (block *Block) Kind() BlockKind {
	return block.kind
}

/*func (block *Block) DebugOpcodes(offset int) {
	fmt.Printf("Opcode Debug:\n-----------\n")
	for i, code := range opcodes {
		if offset == i {
			fmt.Printf("**%d** - %s\n", i, code.String())
			continue
		}
		fmt.Printf("%d - %s\n", i, code.String())
	}
	fmt.Printf("-----------\n")
}
*/
