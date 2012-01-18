package refactor

import (
	"go/ast"
	"go/token"
	"fmt"
//	"reflect"
)

func Rename(expr, oldName, newName string) (ast.Node, *token.FileSet) {
	return RenameAt(expr, oldName, newName, -1, -1)
}

func RenameAt(expr, oldName, newName string, row, col int) (ast.Node, *token.FileSet) {
	contents, fs, _ := ParseExpr(expr)
	visitor := NewRenameVisitor(fs, oldName, newName, row, col)
	ast.Walk(visitor, contents)
	visitor.FinishVisit()
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
	case nil:
		v.AddIdentToScope(i)
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
	case nil:
		v.decls = append(v.decls, name)
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
	//fmt.Printf("%s<%v>\n", tabs(v.tabcount), reflect.TypeOf(v.node))
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
	if v.parent != nil && !contains(v.decls, v.origName) && v.doRename {
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
	//fmt.Printf("%s</%s>\n", tabs(v.tabcount), reflect.TypeOf(v.node))
	switch n := v.node.(type) {
	case nil:
		v.PerformRename()
	case *ast.File:
		v.PerformRename()
	case *ast.Package:
		v.PerformRename()
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
