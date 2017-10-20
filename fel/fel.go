package main

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/evaluator"
)

func addNode() {
	fmt.Printf("test\n")
}

func main() {
	/*	{
		// Takes approx ~600ms on Windows machine
		start := time.Now()
		result, err := babelTranspileJavascript("let test = 1")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Code:\n%s", result)
		elapsed := time.Since(start)
		log.Printf("Binomial took %s", elapsed)
	}*/

	/*{
		// Does not support `let` keyword, ECMAScript 5.1+ only, not ECMAScript 2015.
		vm := goja.New()
		v, err := vm.RunScript("filename_here", `
			var result = 1
			result += 2
		`)
		if err != nil {
			panic(err)
		}
		num := v.Export().(int64)
		fmt.Printf("Result is: %d\n", num)
		panic("DONE!")
	}*/

	program := evaluator.New()
	err := program.RunProject("../testdata/sampleproject/fel")
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
}
