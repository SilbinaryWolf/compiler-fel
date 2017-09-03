package main

import (
	"encoding/json"
	"fmt"

	"github.com/silbinarywolf/compiler-fel/data"
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

	// Test VM implementation ideas
	/*{
		v := vm.New()
		configStructDefinition := v.NewStructDefinition(
			"Config",
			vm.StructDefinitionProperty{
				Name: "template_output_directory",
				Type: data.KindString,
			},
			vm.StructDefinitionProperty{
				Name: "css_output_directory",
				Type: data.KindString,
			},
			vm.StructDefinitionProperty{
				Name: "css_files",
				Type: data.KindString,
			},
		)
		//v.AddOpcodePushStruct("Config")

		//myVM.PrintDebug()
		v.AddOpcode(vm.Push, configStructDefinition)
		v.AddOpcode(vm.CallAllocStruct, nil)
		v.AddOpcode(vm.Push, "../public") // template_output_directory
		v.AddOpcode(vm.StoreInStruct, 0)

		/*v.AddOpcode(vm.Push, "www.google.com")
		v.AddOpcode(vm.Push, "url")
		v.AddOpcode(vm.Push, 1)
		v.AddOpcode(vm.CallPushNode, nil)
		v.AddOpcode(vm.Nil, nil)*/
		v.Run()
		json, _ := json.MarshalIndent(instance, "", "   ")
		fmt.Printf("%s\n", string(json))
		return
	}*/

	program := evaluator.New()
	err := program.RunProject("../testdata/sampleproject/fel")
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
}
