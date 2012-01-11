package refactor

import (
	"testing"
	"go/ast"
	"go/parser"
	"go/token"
	"fmt"
	"go/printer"
//	"bufio"
	"os"
	"bytes"
)

func Rename(expr, oldName, newName string) (ast.Node, *token.FileSet) {
	return RenameAt(expr, oldName, newName, -1, -1)
}
func RenameAt(expr, oldName, newName string, row, col int) (ast.Node, *token.FileSet) {
	contents, fs, _ := ParseExpr(expr)
	visitor := NewRenameVisitor(fs, oldName, newName, row, col)
	ast.Walk(visitor, contents)
	return contents, fs
}

type RenameVisitor struct {
	fs *token.FileSet
	origName, newName string
	row, col int
}

func (v *RenameVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.Ident:
		pos := v.fs.Position(node.Pos())
		fmt.Printf("--? %v,%v vs %v,%v && %s vs %v\n", v.row, v.col, pos.Line, pos.Column, n.Name, v.origName)
		if v.row == pos.Line && v.col == pos.Column && n.Name == v.origName {
			fmt.Printf("Renaming (complex)\n")
			n.Name = v.newName
			return NewRenameVisitor(v.fs, v.origName, v.newName, -1, -1)
		}
		if v.row == -1 && n.Name == v.origName {
			fmt.Printf("Renaming (simple)\n")
			n.Name = v.newName
		}
	}
	return v
}

func NewRenameVisitor(fs *token.FileSet, origName, newName string, row, col int) *RenameVisitor {
	v := new(RenameVisitor)
	v.fs = fs
	v.origName = origName
	v.newName = newName
	v.row, v.col = row, col
	return v
}

type ByteSlice []byte
func (bs *ByteSlice) Write(data []byte) (num int, err os.Error) {
	p := append(*bs, data...)
	*bs = p
	return len(data), nil
}

func (bs ByteSlice) Read(to []byte) (num int, err os.Error) {
	return copy(to, bs), nil
}

func Print(expr ast.Node, fs *token.FileSet) *bytes.Buffer {
	buf := make([]byte, 0)
	rw := bytes.NewBuffer(buf)
	printer.Fprint(rw, fs, expr)
	return rw
}

func ParseExpr(exprString string) (ast.Node, *token.FileSet, os.Error) {
	fs := token.NewFileSet()
	var expr ast.Node
	var err os.Error
	expr, err = parser.ParseExpr(fs, "hi.go", exprString)
	if err != nil {
		var stmts []ast.Stmt
		stmts, err = parser.ParseStmtList(fs, "hi.go", exprString)
		if err == nil {
			blockStmt := new(ast.BlockStmt)
			blockStmt.List = stmts
			blockStmt.Lbrace = 0
			blockStmt.Rbrace = stmts[len(stmts)-1].End()
			expr = blockStmt
		}  else {
			expr, err = parser.ParseFile(fs, "hi.go", exprString, 0)
		}
	}
	return expr, fs, err
}

func ParseExprAndPrint(exprString string) *bytes.Buffer {
	expr, fs, err := ParseExpr(exprString)
	if err != nil {
		fmt.Printf("Error parsing expression: %v\n", err)
		return nil
	}
	return Print(expr, fs)
}

func stringsAreEqual(a, b string) (bool, int) {
	if len(a) != len(b) {
		fmt.Printf("Lengths aren't equal: %v vs %v\n", len(a), len(b))
		return false, -1
	}
	for i := range a {
		if a[i] != b[i] {
			return false, i
		}
	}
	return true, -1
}

func hasExpectedOutput(buffer *bytes.Buffer, expected string) (bool, string) {
	if buffer.String() != expected {
		return false, fmt.Sprintf("Actual [%s]\nExpected [%s]\n", buffer.String(), expected)
	}
	return true, ""
}

func checkBufEqString(t *testing.T, buf *bytes.Buffer, expected string) {
	if equal, msg := hasExpectedOutput(buf, expected); !equal {
		t.Errorf(msg)
	}
}

func TestInputEqualsOutputWithNoTransform(t *testing.T) {
	expr := "j + 10"
	rw := ParseExprAndPrint(expr)
	checkBufEqString(t, rw, expr)
}

func TestCanPrintMyOwnAst(t *testing.T) {
	expr := new(ast.Ellipsis)
	fs := token.NewFileSet()
	expr.Ellipsis = 10
	expr.Elt = nil
	buf := Print(expr, fs)
	checkBufEqString(t, buf, "...")
}


func TestCanRenameVariableInExpr(t *testing.T) {
	contents, fs := Rename("someName + 10", "someName", "renamed")
	buf := Print(contents, fs)
	checkBufEqString(t, buf, "renamed + 10")
}

func TestCanRenameVariableInStmts(t *testing.T) {
	contents, fs := Rename("someName := 1\nplusTwo := someName + 2", "someName", "otherName")
	buf := Print(contents, fs)
	expected := ParseExprAndPrint("otherName := 1\nplusTwo := otherName + 2")
	checkBufEqString(t, buf, expected.String())
}

func TestCanRenameVariableOnType(t *testing.T) {
	contents, fs := Rename("type A struct {\nmyField int\n}", "myField", "newName")
	buf := Print(contents, fs)
	expected := ParseExprAndPrint("type A struct {\nnewName int\n}")
	checkBufEqString(t, buf, expected.String())
}

func TestOnlyRenamesVarInScope(t *testing.T) {
	contents, fs := RenameAt("{\na := 0\na++\n}\n{\na := 2\na--\n}", "a", "b", 2, 1)
	buf := Print(contents, fs)
	expected := ParseExprAndPrint("{\nb := 0\nb++\n}\n{\na := 2\na--\n}")
	checkBufEqString(t, buf, expected.String())
}
