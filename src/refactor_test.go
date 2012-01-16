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
	"reflect"
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
	idents []*ast.Ident
	decls []string
	parent *RenameVisitor
	node ast.Node
	doRename bool
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
	if node != nil {
		child :=  NewParentedRenameVisitor(v, node)
		child.StartVisit()
		return child
	}
	v.FinishVisit()
	return nil
}

func (v *RenameVisitor) AddIdentToScope(i *ast.Ident) {
	if v.parent == nil || contains(v.decls, i.Name) {
		v.idents = append(v.idents, i)
	} else {
		v.parent.AddIdent(i)
	}
}

func (v *RenameVisitor) AddIdent(i *ast.Ident) {
	switch n := v.node.(type) {
	case *ast.FuncDecl:
		v.AddIdentToScope(i)
	case *ast.BlockStmt:
		v.AddIdentToScope(i)
	default:
		if v.parent != nil {
			v.parent.AddIdent(i)
		}
	}
}

func contains(coll []string, target string) bool {
	for _, i := range coll {
		if i == target {
			return true
		}
	}
	return false
}

func (v *RenameVisitor) AddDecl(name string) {
	switch n := v.node.(type) {
	case *ast.FuncDecl:
		v.decls = append(v.decls, name)
	case *ast.BlockStmt:
		v.decls = append(v.decls, name)
	default:
		if v.parent != nil {
			v.parent.AddDecl(name)
		}
	}
}

func positionMatches(row, col int, pos token.Position) bool {
	return pos.Line == row && pos.Column == col
}

func (v *RenameVisitor) StartVisit() {
	fmt.Printf("%s<%v>\n", tabs(v.tabcount), reflect.TypeOf(v.node))
	switch n := v.node.(type) {
	case *ast.Field:
		for _, i := range n.Names {
			v.parent.AddDecl(i.Name)
			v.parent.AddIdent(i)
		}
	case *ast.ValueSpec:
		for _, i := range n.Names {
			v.parent.AddDecl(i.Name)
			v.parent.AddIdent(i)
		}
	case *ast.Ident:
		p := n.Pos()
		pos := v.fs.Position(p)
		if positionMatches(v.row, v.col, pos) && n.Name == v.origName {
			v.doRename = true
		}
		v.parent.AddIdent(n)
	}
}

func (v *RenameVisitor) PerformRename() {
	if !contains(v.decls, v.origName) && v.doRename {
		v.parent.doRename = v.doRename
	} else if v.parent == nil || v.doRename {
		for _, i := range v.idents {
			if i.Name == v.origName {
				i.Name = v.newName
			}
		}
	}
}

func (v *RenameVisitor) FinishVisit() {
	fmt.Printf("%s</%s>\n", tabs(v.tabcount), reflect.TypeOf(v.node))
	switch n := v.node.(type) {
	case *ast.FuncDecl:
		v.PerformRename()
	case *ast.BlockStmt:
		v.PerformRename()
	default:
		if v.doRename {
			v.parent.doRename = v.doRename
		}
	}
}

func NewRenameVisitor(fs *token.FileSet, origName, newName string, row, col int) *RenameVisitor {
	v := new(RenameVisitor)
	v.fs = fs
	v.origName = origName
	v.newName = newName
	v.row = row
	v.col = col
	return v
}

func NewParentedRenameVisitor(p *RenameVisitor, node ast.Node) *RenameVisitor {
	v := new(RenameVisitor)
	v.fs = p.fs
	v.origName = p.origName
	v.newName = p.newName
	v.row, v.col = p.row, p.col
	v.parent = p
	v.idents = make([]*ast.Ident, 0)
	v.decls = make([]string, 0)
	v.tabcount = p.tabcount + 1
	v.node = node
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

func testInputEqualsOutputWithNoTransform(t *testing.T) {
	expr := "j + 10"
	rw := ParseExprAndPrint(expr)
	checkBufEqString(t, rw, expr)
}

func testCanPrintMyOwnAst(t *testing.T) {
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
	input, oldName, newName, row, col := "{\nsomeName + 10\n}", "someName", "renamed", 2, 1
	expected := "{\nrenamed + 10}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameVariableInStmts(t *testing.T) {
	input, oldName, newName, row, col := "{\nsomeName := 1\nplusTwo := someName + 2\n}", "someName", "otherName", 2, 1
	expected := "{\notherName := 1\nplusTwo := otherName + 2\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameVariableOnType(t *testing.T) {
	input, oldName, newName, row, col := "type A struct {\nmyField int\n}", "myField", "newName", 2, 1
	expected := "type A struct {\nnewName int\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testOnlyRenamesVarInScope(t *testing.T) {
	input, oldName, newName, row, col := "{\n{\na := 0\na++\n}\n{\na := 2\na--\n}\n}", "a", "b", 3, 1
	expected := "{\n{\nb := 0\nb++\n}\n{\na := 2\na--\n}\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameFields(t *testing.T) {
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
func testCanLimitRenameToFieldInExpressions(t *testing.T) {
	input := "package main\ntype A struct {\n fld int\n}\nfunc (v A) Do(fld int) {\nv.fld += fld\n}"
	oldName, newName := "fld", "newFld"
	row, col := 6, 3
	expected := "package main\ntype A struct {\n newFld int\n}\nfunc (v A) Do(fld int) {\nv.newFld += fld\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}
