package refactor

import (
	"testing"
	"go/ast"
	"go/parser"
	"go/token"
	"fmt"
	"go/printer"
	"os"
	"bytes"
)

type ByteSlice []byte
func (bs *ByteSlice) Write(data []byte) (num int, err os.Error) {
	p := append(*bs, data...)
	*bs = p
	return len(data), nil
}

func (bs ByteSlice) Read(to []byte) (num int, err os.Error) {
	return copy(to, bs), nil
}

func RefactorPrint(expr ast.Node, fs *token.FileSet) *bytes.Buffer {
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
	return RefactorPrint(expr, fs)
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
	buf := RefactorPrint(expr, fs)
	checkBufEqString(t, buf, "...")
}

func verifyRename(t *testing.T, input, expected, oldName, newName string, row, col int) {
	contents, fs := RenameAt(input, oldName, newName, row, col)
	buf := RefactorPrint(contents, fs)
	expectedBuf := ParseExprAndPrint(expected)
	//ast.Print(fs, contents)
	checkBufEqString(t, buf, expectedBuf.String())
}

func TestCanRenameVariableInExpr(t *testing.T) {
	input, oldName, newName, row, col := "{\nsomeName + 10\n}", "someName", "renamed", 2, 1
	expected := "{\nrenamed + 10}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenameVariableInStmts(t *testing.T) {
	input, oldName, newName, row, col := "{\nsomeName := 1\nplusTwo := someName + 2\n}", "someName", "otherName", 2, 1
	expected := "{\notherName := 1\nplusTwo := otherName + 2\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenameVariableOnType(t *testing.T) {
	input, oldName, newName, row, col := "type A struct {\nmyField int\n}", "myField", "newName", 2, 1
	expected := "type A struct {\nnewName int\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestOnlyRenamesVarInScope(t *testing.T) {
	input, oldName, newName, row, col := "{\n{\nvar a int = 0\na++\n}\n{\nvar a int = 2\na--\n}\n}", "a", "b", 3, 1
	expected := "{\n{\nvar b int = 0\nb++\n}\n{\nvar a int = 2\na--\n}\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenameFields(t *testing.T) {
	input, oldName, newName, row, col := "type A struct {\n fld int\n}", "fld", "newFld", 2, 2
	expected := "type A struct {\n newFld int\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenameFieldsInExpressions(t *testing.T) {
	input := "{\nv.fld += 10\n}"
	oldName, newName := "fld", "newFld"
	row, col := 2, 3
	expected := "{\nv.newFld += 10\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenamePreviousVars(t *testing.T) {
	input := "{\nfld += 5\nfld += 10\n}"
	oldName, newName := "fld", "newFld"
	row, col := 3, 1
	expected := "{\nnewFld += 5\nnewFld += 10\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestRenamesAllInstancesInsideDefiningScope(t *testing.T) {
	input := "{\nvar fld int = 0\n{\nfld += 5\n}\n}"
	oldName, newName := "fld", "newFld"
	row, col := 4, 1
	expected := "{\nvar newFld int = 0\n{\nnewFld += 5\n}\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenameParameter(t *testing.T) {
	input := "package main\nfunc Do(fld int) {\nfld += 5\n}"
	oldName, newName := "fld", "newFld"
	row, col := 3, 1
	expected := "package main\nfunc Do(newFld int) {\nnewFld += 5\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

// Ignore this one until the renameParam is working
func TestCanLimitRenameToFieldInExpressions(t *testing.T) {
	input := "package main\ntype A struct {\n fld int\n}\nfunc (v A) Do(fld int) {\nv.fld += fld\n}"
	oldName, newName := "fld", "newFld"
	row, col := 6, 3
	expected := "package main\ntype A struct {\n newFld int\n}\nfunc (v A) Do(fld int) {\nv.newFld += fld\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}
