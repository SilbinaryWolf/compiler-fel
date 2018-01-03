package parser

// NOTE(Jake): 2018-01-03
//
// 1) cd to `parser` directory
// 2) Run `go test -bench=Benchmark -benchtime=1s`
//

import (
	"testing"

	"github.com/silbinarywolf/compiler-fel/ast"
)

var basicComponentTest = `
Layout :: css {
	html, 
	body {
		padding: 0
	}

	.class-to-optimize-out {
		height: 0
	}
}

Layout :: html {
	:: struct {
		body_class := ""
	}

	html(lang="en-AU") {
		head {
			meta(charset="utf-8")
			meta(http-equiv="X-UA-Compatible", content="IE=edge")
			meta(name="viewport", content="width=device-width, height=device-height, initial-scale=1.0, user-scalable=0, minimum-scale=1.0, maximum-scale=1.0")
			title {
				"My website"
			}
			link(rel="stylesheet", type="text/css", href="../css/main.css")
		}
		body(class="no-js "+body_class) {
			Header(isBlue=false)
			children
		}
	}
}
`

var homePageTemplate = `
:: css {
	.exists {
		color: green
	}

	.dont-exist {
		color: red
	}
}

Layout(body_class="HomePage") {
	div(class="only-if-true") {
	}
	div(class="exists") {
		"Test"
	}
}

`

func BenchmarkParser(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p := New()
		parseString(b, p, basicComponentTest)
	}
}

func BenchmarkParserAndTypecheck(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p := New()
		astFiles := make([]*ast.File, 0, 10)
		astFiles = append(astFiles, parseString(b, p, basicComponentTest))
		// NOTE(Jake): 2018-01-03
		//
		// Test with multiple template files to see how it affects typechecking
		// speed. Might eventually want this in a seperate function so we can compare
		// with 4, 10, 15, 30, 100 template files.
		//
		for i := 0; i < 4; i++ {
			astFiles = append(astFiles, parseString(b, p, homePageTemplate))
		}
		p.TypecheckAndFinalize(astFiles)
		if p.HasErrors() {
			p.PrintErrors()
			b.Fatalf("Typechecker has hit errors.")
		}
	}
}

func parseString(b *testing.B, p *Parser, template string) *ast.File {
	astFile := p.Parse([]byte(template), "DummyFilename.fel")
	if astFile == nil {
		b.Fatalf("Unable to parse provided template.")
	}
	return astFile
}
