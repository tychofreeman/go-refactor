package refactor

import (
	"fmt"
	"go/parser"
	"go/ast"
	"go/token"
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
			if identContainsPosition(varName, pos, row, column) {
				return varName
			}
		}
	}
	return ""
}

func (src *Refactor) PositionsForSymbolAt(row, column int) []token.Position {
	for varName, _ := range src.varSites.varSites {
		for _, pos := range src.varSites.GetSites(varName) {
			if identContainsPosition(varName, pos, row, column) {
				return src.varSites.GetSites(varName)
			}
		}
	}
	return nil
}

func identContainsPosition(varName string, identPosition token.Position, row, column int) bool {
	return identPosition.Line == row && identPosition.Column <= column && identPosition.Column + len(varName) > column
}
