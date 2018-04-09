package parser

import (
	"github.com/silbinarywolf/compiler-fel/ast"
	"github.com/silbinarywolf/compiler-fel/errors"
)

type Typer struct {
	errors.ErrorHandler
	typeinfo                      TypeInfoManager
	typecheckHtmlNodeDependencies map[string]*ast.Call
}

func NewTyper() *Typer {
	p := new(Typer)

	p.ErrorHandler.Init()
	p.ErrorHandler.SetDeveloperMode(true)

	p.typeinfo.Init()

	// NOTE(Jake): 2018-04-09
	//
	// Keeping this value unset is intentional.
	// ... I think.
	//
	//p.typecheckHtmlNodeDependencies = nil

	return p
}
