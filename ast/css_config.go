package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
	"path"
)

type CSSConfigDefinition struct {
	Name  token.Token
	Rules []*CSSConfigRule
}

func (topNode *CSSConfigDefinition) Nodes() []Node {
	return nil
}

type CSSConfigRule struct {
	Selectors          []CSSConfigMatchPart
	SelectorsAsPattern []string
	CSSConfigSettings
}

type CSSConfigSettings struct {
	ModifyName bool
}

type CSSConfigMatchPart []string

func initDefaultCSSConfigSettings(result *CSSConfigSettings) {
	result.ModifyName = true
}

func NewCSSConfigRule() *CSSConfigRule {
	result := new(CSSConfigRule)
	initDefaultCSSConfigSettings(&result.CSSConfigSettings)
	return result
}

func (cssConfigDefinition *CSSConfigDefinition) GetSettings(name string) CSSConfigSettings {
	result := CSSConfigSettings{}
	initDefaultCSSConfigSettings(&result)
	if cssConfigDefinition == nil {
		return result
	}

	for _, rule := range cssConfigDefinition.Rules {
		for _, pattern := range rule.SelectorsAsPattern {
			//pattern := ""
			//for _, part := range selector {
			//	pattern += part
			//}
			ok, err := path.Match(pattern, name)
			if err != nil {
				panic(err)
			}
			if ok {
				result = rule.CSSConfigSettings
				break
			}
		}

	}
	return result
}
