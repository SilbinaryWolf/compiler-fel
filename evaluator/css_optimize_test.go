package evaluator

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/generate"
	"github.com/silbinarywolf/compiler-fel/parser"
	"strings"
	"testing"
)

func TestOptimizeCSSClass(t *testing.T) {
	// TODO(JAKE): Add test to ensure if no "children"
	//  		   keyword exists, that the typechecker
	//			   will throw an error.

	// todo(Jake): Fix bug where if the following comment exists below
	//			   scanner gets an error.

	//
	//
	// children

	CSSOptimizeRuleCheck(t, `
		MyComponent :: css {
			.exists-2 {
				color: green;
			}
			.no-exists-2 {
				color: red;
			}
		}

		MyComponent :: html {
			div(class="exists-2") {
				// children
			}
		}

		:: css {
			.exists {
				color: green;
			}
			.no-exists {
				color: red;
			}
		}
		div(class="exists") {
			MyComponent() {
			}
		}
	`, []string{
		".exists",
		".MyComponent__exists-2",
	}, []string{
		".no-exists",
		".MyComponent__no-exists-2",
	})
}

func TestOptimizeCSSTagName(t *testing.T) {
	CSSOptimizeRuleCheck(t, `
		:: css {
			a {
				color: red;
			}
			div {
				color: blue;
			}
		}

		div() {}
	`, []string{
		"div",
	}, []string{
		"a",
	})
}

func TestOptimizeCSSSibling(t *testing.T) {
	CSSOptimizeRuleCheck(t, `
		:: css {
			div ~ div {
				color: blue;
			}
			div > div {
				color: red;
			}
		}

		div() {}
		div() {}
	`, []string{
		"div~div",
	}, []string{
		"div>div",
	})
}

func TestOptimizeCSSAdjacent(t *testing.T) {
	CSSOptimizeRuleCheck(t, `
		:: css {
			p + p {
				color: blue;
			}
			div + div {
				color: red;
			}
		}

		div() {
			p() {}
			p() {}
		}
		section() {}
		div() {}
	`, []string{
		"p+p",
	}, []string{
		"div+div",
	})
}

func CSSOptimizeRuleCheck(t *testing.T, template string, successContainsList []string, failContainsList []string) {
	p := parser.New()
	astFile := p.Parse([]byte(template), "Layout.fel")
	if astFile == nil {
		t.Fatalf("p.Parse should not return nil.")
	}
	p.TypecheckAndFinalize([]*ast.File{astFile})
	if p.HasErrors() {
		p.PrintErrors()
		t.Fatalf("Stopping due to scanning/parsing errors.")
	}
	program := New()
	node, err := program.evaluateTemplate(astFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	cssDefinitionSet := program.evaluateOptimizeAndReturnUsedCSS()
	if len(cssDefinitionSet) == 0 {
		htmlOutput := generate.PrettyHTMLComponentNode(node)
		t.Fatalf("Expected at least 1 CSS definition to be returned. Not %d.\n\nOutput HTML:\n%s", len(cssDefinitionSet), htmlOutput)
	}

	cssOutput := ""
	for _, cssDefinition := range cssDefinitionSet {
		cssOutput += generate.PrettyCSS(cssDefinition) + "\n"
	}

	//
	outputCSSWithFatal := false
	for _, successContains := range successContainsList {
		if !strings.Contains(cssOutput, successContains) {
			t.Errorf("Expected to keep used rule \"%s\".\n", successContains)
			outputCSSWithFatal = true
		}
	}
	for _, failContains := range failContainsList {
		if strings.Contains(cssOutput, failContains) {
			t.Errorf("Expected to remove unused rule \"%s\".\n", failContains)
			outputCSSWithFatal = true
		}
	}
	if outputCSSWithFatal {
		htmlOutput := generate.PrettyHTML([]data.Type{node})
		t.Fatalf("\nCSS:\n%s\n\nHTML:\n%s", cssOutput, htmlOutput)
	}
}
