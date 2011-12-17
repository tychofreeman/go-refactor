package refactor

import (
	"fmt"
	"go/ast"
	"go/token"
	//"reflect"
)

type RefactorVisitor struct {
	sites *Scope
	parent *RefactorVisitor
	indent int
	targetColumn int
	targetLine int
	gimme chan chan token.Position
	isStmt bool
	isDefn bool
	lhs []ast.Expr
	amLhs bool
	fileSet *token.FileSet
}

func newRefactorVisitor(targetPosition token.Position, sites *Scope, gimme chan chan token.Position, indent int, pw *RefactorVisitor, fileSet *token.FileSet) (visitor *RefactorVisitor) {
	visitor = new(RefactorVisitor)
	visitor.sites = sites
	visitor.targetColumn  = targetPosition.Column
	visitor.targetLine = targetPosition.Line
	visitor.parent = pw
	visitor.indent = indent + 1
	visitor.fileSet = fileSet
	return
}

func (pw *RefactorVisitor) Visit(node ast.Node) (visitor ast.Visitor) {
	if node != nil {
		pw.findIdentifier(node)
	}
	visitor = newRefactorVisitor(token.Position{}, pw.sites, pw.gimme, pw.indent, pw, pw.fileSet)
	return visitor
}

func (visitor *RefactorVisitor) findDeclaringScope(name string) *RefactorVisitor {
	for v := visitor; v != nil ;v = v.parent {
		if _, ok := v.sites.decls[name]; ok {
			return v;
		}
	}
	return nil
}

func (pw *RefactorVisitor) findIdentifier(node ast.Node) (v *RefactorVisitor) {
	//printSpaces(pw)
	//fmt.Printf("Found type %T at %v\n", node, node.Pos())
	_, pw.isStmt = node.(ast.Stmt)
	switch n := node.(type) {
/*		case *ast.Expr:
			if pw.parent.isStmt && len(n) > 0 && len(pw.parent.lhs) > 0 && n[0] == pw.parent.lhs[0] {
				pw.amLhs = true
			}
*/		case *ast.AssignStmt:
			pw.lhs = n.Lhs
			if n.Tok == token.DEFINE {
				pw.isDefn = true
			}
		case *ast.Ident:
			if pw.enclosingStmtIsDefn() && pw.parent.amLhs {
				pw.sites.AddDefn(n.String(), pw.fileSet.Position(n.Pos()))
			}
			pw.sites.AddSite(n.String(), pw.fileSet.Position(n.Pos()))
			if identContainsPosition(n.String(), pw.fileSet.Position(n.Pos()), pw.targetColumn, pw.targetLine) {
				declaringScope := pw.findDeclaringScope(n.String())
				if declaringScope == nil {
					panic(fmt.Sprintf("Could not find declaring scope for %v in %v\n", n.String(), pw))
				}
				for scope := pw; 
				    declaringScope != nil && scope != nil && scope != declaringScope.parent;
				    scope = scope.parent {
					out := make(chan token.Position)
					pw.gimme <- out
					for _,p := range scope.sites.GetSites(n.String()) {
						out <- p
					}
				}
			}
		case *ast.FuncDecl:
			v = newRefactorVisitor(token.Position{ Line: pw.targetLine, Column: pw.targetColumn }, NewChildScope(pw.sites), pw.gimme, pw.indent, pw, pw.fileSet)
			pw.sites = NewChildScope(pw.sites)
	}
	return pw
}

func (pw *RefactorVisitor) enclosingStmtIsDefn() bool {
	parent := pw.parent
	for ; parent != nil && parent.isStmt == false ; parent = parent.parent {
	}
	return parent != nil && parent.isDefn
}

func printSpaces(rv *RefactorVisitor) {
	for i := 0; i < rv.indent; i++ {
		fmt.Printf(" ")
	}
}
