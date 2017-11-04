package evaluator

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
)

func optimizeRules(definition *data.CSSDefinition, htmlNodeInfo *HTMLComponentNodeInfo, cssConfigDefinition *ast.CSSConfigDefinition, onlyScanCurrentHTMLComponentScope bool) {
	for ruleIndex := 0; ruleIndex < len(definition.ChildNodes); ruleIndex++ {
		cssRule := definition.ChildNodes[ruleIndex]

		// If no properties on rule, remove it completely.
		if len(cssRule.Properties) == 0 {
			definition.ChildNodes = append(definition.ChildNodes[:ruleIndex], definition.ChildNodes[ruleIndex+1:]...)
			ruleIndex--
			continue
		}

	SelectorLoop:
		for selectorIndex := 0; selectorIndex < len(cssRule.Selectors); selectorIndex++ {
			selector := cssRule.Selectors[selectorIndex]

			// If part of a selector has "modify: false" rule, do not optimize
			// this selector away.
			// ie. Keeping "js-my-hook", "is-active", "active"
			for _, part := range selector {
				partString := part.String()
				config := cssConfigDefinition.GetSettings(partString)
				// Don't optimize away if specifically flagged to not modify.
				if !config.Modify {
					continue SelectorLoop
				}
			}

			/*{
				res := ""
				for _, htmlNode := range htmlNodeInfo.Nodes {
					res += generate.PrettyHTML([]data.Type{htmlNode}) + "\n"
				}
				panic(fmt.Sprintf("%s\nDone %d nodes for \"%s\"", res, len(htmlNodeInfo.Nodes), htmlNodeInfo.HTMLDefinition.Name))
			}*/

			// Check for matches
			nodesMatchedCount := 0
			for _, htmlNode := range htmlNodeInfo.Nodes {
				nodesMatched := htmlNode.HTMLNode.QuerySelectorAllWithOptions(selector, onlyScanCurrentHTMLComponentScope)
				nodesMatchedCount += len(nodesMatched)
			}
			if nodesMatchedCount == 0 {
				// Remove if no match
				cssRule.Selectors = append(cssRule.Selectors[:selectorIndex], cssRule.Selectors[selectorIndex+1:]...)
				selectorIndex--
				continue SelectorLoop
			}
			// If found a match, stop looking for matches with this
			// selector
			continue SelectorLoop
		}

		// If no selectors (ie. removed all the ones that didnt match, remove this rule)
		if len(cssRule.Selectors) == 0 {
			definition.ChildNodes = append(definition.ChildNodes[:ruleIndex], definition.ChildNodes[ruleIndex+1:]...)
			ruleIndex--
			continue
		}
	}
}

func (program *Program) evaluateOptimizeAndReturnUsedCSS() []*data.CSSDefinition {
	outputCSSDefinitionSet := make([]*data.CSSDefinition, 0, 3)

	// Output named "MyComponent :: css" blocks
	for _, htmlNodeInfo := range program.htmlDefinitionUsed {
		htmlDefinition := htmlNodeInfo.HTMLDefinition
		if htmlDefinition == nil {
			panic("Unexpected error. HTMLNodeSet should always have a HTMLDefinition.")
		}

		cssDefinition := htmlDefinition.CSSDefinition
		if cssDefinition == nil {
			continue
		}

		// Process CSSDefinition
		program.currentComponentScope = append(program.currentComponentScope, htmlNodeInfo.HTMLDefinition)
		dataCSSDefinition := program.evaluateCSSDefinition(cssDefinition, program.globalScope)
		program.currentComponentScope = program.currentComponentScope[:len(program.currentComponentScope)-1]

		// Debug
		//{
		//	json, _ := json.MarshalIndent(htmlNodeInfo, "", "   ")
		//	fmt.Printf("%s", string(json))
		//	panic("ROY")
		//}

		// Optimize
		optimizeRules(dataCSSDefinition, htmlNodeInfo, htmlDefinition.CSSConfigDefinition, true)

		// Add output
		if len(dataCSSDefinition.ChildNodes) > 0 {
			outputCSSDefinitionSet = append(outputCSSDefinitionSet, dataCSSDefinition)
		}
	}

	// Output anonymous ":: css" blocks
	htmlNodeInfo := new(HTMLComponentNodeInfo)
	for _, itHtmlNodeInfo := range program.htmlTemplatesUsed {
		// NOTE: Packing each seperate template into this so `optimizeRules`
		//		 code can be reused easily.
		htmlNodeInfo.Nodes = append(htmlNodeInfo.Nodes, itHtmlNodeInfo.Nodes...)
	}
	for _, cssDefinition := range program.anonymousCSSDefinitionsUsed {
		dataCSSDefinition := program.evaluateCSSDefinition(cssDefinition, program.globalScope)

		optimizeRules(dataCSSDefinition, htmlNodeInfo, nil, false)

		// Add output
		if len(dataCSSDefinition.ChildNodes) > 0 {
			outputCSSDefinitionSet = append(outputCSSDefinitionSet, dataCSSDefinition)
		}

		// todo(Jake): Get all templates, iterate over each
		//			   and optimize away anonymous CSS
	}

	return outputCSSDefinitionSet
}
