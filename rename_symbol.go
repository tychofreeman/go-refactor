package  refactor

import (
	"fmt"
	"go/parser"
	"go/ast"
	"reflect"
)

type Refactor struct {
	stmts []ast.Decl
	scope *ast.Scope
}

func RefactorSource(src string) *Refactor{
	refactor := new(Refactor)
	scope := ast.NewScope(nil)
	stmts, err := parser.ParseDeclList("", src, scope)
	if err != nil {
		panic(fmt.Sprintf("Could not parse input. %v", err))
	}
	refactor.stmts = stmts
	refactor.scope = scope
	return refactor
}

type PrintVisitor struct {
}

func (pw PrintVisitor) Visit(node interface{}) (ast.Visitor) {
	fmt.Printf("Visiting %v\n", node, reflect.Typeof(node).Name())
	return pw
}

func (src *Refactor) GetVariableNameAt(position int) (symbolName string) {
	for _, stmt := range src.stmts {
		ast.Walk(new(PrintVisitor), stmt)
		switch st := stmt.(type) {
			case ast.Stmt:
				fmt.Printf("Found statement of %v at position %v\n", st.Pos(), st.Pos())
			case ast.Decl:
				fmt.Printf("Found decl of %v at position %v\n", st.Pos(), st.Pos())
		}
		for v, k := range src.scope.Objects {
			fmt.Printf("Found object %v at %v\n", k, v)
		}
		fmt.Printf("Found stmt %v\n", stmt)
	}
	return ""
}
