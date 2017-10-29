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

	/*{
		ctx := duktape.New()
		ctx.PevalString(`let result = 2 + 3; return result;`)
		result := ctx.GetNumber(-1)
		ctx.Pop()
		fmt.Println("result is:", result)

		// To prevent memory leaks, don't forget to clean up after
		// yourself when you're done using a context.
		//ctx.DestroyHeap()

		//panic("Done!")
	}*/

	program := evaluator.New()
	err := program.RunProject("../testdata/sampleproject/fel")
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
}
