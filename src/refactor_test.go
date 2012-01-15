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
	idents map[ast.Node][]*ast.Ident
	decls map[ast.Node][]*ast.Decl
	parent *RenameVisitor
	node ast.Node
	doRename map[ast.Node]bool
	tabcount int
}

func tabs(count int) string {
	if count == 0 {
		return ""
	}
	return "\t" + tabs(count - 1)
}

func (v *RenameVisitor) positions(node ast.Node) string {
	pp := v.fs.Position(node.Pos())
	pe:= v.fs.Position(node.End())
	return fmt.Sprintf("(%v %v,%v-%v,%v)", pp.Filename, pp.Line, pp.Column, pe.Line, pe.Column)
}

func (v *RenameVisitor) Positions() string {
	if v.node == nil {
		return ""
	}
	return v.positions(v.node)
}

func (v *RenameVisitor) Visit(node ast.Node) ast.Visitor {
	if v.node != nil {
		fmt.Printf("Finishing visit on node %v\n", node)
		v.FinishVisit()
	}
	v.node = node
	if node != nil {
		return v.StartVisit(node)
	}
	return v
}

func (v *RenameVisitor) StartVisit(node ast.Node) ast.Visitor {
	v.node = node
	switch n := node.(type) {
	default:
		return NewParentedRenameVisitor(v.fs, v.origName, v.newName, v.row, v.col, v)
	case *ast.Ident:
		v.parent.AddIdent(n)
		pos := v.fs.Position(node.Pos())
		if v.row == pos.Line && v.col == pos.Column && n.Name == v.origName {
			v.parent.ShouldDoRename()
		}
	}
	return v
}

func (v *RenameVisitor) ShouldDoRename() {
	switch v.node.(type) {
	case *ast.BlockStmt:
		v.doRename[v.node] = true
		fmt.Printf("Should Do Rename for %s on node at %v! %v\n", v.origName, v.positions(v.node), v.doRename[v.node])
		break
	default:
		v.parent.ShouldDoRename()
	}
}

func (v *RenameVisitor) AddIdent(ident *ast.Ident) {
	switch v.node.(type) {
	case *ast.BlockStmt:
		idents, found := v.idents[v.node]
		if !found {
			idents = make([]*ast.Ident, 0)
			v.idents[v.node] = idents
		}
		v.idents[v.node] = append(idents, ident)
		break
	default:
		if v.parent != nil {
			v.parent.AddIdent(ident)
		}
	}
}

func (v *RenameVisitor) AddDecl(decl *ast.Decl) {
	switch v.node.(type) {
	case *ast.BlockStmt:
		if v.decls[v.node] == nil
		v.decls = append(v.decls, decl)
	defaults:
		if v.parent!= nil {
			v.parent.AddDecl(decl)
		}
	}
}

func (v *RenameVisitor) FinishVisit() ast.Visitor {
	fmt.Printf("Do Rename? for %s on node at %v? %v\n", v.origName, v.positions(v.node), v.doRename[v.node])
	if v.doRename[v.node] {
		for _, node := range v.idents[v.node] {
			if node.Name == v.origName {
				node.Name = v.newName
			}
		}
	}
	return nil
}

func NewRenameVisitor(fs *token.FileSet, origName, newName string, row, col int) *RenameVisitor {
	return NewParentedRenameVisitor(fs, origName, newName, row, col, nil)
}

func NewParentedRenameVisitor(fs *token.FileSet, origName, newName string, row, col int, parent *RenameVisitor) *RenameVisitor {
	v := new(RenameVisitor)
	v.fs = fs
	v.origName = origName
	v.newName = newName
	v.row, v.col = row, col
	v.parent = parent
	v.idents = make(map[ast.Node][]*ast.Ident, 0)
	if parent != nil {
		v.tabcount = v.parent.tabcount + 1
	}
	v.doRename = make(map[ast.Node]bool)
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

func testCanRenameVariableInExpr(t *testing.T) {
	input, oldName, newName, row, col := "{\nsomeName + 10\n}", "someName", "renamed", -1, -1
	expected := "{\nrenamed + 10}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenameVariableInStmts(t *testing.T) {
	input, oldName, newName, row, col := "{\nsomeName := 1\nplusTwo := someName + 2\n}", "someName", "otherName", -1, -1
	expected := "{\notherName := 1\nplusTwo := otherName + 2\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenameVariableOnType(t *testing.T) {
	input, oldName, newName, row, col := "type A struct {\nmyField int\n}", "myField", "newName", -1, -1
	expected := "type A struct {\nnewName int\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestOnlyRenamesVarInScope(t *testing.T) {
	input, oldName, newName, row, col := "{\n{\na := 0\na++\n}\n{\na := 2\na--\n}}", "a", "b", 3, 1
	expected := "{{\nb := 0\nb++\n}\n{\na := 2\na--\n}}"
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
	row, col := 1, 3
	expected := "{\nv.newFld += 10\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestCanRenamePreviousVars(t *testing.T) {
	input := "{\nfld += 5\nfld += 10\n}"
	oldName, newName := "fld", "newFld"
	row, col := 2, 1
	expected := "{\nnewFld += 5\nnewFld += 10\n}"
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
func testCanLimitRenameToFieldInExpressions(t *testing.T) {
	input := "package main\ntype A struct {\n fld int\n}\nfunc (v A) Do(fld int) {\nv.fld += fld\n}"
	oldName, newName := "fld", "newFld"
	row, col := 6, 3
	expected := "package main\ntype A struct {\n newFld int\n}\nfunc (v A) Do(fld int) {\nv.newFld += fld\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}
