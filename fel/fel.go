package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/emitter"
	"github.com/silbinarywolf/compiler-fel/evaluator"
	"github.com/silbinarywolf/compiler-fel/parser"
	"github.com/silbinarywolf/compiler-fel/vm"
)

type Test struct {
	test int
}

func compileProject(projectDirpath string) error {
	configFilepath := projectDirpath + "config.fel"
	workspaces, err := evaluator.GetWorkspacesFromConfig(configFilepath)
	if err != nil {
		return err
	}
	if len(workspaces) == 0 {
		return fmt.Errorf("No workspaces found in config.fel file.")
	}

	var diskIOTimeSpent time.Duration
	var parsingTimeSpent time.Duration
	var typerTimeSpent time.Duration

	for i, workspace := range workspaces {
		templateInputDirectory := workspace.TemplateInputDirectory()
		if templateInputDirectory == "" {
			return fmt.Errorf("template_input_directory has not been configured.")
		}
		templateOutputDirectory := workspace.TemplateOutputDirectory()
		if templateOutputDirectory == "" {
			return fmt.Errorf("template_output_directory has not been configured.")
		}
		cssOutputDirectory := workspace.CSSOutputDirectory()
		if cssOutputDirectory == "" {
			return fmt.Errorf("css_output_directory has not been configured.")
		}

		templateInputDirectory = path.Clean(fmt.Sprintf("%s/%s", projectDirpath, templateInputDirectory))
		templateOutputDirectory = path.Clean(fmt.Sprintf("%s/%s", projectDirpath, templateOutputDirectory))
		cssOutputDirectory = path.Clean(fmt.Sprintf("%s/%s", projectDirpath, cssOutputDirectory))

		// Check if configured folders exist, create output folders automatically if it doesn't.
		//fmt.Printf("Creating output folders defined from \"config.fel\"...\n")
		//fmt.Printf("------------------------------------------\n")
		err = folderExistsMaybeCreate(templateInputDirectory, "template_input_directory", false)
		if err != nil {
			return err
		}
		err = folderExistsMaybeCreate(templateOutputDirectory, "template_output_directory", true)
		if err != nil {
			return err
		}
		err = folderExistsMaybeCreate(cssOutputDirectory, "css_output_directory", true)
		if err != nil {
			return err
		}

		// Get list of all files in folder recursively with *.fel
		filepathSet := make([]string, 0, 50)
		{
			diskIOTimeSpentTimer := time.Now()
			err := filepath.Walk(projectDirpath, func(path string, f os.FileInfo, _ error) error {
				if !f.IsDir() && filepath.Ext(f.Name()) == ".fel" {
					// Replace Windows-slash with forward slash.
					// NOTE: This ensures consistent comparison of filepath strings and fixes a bug.
					path = strings.Replace(path, "\\", "/", -1)
					filepathSet = append(filepathSet, path)
				}
				return nil
			})
			diskIOTimeSpent += time.Since(diskIOTimeSpentTimer)
			if err != nil {
				return fmt.Errorf("An error occurred reading: %v, Error Message: %v", templateInputDirectory, err)
			}
			if len(filepathSet) == 0 {
				return fmt.Errorf("No *.fel files found in your project's \"templates\" directory: %v", templateInputDirectory)
			}
		}

		// Parse files
		astFiles := make([]*ast.File, 0, 50)
		p := parser.New()
		for _, filepath := range filepathSet {
			diskIOTimeSpentTimer := time.Now()
			filecontentsAsBytes, err := ioutil.ReadFile(filepath)
			diskIOTimeSpent += time.Since(diskIOTimeSpentTimer)

			if err != nil {
				return fmt.Errorf("An error occurred reading file: %v, Error message: %v", filepath, err)
			}

			parseSpentTimer := time.Now()
			astFile := p.Parse(filecontentsAsBytes, filepath)
			parsingTimeSpent += time.Since(parseSpentTimer)

			if astFile == nil {
				if p.HasErrors() {
					p.PrintErrors()
				}
				return fmt.Errorf("Empty source file: %s.", filepath)
			}
			if p.Scanner.HasErrors() {
				p.PrintErrors()
				return fmt.Errorf("Stopping due to scanning errors.")
			}
			astFiles = append(astFiles, astFile)
		}

		// Typecheck when we've parsed all files
		{
			typerSpentTimer := time.Now()
			p := parser.NewTyper()
			p.TypecheckAndFinalize(astFiles)
			typerTimeSpent += time.Since(typerSpentTimer)
			if p.HasErrors() {
				p.PrintErrors()
				return fmt.Errorf("Stopping due to parsing errors.")
			}
		}

		// Emit bytecode
		emit := emitter.New()
		for _, astFile := range astFiles {
			// NOTE(Jake): 2018-03-16
			//
			// Not pulled out as dependencies aren't resolved properly yet
			//
			if len(astFile.Nodes()) == 0 {
				continue
			}
			emit.EmitGlobalScope(astFile)
		}

		// Emit template directories
		for _, astFile := range astFiles {
			if !strings.HasPrefix(astFile.Filepath, templateInputDirectory) ||
				len(astFile.Nodes()) == 0 {
				continue
			}
			codeBlock := emit.EmitBytecode(astFile, emitter.FileOptions{
				IsTemplateFile: true,
			})

			result := vm.ExecuteNewProgram(codeBlock)
			switch result := result.(type) {
			case *bytecode.HTMLElement:
				panic(result.Debug())
			case nil:
				// no-op
			default:
				panic(fmt.Sprintf("Unknown type: %T", result))
			}
		}

		fmt.Printf("\n")
		fmt.Printf("Building workspace #%d \"%s\"...\n", i, workspace.Name())
		fmt.Printf("Disk IO time: %s\n", diskIOTimeSpent)
		fmt.Printf("Parsing time: %s (Typer: %s)\n", parsingTimeSpent+typerTimeSpent, typerTimeSpent)

	}
	return nil
}

func folderExistsMaybeCreate(directory string, configName string, createIfDoesntExist bool) error {
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		if !createIfDoesntExist {
			return fmt.Errorf("%s: does not exist: %s", configName, directory)
		}
		fmt.Printf("%s: Creating missing folder \"%s\"\n", configName, directory)
		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("%s: error: %v", configName, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: OS error: %v", configName, err)
	}
	return nil
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

	compileProject("testdata/sampleproject/fel/")

	/*program := evaluator.New()
	err := program.RunProject("testdata/sampleproject/fel")
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}*/
	fmt.Printf("Done.")
}
