package refactor

import (
	"fmt"
	"go/parser"
	"go/ast"
)

type Refactor struct {
	//SymbolTable
	stmts []ast.Decl
	scope *ast.Scope
}

func RefactorSource(src string) *Refactor{
	refactor := new(Refactor)
	refactor.scope = ast.NewScope(nil)
	stmts, err := parser.ParseDeclList("", src, refactor.scope)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	refactor.stmts = stmts
	/*
	visitor := newRefactorVisitor()
	for _, stmt := range src.stmts {
		ast.Walk(visitor, stmt)
	}
	*/
	return refactor
}

func (src *Refactor) GetVariableNameAt(position int) (symbolName string) {
	for _, v := range src.scope.Objects {
		if v.Pos.Column <= position && v.Pos.Column + len(v.Name) > position {
			return v.Name
		}
	}
	return 
}
