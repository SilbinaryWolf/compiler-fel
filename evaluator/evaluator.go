package evaluator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/parser"
)

type Program struct {
	globalScope *Scope
}

func New() *Program {
	p := new(Program)
	p.globalScope = NewScope(nil)
	return p
}

func (program *Program) RunProject(projectDirpath string) error {
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
		if errors := p.GetErrors(); len(errors) > 0 {
			errorOrErrors := "errors"
			if len(errors) == 1 {
				errorOrErrors = "error"
			}
			fmt.Printf("Found %d %s in \"%s\"\n", len(errors), errorOrErrors, filepath)
			for _, err := range errors {
				fmt.Printf("- %v \n\n", err)
			}
			return fmt.Errorf("Error parsing file: %v", filepath)
		}
		configAstFile = astFile
	}

	if configAstFile == nil {
		return fmt.Errorf("Cannot find config.fel in root of project directory: %v", projectDirpath)
	}

	// Evaluate config file
	program.evaluateBlock(configAstFile, program.globalScope)
	value, ok := program.globalScope.GetCurrentScope("templateOutputDirectory")
	if !ok {
		return fmt.Errorf("%s is undefined in config.fel. This definition is required.", "templateOutputDirectory")
	}
	if value.Kind() != KindString {
		return fmt.Errorf("%s is expected to be a string.", "templateOutputDirectory")
	}

	templateOutputDirectory := fmt.Sprintf("%s/%s", projectDirpath, value.String())
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

	// Parse files
	astFiles := make([]*ast.File, 0, 50)
	p := parser.New()
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
		if errors := p.GetErrors(); len(errors) > 0 {
			errorOrErrors := "errors"
			if len(errors) == 1 {
				errorOrErrors = "error"
			}
			fmt.Printf("Found %d %s in \"%s\"\n", len(errors), errorOrErrors, filepath)
			for _, err := range errors {
				fmt.Printf("- %v \n\n", err)
			}
			return fmt.Errorf("Error parsing file: %v", filepath)
		}
		astFiles = append(astFiles, astFile)
	}

	// DEBUG
	json, _ := json.MarshalIndent(astFiles, "", "   ")
	fmt.Printf("%s", string(json))

	// Execute templates
	// todo(Jake): Refactor above filewalk logic to find "config.fel" directly first, then walk "components"
	//			   then walk "templates"
	fmt.Printf("templateOutputDirectory: %s\n", templateOutputDirectory)
	panic("Evaluator::RunProject(): todo(Jake): The rest of the function")
	return nil
}
