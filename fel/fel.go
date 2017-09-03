package main

import (
	"fmt"
	"github.com/silbinarywolf/compiler-fel/evaluator"
	"github.com/silbinarywolf/compiler-fel/vm"
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

	{
		v := vm.New()
		//myVM.PrintDebug()
		v.AddOpcode(vm.Push, "www.google.com")
		v.AddOpcode(vm.Push, "url")
		v.AddOpcode(vm.Push, 1)
		v.AddOpcode(vm.CallPushNode, nil)
		v.AddOpcode(vm.Nil, nil)
		v.Run()
		return
	}

	program := evaluator.New()
	err := program.RunProject("../testdata/sampleproject/fel")
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
}
