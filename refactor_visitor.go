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
	gimme chan chan token.Position
	isStmt bool
	isDefn bool
	lhs []ast.Expr
	amLhs bool
}

func newRefactorVisitor(targetPosition token.Position, sites *Scope, gimme chan chan token.Position, indent int, pw *RefactorVisitor) (visitor *RefactorVisitor) {
	visitor = new(RefactorVisitor)
	visitor.sites = sites
	visitor.targetColumn  = targetPosition.Column
	visitor.targetLine = targetPosition.Line
	visitor.parent = pw
	visitor.indent = indent + 1
	return
}

func (pw *RefactorVisitor) Visit(node interface{}) (visitor ast.Visitor) {
	if node != nil {
		pw.findIdentifier(node)
	}
	visitor = newRefactorVisitor(token.Position{}, pw.sites, pw.gimme, pw.indent, pw)
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

func (pw *RefactorVisitor) findIdentifier(node interface{}) (v *RefactorVisitor) {
	printSpaces(pw)
	fmt.Printf("Found type %T\n", node)
	_, pw.isStmt = node.(ast.Stmt)
	switch n := node.(type) {
		case []ast.Expr:
			if pw.parent.isStmt && n[0] == pw.parent.lhs[0] {
				pw.amLhs = true
			}
		case *ast.AssignStmt:
			pw.lhs = n.Lhs
			if n.Tok == token.DEFINE {
				fmt.Printf("Definition!\n")
				pw.isDefn = true
			}
		case *token.Token:
			fmt.Printf("Token %v\n", n)
		case *ast.DeclStmt:
			fmt.Printf("Declaration: %v\n", n)
		case []*ast.Ident:
			for ident := range n {
				pw.findIdentifier(ident)
			}
		case *ast.Type:
			fmt.Printf("=============TYPE!!===========\n")
		case *ast.Object:
			fmt.Printf("============OBJECT!!!===========\n")
		case *ast.Ident:
			if pw.enclosingStmtIsDefn() && pw.parent.amLhs {
				pw.sites.AddDefn(n.String(), n.Pos())
				printSpaces(pw)
				fmt.Printf("Identifier %v defined in this stmt!\n", n.String())
			}
			pw.sites.AddSite(n.String(), n.Pos())
			if identContainsPosition(n.String(), n.Pos(), pw.targetColumn, pw.targetLine) {
				declaringScope := pw.findDeclaringScope(n.String())
				for scope := pw; scope != declaringScope.parent; scope = scope.parent {
					out := make(chan token.Position)
					pw.gimme <- out
					for _,p := range scope.sites.GetSites(n.String()) {
						out <- p
					}
				}
			}
		case *ast.FuncDecl:
			v = newRefactorVisitor(token.Position{ Line: pw.targetLine, Column: pw.targetColumn }, NewChildScope(pw.sites), pw.gimme, pw.indent, pw)
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
