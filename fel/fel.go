package main

import (
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/parser"
)

func main() {
	p := parser.New()
	node := p.ParseFile("../testdata/sampleproject/fel/config.fel")
	if p.HasError() {
		for _, err := range p.GetErrors() {
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
