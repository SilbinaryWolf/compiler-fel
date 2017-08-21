package evaluator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/data"
	"github.com/silbinarywolf/compiler-fel/generate"
	"github.com/silbinarywolf/compiler-fel/parser"
)

type TemplateFile struct {
	Filepath string
	Content  string
}

type Program struct {
	globalScope *Scope
	debugLevel  int
}

func New() *Program {
	p := new(Program)
	p.globalScope = NewScope(nil)
	return p
}

func (program *Program) GetConfigString(configName string) (string, error) {
	value, ok := program.globalScope.Get(configName)
	if !ok {
		return "", fmt.Errorf("%s is undefined in config.fel. This definition is required.", configName)
	}
	if value.Kind() != data.KindString {
		return "", fmt.Errorf("%s is expected to be a string.", configName)
	}
	return value.String(), nil
}

func (program *Program) RunProject(projectDirpath string) error {
	totalTimeStart := time.Now()

	configFilepath := projectDirpath + "/config.fel"
	if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
		return fmt.Errorf("Cannot find config.fel in root of project directory: %v", configFilepath)
	}

	templateInputDirectory := projectDirpath + "/templates"
	// Check if input templates directory exists
	{
		_, err := os.Stat(templateInputDirectory)
		if err != nil {
			return fmt.Errorf("Error with directory \"templates\" directory in project directory: %v", err)
		}
		if os.IsNotExist(err) {
			return fmt.Errorf("Expected to find \"templates\" directory in: %s", projectDirpath)
		}
	}

	// Find and parse config.fel
	var configAstFile *ast.File
	{
		p := parser.New()
		filepath := configFilepath
		filecontentsAsBytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("An error occurred reading file: %v, Error message: %v", filepath, err)
		}
		astFile := p.Parse(filecontentsAsBytes, filepath)
		if astFile == nil {
			panic("Unexpected parse error (Parse() returned a nil ast.File node)")
		}
		configAstFile = astFile
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
	templateOutputDirectory, err := program.GetConfigString("templateOutputDirectory")
	if err != nil {
		return err
	}
	templateOutputDirectory = fmt.Sprintf("%s/%s", projectDirpath, templateOutputDirectory)
	cssOutputDirectory, err := program.GetConfigString("cssOutputDirectory")
	if err != nil {
		return err
	}
	cssOutputDirectory = fmt.Sprintf("%s/%s", projectDirpath, cssOutputDirectory)

	// Check if output templates directory exists
	{
		_, err := os.Stat(templateOutputDirectory)
		if err != nil {
			return fmt.Errorf("Error with directory: %v", err)
		}
		if os.IsNotExist(err) {
			return fmt.Errorf("templateOutputDirectory does not exist: %s", templateOutputDirectory)
		}
	}

	// Get all files in folder recursively with *.fel
	filepathSet := make([]string, 0, 50)
	{
		err := filepath.Walk(templateInputDirectory, func(path string, f os.FileInfo, _ error) error {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".fel" {
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
	}

	// Parse files
	astFiles := make([]*ast.File, 0, 50)
	p := parser.New()
	parsingStart := time.Now()
	for _, filepath := range filepathSet {
		filecontentsAsBytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("An error occurred reading file: %v, Error message: %v", filepath, err)
			//continue
		}
		astFile := p.Parse(filecontentsAsBytes, filepath)
		if astFile == nil {
			panic("Unexpected parse error (Parse() returned a nil ast.File node)")
		}
		astFiles = append(astFiles, astFile)
	}
	parsingElapsed := time.Since(parsingStart)

	if p.HasErrors() {
		p.PrintErrors()
		return fmt.Errorf("Stopping due to parsing errors.")
	}

	/*{
		json, _ := json.MarshalIndent(astFiles, "", "   ")
		fmt.Printf("%s", string(json))
	}*/

	outputTemplateFileSet := make([]TemplateFile, 0, len(astFiles))
	outputCSSDefinitionSet := make([]*data.CSSDefinition, 0, 3)

	// Execute template
	executionStart := time.Now()
	for _, astFile := range astFiles {
		program.globalScope = NewScope(nil)

		scope := program.globalScope
		htmlNode := program.evaluateTemplate(astFile, scope)

		if len(htmlNode.ChildNodes) == 0 {
			return fmt.Errorf("No top level HTMLNode or HTMLText found in %s.", astFile.Filepath)
		}
		if htmlNode == nil {
			panic(fmt.Sprintf("No html node found in %s.", astFile.Filepath))
		}

		// Print CSS definitions
		cssDefinitionList := scope.cssDefinitions
		if len(cssDefinitionList) > 0 {
			for _, cssDefinition := range cssDefinitionList {
				outputCSSDefinitionSet = append(outputCSSDefinitionSet, cssDefinition)
			}
		}

		baseFilename := astFile.Filepath[len(templateInputDirectory) : len(astFile.Filepath)-4]
		outputFilepath := filepath.Clean(fmt.Sprintf("%s%s.html", templateOutputDirectory, baseFilename))
		result := TemplateFile{
			Filepath: outputFilepath,
			Content:  generate.PrettyHTML(htmlNode),
		}
		outputTemplateFileSet = append(outputTemplateFileSet, result)
	}
	executionElapsed := time.Since(executionStart)

	// Output CSS definitions
	{
		var cssOutput bytes.Buffer
		for _, cssDefinition := range outputCSSDefinitionSet {
			cssOutput.WriteString(fmt.Sprintf("/* Name: %s */\n", cssDefinition.Name))
			cssOutput.WriteString(generate.PrettyCSS(cssDefinition))
		}
		outputFilepath := filepath.Clean(fmt.Sprintf("%s/%s.css", cssOutputDirectory, "main"))
		fmt.Printf("%s\n", outputFilepath)
		err := ioutil.WriteFile(
			outputFilepath,
			cssOutput.Bytes(),
			0644,
		)
		if err != nil {
			panic(err)
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

	//fmt.Printf("templateOutputDirectory: %s\n", templateOutputDirectory)

	fmt.Printf("Parsing time: %s\n", parsingElapsed)
	fmt.Printf("Execution time: %s\n", executionElapsed)
	totalTimeElapsed := time.Since(totalTimeStart)
	fmt.Printf("Total time: %s\n", totalTimeElapsed)
	panic("Evaluator::RunProject(): todo(Jake): The rest of the function")
	return nil
}
