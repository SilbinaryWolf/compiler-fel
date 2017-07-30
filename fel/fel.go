package main

import (
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/evaluator"
	"github.com/silbinarywolf/compiler-fel/parser"
)

func main() {
	program := evaluator.New()
	err := program.RunProject("../testdata/sampleproject/fel")
	if err != nil {
		panic(err)
	}

	return

	// ----------------------
	// IGNORE BELOW
	// ------------------------
	p := parser.New()
	node, err := p.ParseFile("../testdata/sampleproject/fel/config.fel")
	if err != nil {
		panic(err)
	}
	errors := p.GetErrors()
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf(err.Error())
		}
		panic("errors!")
	}
	if node == nil {
		panic("no node")
	}
	// DEBUG
	json, _ := json.MarshalIndent(node, "", "   ")
	fmt.Printf("%s", string(json))
	panic("end of main")
}
