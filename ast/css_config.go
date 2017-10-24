package ast

import (
	"github.com/silbinarywolf/compiler-fel/token"
	"path"
	"strings"
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
	Modify bool
}

type CSSConfigMatchPart []string

func initDefaultCSSConfigSettings(result *CSSConfigSettings) {
	result.Modify = true
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
			// todo(Jake): Make it so you need * for before/after matching
			//			   Doesn't do this currently due to path.Match()
			if strings.Contains(pattern, "*") {
				ok, err := path.Match(pattern, name)
				if err != nil {
					panic(err)
				}
				if ok {
					result = rule.CSSConfigSettings
					break
				}
				continue
			}
			// If no * in rule, assume it needs to be an exact match
			if pattern == name {
				result = rule.CSSConfigSettings
				break
			}
		}

	}
	return result
}
