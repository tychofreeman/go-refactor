package refactor

import (
	"fmt"
	"go/parser"
	"go/ast"
)

type Refactor struct {
	varSites *VarSites;
}

func RefactorSource(src string) *Refactor {
	refactor := new(Refactor)
	refactor.varSites = NewVarSites()
	scope := ast.NewScope(nil)
	stmts, err := parser.ParseDeclList("", src, scope)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	visitor := newRefactorVisitor(refactor.varSites)
	for _, stmt := range stmts {
		ast.Walk(visitor, stmt)
	}
	return refactor
}

func (src *Refactor) GetVariableNameAt(row, column int) string {
	for varName, _ := range src.varSites.varSites {
		for _, pos := range src.varSites.GetSites(varName) {
			if pos.Line == row && pos.Column <= column && pos.Column + len(varName) > column {
				return varName
			}
		}
	}
	return ""
}
