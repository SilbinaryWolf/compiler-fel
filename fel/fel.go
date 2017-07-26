package main

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/parser"
)

func main() {
	p := parser.New()
	p.ParseFile("../testdata/sampleproject/fel/config.fel")
	if p.HasError() {
		for _, err := range p.GetErrors() {
			fmt.Printf(err.Error())
		}
		panic("errors!")
	}
	panic("end of main")
}
