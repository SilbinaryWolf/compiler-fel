package evaluator

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/generate"
	"github.com/silbinarywolf/compiler-fel/parser"
	"strings"
	"testing"
)

func TestCSSOptimizeRule(t *testing.T, template string) {

}

func TestOptimizeCSS(t *testing.T) {
	p := parser.New()
	astFile := p.Parse([]byte(`
:: css {
	.exists {
		color: green;
	}

	.no-exists {
		color: red;
	}
}

div(class="exists") {
	"Test"
}`), "Layout.fel")
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
	if len(cssDefinitionSet) != 1 {
		htmlOutput := generate.PrettyHTMLComponentNode(node)
		t.Fatalf("Expected 1 CSS definition to be returned. Not %d.\n\nOutput HTML:\n%s", len(cssDefinitionSet), htmlOutput)
	}
	cssOutput := generate.PrettyCSS(cssDefinitionSet[0])
	if strings.Contains(cssOutput, ".no-exists") {
		t.Fatalf("Expected to optimize away unused .no-exists.\n\nOutput CSS:\n%s", cssOutput)
	}
}
