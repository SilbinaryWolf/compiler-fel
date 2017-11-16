package parser

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/token"
	"github.com/silbinarywolf/compiler-fel/types"
)

type TypeInfoManager struct {
	registeredTypes map[string]TypeInfo

	// built-in
	intInfo TypeInfo_Int
}

func (manager *TypeInfoManager) Init() {
	if manager.registeredTypes != nil {
		panic("Cannot initialize TypeInfoManager twice.")
	}
	manager.registeredTypes = make(map[string]TypeInfo)
	manager.register("int", manager.NewTypeInfoInt())
	manager.register("string", types.String())
	manager.register("float", types.Float())
	manager.register("bool", types.Bool())
}

func (manager *TypeInfoManager) register(name string, typeInfo TypeInfo) {
	_, ok := manager.registeredTypes[name]
	if ok {
		panic(fmt.Sprintf("Already registered \"%s\" type.", name))
	}
	manager.registeredTypes[name] = typeInfo
}

func (manager *TypeInfoManager) get(name string) TypeInfo {
	return manager.registeredTypes[name]
}

type TypeInfo interface {
	String() string
	Create() data.Type
}

// Int
type TypeInfo_Int struct{}

func (info *TypeInfo_Int) String() string    { return "int" }
func (info *TypeInfo_Int) Create() data.Type { return new(data.Integer64) }

func (manager *TypeInfoManager) NewTypeInfoInt() *TypeInfo_Int {
	return &manager.intInfo
}

// Functions
func (p *Parser) parseType() ast.Type {
	result := ast.Type{}

	t := p.GetNextToken()
	if t.Kind == token.BracketOpen {
		// Parse array / array-of-array / etc
		// ie. []string, [][]string, [][][]string, etc
		result.ArrayDepth = 1
		for {
			t = p.GetNextToken()
			if t.Kind != token.BracketClose {
				p.addErrorToken(p.expect(t, token.BracketClose), t)
				return result
			}
			t = p.GetNextToken()
			if t.Kind == token.BracketOpen {
				result.ArrayDepth++
				continue
			}
			break
		}
	}
	if t.Kind != token.Identifier {
		p.addErrorToken(p.expect(t, token.Identifier), t)
		return result
	}
	result.Name = t
	return result
}

func (p *Parser) DetermineType(node *ast.Type) types.TypeInfo {
	var resultType types.TypeInfo

	str := node.Name.String()
	resultType = p.typeinfo.get(str)
	if resultType == nil {
		p.addErrorToken(fmt.Errorf("Undeclared type \"%s\".", str), node.Name)
	}
	if node.ArrayDepth > 0 {
		for i := 0; i < node.ArrayDepth; i++ {
			arrayItemType := resultType
			resultType = types.Array(arrayItemType)
		}
	}
	return resultType
}
