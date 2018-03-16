package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/bytecode"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/emitter"
	"github.com/silbinarywolf/compiler-fel/generate"
	"github.com/silbinarywolf/compiler-fel/parser"
	"github.com/silbinarywolf/compiler-fel/vm"
)

/*func (program *Program) CreateDataType(t token.Token) data.Type {
	typename := t.String()
	switch typename {
	case "string":
		return new(data.String)
	case "html_node":
		var empty *data.HTMLNode
		return empty
	default:
		panic(fmt.Sprintf("Unknown type name: %s", typename))
	}
}*/

func (program *Program) GetConfigString(configName string) (string, error) {
	value, ok := program.globalScope.Get(configName)
	if !ok {
		return "", fmt.Errorf("%s is undefined in config.fel. This definition is required.", configName)
	}
	valueAsserted, ok := value.(*data.String)
	if !ok {
		return "", fmt.Errorf("%s is expected to be a string.", configName)
	}
	return valueAsserted.String(), nil
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

func (program *Program) RunProject(projectDirpath string) error {
	totalTimeStart := time.Now()

	configFilepath := projectDirpath + "/config.fel"
	if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
		return fmt.Errorf("Cannot find config.fel in root of project directory: %v", configFilepath)
	}

	// Find and parse config.fel
	var configAstFile *ast.File
	var readFileTime time.Duration

	{
		filepath := configFilepath

		p := parser.New()

		fileReadStart := time.Now()
		filecontentsAsBytes, err := ioutil.ReadFile(filepath)
		readFileTime += time.Since(fileReadStart)

		if err != nil {
			return fmt.Errorf("An error occurred reading file: %v, Error message: %v", filepath, err)
		}
		astFile := p.Parse(filecontentsAsBytes, filepath)
		if astFile == nil {
			if p.HasErrors() {
				p.PrintErrors()
			}
			return fmt.Errorf("Parse errors in config.fel in root of project directory")
		}
		configAstFile = astFile
		p.TypecheckFile(configAstFile, nil)
		if p.HasErrors() {
			p.PrintErrors()
			return fmt.Errorf("Parse errors in config.fel in root of project directory")
		}
		if configAstFile == nil {
			return fmt.Errorf("Cannot find config.fel in root of project directory: %v", projectDirpath)
		}
	}

	// Evaluate config file
	for _, node := range configAstFile.Nodes() {
		program.evaluateStatement(node, program.globalScope)
	}
	//panic("Finished evaluating config file")

	// Get config variables

	templateInputDirectory, err := program.GetConfigString("template_input_directory")
	if err != nil {
		return err
	}
	templateInputDirectory = path.Clean(fmt.Sprintf("%s/%s", projectDirpath, templateInputDirectory))

	templateOutputDirectory, err := program.GetConfigString("template_output_directory")
	if err != nil {
		return err
	}
	templateOutputDirectory = path.Clean(fmt.Sprintf("%s/%s", projectDirpath, templateOutputDirectory))

	cssOutputDirectory, err := program.GetConfigString("css_output_directory")
	if err != nil {
		return err
	}
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
	fmt.Printf("\n")

	// Get all files in folder recursively with *.fel
	filepathSet := make([]string, 0, 50)
	{
		fileReadStart := time.Now()
		err := filepath.Walk(projectDirpath, func(path string, f os.FileInfo, _ error) error {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".fel" {
				// Replace Windows-slash with forward slash.
				// NOTE: This ensures consistent comparison of filepath strings and fixes a bug.
				path = strings.Replace(path, "\\", "/", -1)
				filepathSet = append(filepathSet, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("An error occurred reading: %v, Error Message: %v", templateInputDirectory, err)
		}

		if len(filepathSet) == 0 {
			return fmt.Errorf("No *.fel files found in your project's \"templates\" directory: %v", templateInputDirectory)
		}
		readFileTime += time.Since(fileReadStart)
	}

	// Parse files
	astFiles := make([]*ast.File, 0, 50)
	p := parser.New()
	parsingStart := time.Now()
	for _, filepath := range filepathSet {
		fileReadStart := time.Now()
		filecontentsAsBytes, err := ioutil.ReadFile(filepath)
		readFileTime += time.Since(fileReadStart)
		if err != nil {
			return fmt.Errorf("An error occurred reading file: %v, Error message: %v", filepath, err)
		}
		astFile := p.Parse(filecontentsAsBytes, filepath)
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
	p.TypecheckAndFinalize(astFiles)
	parsingElapsed := time.Since(parsingStart)
	if p.HasErrors() {
		p.PrintErrors()
		return fmt.Errorf("Stopping due to parsing errors.")
	}

	type TemplateFile struct {
		Filepath string
		Content  string
	}

	outputTemplateFileSet := make([]TemplateFile, 0, len(astFiles))

	// EXPERIMENTAL: Bytecode
	{
		emit := emitter.New()
		for _, astFile := range astFiles {
			if strings.Contains(astFile.Filepath, "Header.fel") {
				if len(astFile.Nodes()) == 0 {
					panic("Missing parsed nodes from file.")
				}

				json, _ := json.MarshalIndent(astFile, "", "   ")
				fmt.Printf("%s\nJSON AST\n---------------\n", string(json))

				// NOTE(Jake): 2018-03-16
				//
				// Not pulled out as dependencies aren't resolved properly yet
				//
				emit.EmitGlobalScope(astFile)
				codeBlock := emit.EmitBytecode(astFile, emitter.FileOptions{
					IsTemplateFile: true,
				})

				// TEST: Workspace
				for _, workspaceCode := range emit.Workspaces() {
					result := vm.ExecuteNewProgram(workspaceCode)
					panic(fmt.Sprintf("Workspace type: %T", result))
				}
				panic(fmt.Sprintf("workspace to test, count: %d", len(emit.Workspaces())))

				// TEST: Running bytecode of template
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
		}
		panic("Finished emitBytecode in Evaluator")
	}

	// Execute template
	executionStart := time.Now()
	for _, astFile := range astFiles {
		if !strings.HasPrefix(astFile.Filepath, templateInputDirectory) {
			continue
		}
		componentNode, err := program.evaluateTemplate(astFile)
		if err != nil {
			return fmt.Errorf("File %s\n- %v", astFile.Filepath, err)
		}
		nodes := componentNode.Nodes()
		if len(nodes) == 0 {
			return fmt.Errorf("File %s\n- No top-level HTMLNode or HTMLText found.\n\nStopping due to errors.", astFile.Filepath)
		}

		// Get filepath
		baseFilename := astFile.Filepath[len(templateInputDirectory) : len(astFile.Filepath)-4]
		outputFilepath := filepath.Clean(fmt.Sprintf("%s%s.html", templateOutputDirectory, baseFilename))
		result := TemplateFile{
			Filepath: outputFilepath,
			Content:  generate.PrettyHTML(nodes),
		}
		outputTemplateFileSet = append(outputTemplateFileSet, result)
	}

	// Output named "MyComponent :: css" blocks
	outputCSSDefinitionSet := program.evaluateOptimizeAndReturnUsedCSS()

	executionElapsed := time.Since(executionStart)

	// Output
	var generateTimeElapsed time.Duration
	{
		generateStartTime := time.Now()
		// Output CSS definitions
		{
			var cssOutput bytes.Buffer
			for _, cssDefinition := range outputCSSDefinitionSet {
				name := cssDefinition.Name
				if len(name) == 0 {
					name = "<anonymous>"
				}
				cssOutput.WriteString(fmt.Sprintf("/* Name: %s */\n", name))
				cssOutput.WriteString(generate.PrettyCSS(cssDefinition))
			}
			outputFilepath := filepath.Clean(fmt.Sprintf("%s/%s.css", cssOutputDirectory, "main"))
			err := ioutil.WriteFile(
				outputFilepath,
				cssOutput.Bytes(),
				0644,
			)
			if err != nil {
				return err
			}
		}

		// Write to file
		for _, outputTemplateFile := range outputTemplateFileSet {
			err := ioutil.WriteFile(
				outputTemplateFile.Filepath,
				[]byte(outputTemplateFile.Content),
				0644,
			)
			if err != nil {
				panic(err)
			}
		}
		generateTimeElapsed = time.Since(generateStartTime)
	}

	//fmt.Printf("templateOutputDirectory: %s\n", templateOutputDirectory)
	fmt.Printf("File read time: %s\n", readFileTime)
	fmt.Printf("Parsing time: %s\n", parsingElapsed)
	fmt.Printf("Execution time: %s\n", executionElapsed)
	fmt.Printf("Generate/File write time: %s\n", generateTimeElapsed)
	totalTimeElapsed := time.Since(totalTimeStart)
	fmt.Printf("Total time: %s\n", totalTimeElapsed)
	return nil
}
