package evaluator

import (
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
	// Get all files in folder recursively with *.fel
	filepathSet := make([]string, 0, 50)
	err := filepath.Walk(projectDirpath, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".fel" {
			filepathSet = append(filepathSet, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("An error occurred reading: %v, Error Message: %v", projectDirpath, err)
	}

	if len(filepathSet) == 0 {
		return fmt.Errorf("No *.fel files found in your project directory: %v", projectDirpath)
	}

	// Find and parse config.fel
	var configAstFile *ast.File
	{
		p := parser.New()
		for _, filepath := range filepathSet {
			baseFilename := filepath[len(projectDirpath):len(filepath)]
			// If filename isn't config.fel in root dir, skip it.
			if baseFilename != "\\config.fel" {
				continue
			}
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
			break
		}
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

	// Execute templates
	// todo(Jake): Refactor above filewalk logic to find "config.fel" directly first, then walk "components"
	//			   then walk "templates"
	fmt.Printf("templateOutputDirectory: %s\n", templateOutputDirectory)
	panic("Evaluator::RunProject(): todo(Jake): The rest of the function")
	return nil
}
