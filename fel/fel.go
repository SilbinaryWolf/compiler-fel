package main

import (
	"github.com/silbinarywolf/compiler-fel/parser"
)

func main() {
	_, err := parser.ParseFile("../testdata/sampleproject/fel/config.fel")
	if err != nil {
		panic(err)
	}
	panic("end of main")
}
