package refactor

import (
	"fmt"
	"go/parser"
	"go/ast"
	"go/token"
)

type Refactor struct {
	scope *Scope;
	gimme chan token.Position
}

func RefactorSource(file string, src string) *Refactor {
	ref := new(Refactor)
	ref.scope = NewScope()
	stmts, err := parser.ParseDeclList(file, src)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	ref.gimme = make(chan token.Position)
	visitor := newRefactorVisitor(token.Position{}, ref.scope, ref.gimme)
	for _, stmt := range stmts {
		ast.Walk(visitor, stmt)
	}
	close(ref.gimme)
	return ref
}

func (src *Refactor) GetVariableNameAt(row, column int) string {
	for varName := range src.scope.positions {
		for _, pos := range src.scope.GetSites(varName) {
			if identContainsPosition(varName, pos, row, column) {
				return varName
			}
		}
	}
	return ""
}

func (src *Refactor) PositionsForSymbolAt(row, column int) chan token.Position {
	for varName, _ := range src.scope.positions {
		for _, pos := range src.scope.GetSites(varName) {
			go func(pos token.Position) {
				if identContainsPosition(varName, pos, row, column) {
					for _, p := range src.scope.GetSites(varName) {
						src.gimme <- p
					}
				}
				close(src.gimme)
			}(pos)
		}
	}
	return src.gimme
}

func identContainsPosition(varName string, identPosition token.Position, row, column int) bool {
	return identPosition.Line == row && identPosition.Column <= column && identPosition.Column + len(varName) > column
}
