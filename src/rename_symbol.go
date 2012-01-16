package refactor

import (
	"fmt"
	"go/parser"
	"go/ast"
	"go/token"
	"os"
)

type Refactor struct {
	scope *Scope;
	gimme chan chan token.Position
}

func RefactorFile(fileName string) *Refactor {
	fileSet := token.NewFileSet()
	fileInfo, _ := os.Lstat(fileName)
	fileSet.AddFile(fileName, fileSet.Base(), int(fileInfo.Size))
	file, err := parser.ParseFile(fileSet, fileName, nil, 0)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	return RefactorDecls(file.Decls)
}

func NodesLen(ns []ast.Decl) int {
	return int(ns[0].End())
}

func stmtsFileSet(contents []ast.Decl) *token.FileSet {
	return fileSet(NodesLen(contents))
}

func stringFileSet(contents string) *token.FileSet {
	return fileSet(len(contents))
}

func fileSet(contentLength int) *token.FileSet {
	fileSet := token.NewFileSet()
	fileSet.AddFile("source", fileSet.Base(), contentLength)
	return fileSet
}

func RefactorSource(src string) *Refactor {
	files := stringFileSet(src)
	stmts, err := parser.ParseDeclList(files, "", src)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	return RefactorDeclsInFileSet(stmts, files)
}

func RefactorDecls(stmts []ast.Decl) *Refactor{
	return RefactorDeclsInFileSet(stmts, stmtsFileSet(stmts))
}

func RefactorDeclsInFileSet(stmts []ast.Decl, files *token.FileSet) *Refactor {
	ref := new(Refactor)
	ref.scope = NewScope()
	ref.gimme = make(chan chan token.Position, 100)
	visitor := newRefactorVisitor(token.Position{}, ref.scope, ref.gimme, 0, nil, files)
	for _, stmt := range stmts {
		ast.Walk(visitor, stmt)
	}
	return ref
}

func (src *Refactor) GetVariableNameAt(row, column int) (string, *Scope) {
	return GetVariableNameForScopeAt(src.scope, row, column)
}

func GetVariableNameForScopeAt(scope *Scope, row, column int) (string, *Scope) {
	for varName := range scope.positions {
		for _, pos := range scope.GetSites(varName) {
			if identContainsPosition(varName, pos, row, column) {
				return varName, scope
			}
		}
	}
	if( scope.children != nil ) {
		for _, childScope := range scope.children {
			if( childScope != nil ) {
				rtnName, rtnScope := GetVariableNameForScopeAt(childScope, row, column)
				if( rtnName != "" ) {
					return rtnName, rtnScope
				}
			}
		}
	}
	return "", &Scope{}
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
