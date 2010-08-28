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
	depth int
}

func (pw PrintVisitor) Visit(node interface{}) (ast.Visitor) {
	if node != nil {
		for i := 0; i < pw.depth; i++ {
			fmt.Printf("  ")
		}
		fmt.Printf("Visiting %v\n", nodeToString(node))
	}
	childrenWalker := new(PrintVisitor)
	childrenWalker.depth = pw.depth + 1
	return childrenWalker
}

func nodeToString(node interface{}) string {
	switch n := node.(type) {
		case *ast.GenDecl:
			return "[General Declaration]"
		case *ast.ValueSpec:
			return "[ValueSpec]"
		case []ast.Spec:
			return arrayToString(n)
		case []ast.Ident:
			return arrayToString(n)
		case []*ast.Ident:
			return arrayToString(n)
		case []ast.Expr:
			return arrayToString(n)
		case *reflect.SliceValue:
			return arrayToString(n)
		case *ast.Ident:
			return fmt.Sprintf("[Ident: %v at '%v' offset:%v line:%v col:%v]", n.String(), n.Filename, n.Offset, n.Line, n.Column)
		case *ast.BasicLit:
			return fmt.Sprintf("[BasicLiteral: %v %v]", n.Kind, n.Value)
		case *ast.Object:
			return fmt.Sprintf("[Object: %v %v]", n.Kind, n.Name)
		case *reflect.IntValue:
			return fmt.Sprintf("%v ", n.Get())
		case *reflect.StringValue:
			return fmt.Sprintf("\"%v\"", n.Get())
		default:
			return fmt.Sprintf("[Type %v]", reflect.Typeof(node))
	}
	return ""
}

func structToString(s *reflect.StructValue) string {
	t := s.Type()
	desc := " {"
	switch t2 := t.(type) {
		case *reflect.StructType:
			for i := 0; i < s.NumField(); i++ {
				desc = fmt.Sprintf("%v %v:%v", desc, t2.Field(i).Name ,nodeToString(s.Field(i)))
			}
	}
	return fmt.Sprintf("%v }", desc)
}

func arrayToString(origNodes interface{}) string {
	rtn := "["
	nodes := reflect.NewValue(origNodes)
	switch n := nodes.(type) {
		case *reflect.ArrayValue:
			for i :=  0; i < n.Len(); i++ {
				rtn = fmt.Sprintf("%v %v", rtn, nodeToString(n.Elem(i)))
			}
		case *reflect.SliceValue:
			for i :=  0; i < n.Len(); i++ {
				rtn = fmt.Sprintf("%v %v", rtn, nodeToString(n.Elem(i)))
			}
	}
	return fmt.Sprintf("%v]", rtn)
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
