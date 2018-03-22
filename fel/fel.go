package main

import (
	"fmt"

	"github.com/silbinarywolf/compiler-fel/evaluator"
)

type Test struct {
	test int
}

func main() {
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

	workspaces, err := evaluator.GetWorkspacesFromConfig("testdata/sampleproject/fel/config.fel")
	if err != nil {
		panic(err)
	}
	if len(workspaces) == 0 {
		panic(fmt.Errorf("No workspaces found in config.fel file."))
	}

	for i, workspace := range workspaces {
		fmt.Printf("Workspace #%d - %v", i, workspace)
	}

	/*program := evaluator.New()
	err := program.RunProject("testdata/sampleproject/fel")
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}*/
	fmt.Printf("Done.")
}
