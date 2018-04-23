package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/emitter"
	"github.com/silbinarywolf/compiler-fel/evaluator"
	"github.com/silbinarywolf/compiler-fel/parser"
	"github.com/silbinarywolf/compiler-fel/typer"
	"github.com/silbinarywolf/compiler-fel/vm"
)

type TemplateFile struct {
	ast    *ast.File
	code   *bytecode.Block
	output *data.HTMLElement
}

type CSSDefinition struct {
	ast  *ast.CSSDefinition
	code *bytecode.Block
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
	var emitTimeSpent time.Duration
	var executionTimeSpent time.Duration
	totalTimeSpentTimer := time.Now()

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
		{
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
			if p.HasErrors() {
				p.PrintErrors()
				return fmt.Errorf("Stopping due to parsing errors.")
			}
		}

		// Apply type information and typecheck when we've parsed all files
		var htmlComponentsUsed []*ast.HTMLComponentDefinition
		{
			typerSpentTimer := time.Now()
			p := typer.New()
			p.ApplyTypeInfoAndTypecheck(astFiles)
			typerTimeSpent += time.Since(typerSpentTimer)
			if p.HasErrors() {
				p.PrintErrors()
				return fmt.Errorf("Stopping due to parsing errors.")
			}
			htmlComponentsUsed = p.HTMLComponentsUsed()
		}

		// Emit bytecode
		codeRecords := make([]TemplateFile, 0, len(astFiles))
		cssDefinitionBlocks := make([]CSSDefinition, 0, 100)
		{
			emitSpentTimer := time.Now()
			emit := emitter.New()
			emit.EmitGlobalScope(astFiles)

			// Emit CSS
			for _, htmlDefinition := range htmlComponentsUsed {
				cssDef := htmlDefinition.CSSDefinition
				if cssDef == nil {
					continue
				}
				codeBlock := emit.EmitCSSDefinition(cssDef)
				cssDefinitionBlocks = append(cssDefinitionBlocks, CSSDefinition{
					ast:  htmlDefinition.CSSDefinition,
					code: codeBlock,
				})
			}

			// Emit template directories
			fmt.Printf("\n")
			for _, astFile := range astFiles {
				if !strings.HasPrefix(astFile.Filepath, templateInputDirectory) ||
					len(astFile.Nodes()) == 0 {
					continue
				}
				codeBlock := emit.EmitBytecode(astFile, emitter.FileOptions{
					IsTemplateFile: true,
				})

				codeRecords = append(codeRecords, TemplateFile{
					ast:  astFile,
					code: codeBlock,
				})
			}
			emitTimeSpent += time.Since(emitSpentTimer)
		}

		// Execute CSS code
		{
			var buffer bytes.Buffer
			executionSpentTimer := time.Now()
			for _, codeRecord := range cssDefinitionBlocks {
				result := vm.ExecuteNewProgram(codeRecord.code)
				switch result := result.(type) {
				case *data.CSSDefinition:
					//htmlElements = append(htmlElements, result)
					//fmt.Printf("Filename: %s\n%v\n\n", codeRecord.ast.Name.String(), result.Debug())
					buffer.WriteString("/* Filename: ")
					buffer.WriteString(codeRecord.ast.Name.String())
					buffer.WriteString("*/ \n")
					buffer.WriteString(result.Debug())
					buffer.WriteString("\n")
				case nil:
					panic(fmt.Sprintf("Unexpected type: nil"))
				default:
					panic(fmt.Sprintf("Unknown type: %T", result))
				}
			}
			cssOutput := buffer.String()
			executionTimeSpent += time.Since(executionSpentTimer)

			// Generate files
			cssDirpath := projectDirpath + workspace.CSSOutputDirectory() + "/"
			diskIOTimeSpentTimer := time.Now()
			for _, cssFilename := range workspace.CSSFiles() {
				// todo(Jake): 2018-04-23
				//
				// Fix permissions on this file write to a better default
				//
				ioutil.WriteFile(cssDirpath+cssFilename, []byte(cssOutput), 0744)
			}
			diskIOTimeSpent += time.Since(diskIOTimeSpentTimer)
			fmt.Printf("CSS Output:\n%s", cssOutput)
		}

		// Execute template code
		{
			executionSpentTimer := time.Now()
			for i, _ := range codeRecords {
				codeRecord := &codeRecords[i]
				result := vm.ExecuteNewProgram(codeRecord.code)
				switch result := result.(type) {
				case *data.HTMLElement:
					codeRecord.output = result
					fmt.Printf("Filename: %s\n%s\n", codeRecord.ast.Filepath, result.Debug())
				case nil:
					panic(fmt.Sprintf("Unexpected type: nil"))
				default:
					panic(fmt.Sprintf("Unknown type: %T", result))
				}
			}
			executionTimeSpent += time.Since(executionSpentTimer)

			//
			diskIOTimeSpentTimer := time.Now()
			for _, codeRecord := range codeRecords {
				filename := codeRecord.ast.Filepath
				htmlElement := codeRecord.output
				if htmlElement == nil {
					continue
				}

				baseFilename := filename[len(templateInputDirectory) : len(filename)-4]
				outputFilepath := filepath.Clean(fmt.Sprintf("%s%s.html", templateOutputDirectory, baseFilename))

				err := ioutil.WriteFile(
					outputFilepath,
					[]byte(htmlElement.Debug()),
					0644,
				)
				if err != nil {
					panic(err)
				}
			}
			diskIOTimeSpent += time.Since(diskIOTimeSpentTimer)
			/*result := TemplateFile{
				Filepath: outputFilepath,
				Content:  generate.PrettyHTML(nodes),
			}
			outputTemplateFileSet = append(outputTemplateFileSet, result)*/
		}

		fmt.Printf("\n")
		fmt.Printf("Building workspace #%d \"%s\"...\n", i, workspace.Name())
		fmt.Printf("Disk IO time: %s\n", diskIOTimeSpent)
		fmt.Printf("Parsing time: %s (Typer: %s)\n", parsingTimeSpent+typerTimeSpent, typerTimeSpent)
		fmt.Printf("Emitter time: %s\n", emitTimeSpent)
		fmt.Printf("Execution time: %s\n", executionTimeSpent)

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("Total time: %s (Memory used: %fmb)\n", time.Since(totalTimeSpentTimer), float32(m.TotalAlloc/1024)/100)
		//fmt.Printf("\nAlloc = %v\nTotalAlloc = %v\nSys = %v\nNumGC = %v\n\n", m.Alloc/1024, (m.TotalAlloc/1024)/100, m.Sys/1024, m.NumGC)
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
