package refactor

import (
	"fmt"
	"go/parser"
	"go/ast"
	"go/token"
)

type Refactor struct {
	scope *Scope;
	gimme chan chan token.Position
}

func RefactorFile(fileName string) *Refactor {
	file, err := parser.ParseFile(fileName, nil, 0)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	return RefactorDecls(file.Decls)
}

func RefactorSource(src interface{}) *Refactor {
	stmts, err := parser.ParseDeclList("", src)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	return RefactorDecls(stmts)
}

func RefactorDecls(stmts interface{}) *Refactor{
	ref := new(Refactor)
	ref.scope = NewScope()
	ref.gimme = make(chan chan token.Position, 100)
	visitor := newRefactorVisitor(token.Position{}, ref.scope, ref.gimme, 0, nil)
	switch t := stmts.(type) {
		case []interface{}:
			for _, stmt := range t {
				fmt.Printf("Visiting interface %v\n", stmt)
				ast.Walk(visitor, stmt)
			}
		case []ast.Stmt:
			for _, stmt := range t {
				fmt.Printf("Visiting statement %v\n", stmt)
				ast.Walk(visitor, stmt)
			}
		case []ast.Decl:
			for _, stmt := range t {
				fmt.Printf("Visiting declaration %v\n", stmt)
				ast.Walk(visitor, stmt)
			}
		default:
			fmt.Printf("Nope...%T\n", t)
	}
	return ref
}

func (src *Refactor) GetVariableNameAt(row, column int) string {
	return GetVariableNameForScopeAt(src.scope, row, column)
}

func GetVariableNameForScopeAt(scope *Scope, row, column int) string {
	for varName := range scope.positions {
		fmt.Printf("found position %v\n", varName)
		for _, pos := range scope.GetSites(varName) {
			fmt.Printf("  at %v\n", pos)
			if identContainsPosition(varName, pos, row, column) {
				return varName
			}
		}
	}
	if( scope.children != nil ) {
		for _, childScope := range scope.children {
			if childScope != nil {
				return GetVariableNameForScopeAt(childScope, row, column)
			}
		}
	}
	return ""
}

func (src *Refactor) PositionsForSymbolAt(row, column int) chan chan token.Position {
	defer close(src.gimme)
	return PositionsForSymbolAt(src.scope, row, column, src.gimme)
}

func PositionsForSymbolAt(scope *Scope, row, column int, gimme chan chan token.Position) chan chan  token.Position {
	for varName, _ := range scope.positions {
		name := varName
		for _, pos := range scope.GetSites(name) {
			out := make(chan token.Position)
			gimme <- out
			go func(pos token.Position) {
				defer close(out)
				if identContainsPosition(name, pos, row, column) {
					for _, p := range scope.GetSites(name) {
						out <- p
					}
				}
			}(pos)
		}
	}

	if( scope.children != nil ) {
		for _, childScope := range scope.children {
			if childScope != nil {
				PositionsForSymbolAt(childScope, row, column, gimme)
			}
		}
	}
	return gimme
}

func identContainsPosition(varName string, identPosition token.Position, row, column int) bool {
	return identPosition.Line == row && identPosition.Column <= column && identPosition.Column + len(varName) > column
}

func copyChannelToArray(in chan chan token.Position) []token.Position {
	out := make(chan token.Position, 10)
	posCount := 0
	for posChan := range in {
		for pos := range posChan {
			posCount++
			out <- pos
		}
	}
	array := make([]token.Position, posCount)
	for i, _ := range array {
		array[i] = <- out
	}

	return array
}
