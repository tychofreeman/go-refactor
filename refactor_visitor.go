package refactor

import (
	"fmt"
	"go/ast"
	"go/token"
	//"reflect"
)

type RefactorVisitor struct {
	sites *Scope
	decls []string
	parent *RefactorVisitor
	indent int
	targetColumn int
	targetLine int
	gimme chan token.Position
}

func newRefactorVisitor(targetPosition token.Position, sites *Scope, gimme chan token.Position) (visitor *RefactorVisitor) {
	visitor = new(RefactorVisitor)
	visitor.sites = sites
	visitor.targetColumn  = targetPosition.Column
	visitor.targetLine = targetPosition.Line
	return
}

func (pw *RefactorVisitor) Visit(node interface{}) (visitor ast.Visitor) {
	if node != nil {
		 pw.findIdentifier(node)
	}
	visitor = newRefactorVisitor(token.Position{}, pw.sites, pw.gimme)
	return visitor
}

func (visitor *RefactorVisitor) findDeclaringScope(name string) *RefactorVisitor {
	for v := visitor; ;v = v.parent {
		for d := range visitor.decls {
			if name == visitor.decls[d] {
				return v
			}
		}
	}
	return nil
}

func (pw *RefactorVisitor) findIdentifier(node interface{}) (v *RefactorVisitor){
	switch n := node.(type) {
		case *ast.Ident:
			pw.sites.AddSite(n.String(), n.Pos())
			if identContainsPosition(n.String(), n.Pos(), pw.targetColumn, pw.targetLine) {
				declaringScope := pw.findDeclaringScope(n.String())
				for scope := pw; scope != declaringScope.parent; scope = scope.parent {
					for _,p := range scope.sites.GetSites(n.String()) {
						pw.gimme <- p
					}
				}
				// register symbol name
			}
			// if symbolName == n.String() {
				// spit out site
			//}
		case *ast.FuncDecl:
			v = newRefactorVisitor(token.Position{ Line: pw.targetLine, Column: pw.targetColumn }, NewChildScope(pw.sites), pw.gimme)
			pw.sites = NewChildScope(pw.sites)
			fmt.Printf("-----------------\nFound FuncDecl %v %v\n\n", n.Name, n.Type)
	}
	return pw
}
