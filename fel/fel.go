package main

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/evaluator"
)

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

	program := evaluator.New()
	err := program.RunProject("../testdata/sampleproject/fel")
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
}
