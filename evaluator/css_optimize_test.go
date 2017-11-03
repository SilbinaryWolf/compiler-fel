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
		:: css {
			.exists {
				color: green;
			}
			.no-exists {
				color: red;
			}
		}
		div(class="exists") {
		}
	`, []string{
		".exists",
	}, []string{
		".no-exists",
	})

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

		MyComponent() {
		}
	`, []string{
		".MyComponent__exists-2",
	}, []string{
		".MyComponent__no-exists-2",
	})
}

func TestOptimizeCSSTagName(t *testing.T) {
	// NOTE: This doesn't properly test to ensure
	//		 the component <div> exists.
	//		 - 2017-11-03
	CSSOptimizeRuleCheck(t, `
		:: css {
			div {
				color: green;
			}
			a {
				color: red;
			}
		}

		div() {}
	`, []string{
		"div",
	}, []string{
		"a",
	})

	CSSOptimizeRuleCheck(t, `
		MyComponent :: css {
			div {
				color: green;
			}
			a {
				color: red;
			}
		}

		MyComponent :: html {
			div() {}
		}

		MyComponent() {}
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
				color: green;
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

	CSSOptimizeRuleCheck(t, `
		MyComponent :: css {
			.sib ~ .sib {
				color: green;
			}
			.sib > .sib {
				color: red;
			}
		}

		MyComponent :: html {
			div(class="sib") {}
			div(class="sib") {}
		}
		MyComponent() {}
	`, []string{
		".MyComponent__sib~.MyComponent__sib",
	}, []string{
		".MyComponent__sib>.MyComponent__sib",
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

	CSSOptimizeRuleCheck(t, `
		MyComponent :: css {
			.p + .p {
				color: green;
			}
			.div + .div {
				color: red;
			}
		}

		MyComponent :: html {
			div(class="div") {
				p(class="p") {}
				p(class="p") {}
			}
			section() {}
			div(class="div") {}
		}
		
		MyComponent(){}
	`, []string{
		".MyComponent__p+.MyComponent__p",
	}, []string{
		".MyComponent__div+MyComponent__div",
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

	// If test failed, print out errors and output so the issue can be diagnosed
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
