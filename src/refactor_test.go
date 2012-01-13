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
	decls []*ast.Decl
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

func (v *RenameVisitor) Visit(node ast.Node) ast.Visitor {
	fmt.Printf("...visit\n")
	if node != nil {
		return v.StartVisit(node)
	}
	fmt.Printf("%v</Visit rename=%v type=%v position=%v>\n", tabs(v.tabcount), v.doRename, reflect.TypeOf(v.node), v.Positions())
	return v.FinishVisit()
}

func (v *RenameVisitor) StartVisit(node ast.Node) ast.Visitor {
	v.doRename = false
	v.node = node
	fmt.Printf("%v<Visit(%v %v(%v))>\n", tabs(v.tabcount), reflect.TypeOf(node), node.Pos(), v.Positions())
	switch n := node.(type) {
	default:
		return NewParentedRenameVisitor(v.fs, v.origName, v.newName, v.row, v.col, v)
	case *ast.Ident:
		v.parent.AddIdent(n)
		pos := v.fs.Position(node.Pos())
		// If we find a matching identifier, every node below the current block is a valid rename target
		if v.row == pos.Line && v.col == pos.Column && n.Name == v.origName {
			fmt.Printf("%v<ShouldDoRename from=%v to=%v/>\n",tabs(v.tabcount), n.Name, v.newName) 
			v.parent.ShouldDoRename()
		}
	}
	return v
}

func (v *RenameVisitor) ShouldDoRename() {
	switch v.node.(type) {
	case *ast.BlockStmt:
		v.doRename = true
		fmt.Printf("%v<Caught ShouldDoRename s=%v from=%v to=%v>\n", tabs(v.tabcount), v.doRename, v.origName, v.newName)
		break
	default:
		v.parent.ShouldDoRename()
	}
}

func (v *RenameVisitor) AddIdent(ident *ast.Ident) {
	switch v.node.(type) {
	case *ast.BlockStmt:
		v.idents = append(v.idents, ident)
		break
	default:
		if v.parent != nil {
			v.parent.AddIdent(ident)
		}
	}
	v.idents = append(v.idents, ident)
}

func (v *RenameVisitor) Positions() string {
	if v.node == nil {
		return ""
	}
	pp := v.fs.Position(v.node.Pos())
	pe := v.fs.Position(v.node.End())
	return fmt.Sprintf("%v (%v,%v - %v,%v)", pp.Filename, pp.Line, pp.Column, pe.Line, pe.Column)
}

func (v *RenameVisitor) FinishVisit() ast.Visitor {
	switch _ := v.node.(type) {
	case *ast.BlockStmt:
		fmt.Printf("%v<Idents>\n", tabs(v.tabcount))
		for _, i := range v.idents {
			fmt.Printf("\t%v<ReferringTo pos=%v/>\n", tabs(v.tabcount), v.positions(i))
		}
		fmt.Printf("%v</Idents>\n", tabs(v.tabcount))
	}
	if v.doRename {
		fmt.Printf("%v<Renaming identCount=%v/>\n", tabs(v.tabcount), len(v.idents))
		for _, i := range v.idents {
			fmt.Printf("\t%v<Rename from=%v to=%v/>\n", tabs(v.tabcount), i.Name, v.newName)
			if i.Name == v.origName {
				i.Name = v.newName
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
	v.idents = make([]*ast.Ident, 0)
	if parent != nil {
		v.tabcount = v.parent.tabcount + 1
	}
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
	ast.Print(fs, contents)
	checkBufEqString(t, buf, expectedBuf.String())
}

func testCanRenameVariableInExpr(t *testing.T) {
	input, oldName, newName, row, col := "someName + 10", "someName", "renamed", -1, -1
	expected := "renamed + 10"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameVariableInStmts(t *testing.T) {
	input, oldName, newName, row, col := "someName := 1\nplusTwo := someName + 2", "someName", "otherName", -1, -1
	expected := "otherName := 1\nplusTwo := otherName + 2"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameVariableOnType(t *testing.T) {
	input, oldName, newName, row, col := "type A struct {\nmyField int\n}", "myField", "newName", -1, -1
	expected := "type A struct {\nnewName int\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func TestOnlyRenamesVarInScope(t *testing.T) {
	input, oldName, newName, row, col := "{\n{\na := 0\na++\n}\n{\na := 0\na++\n}\n{\na := 2\na--\n}}", "a", "b", 2, 1
	expected := "{{\nb := 0\nb++\n}\n{\na := 2\na--\n}}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameFields(t *testing.T) {
	input, oldName, newName, row, col := "type A struct {\n fld int\n}", "fld", "newFld", 2, 2
	expected := "type A struct {\n newFld int\n}"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameFieldsInExpressions(t *testing.T) {
	input := "v.fld += 10"
	oldName, newName := "fld", "newFld"
	row, col := 1, 3
	expected := "v.newFld += 10"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenamePreviousVars(t *testing.T) {
	input := "fld += 5\nfld += 10"
	oldName, newName := "fld", "newFld"
	row, col := 2, 1
	expected := "newFld += 5\nnewFld += 10"
	verifyRename(t, input, expected, oldName, newName, row, col)
}

func testCanRenameParameter(t *testing.T) {
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
