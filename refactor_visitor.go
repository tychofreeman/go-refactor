package refactor

import (
	"fmt"
	"go/ast"
	//"reflect"
)

type RefactorVisitor struct {
	sites *VarSites
}

func newRefactorVisitor(sites *VarSites) (visitor *RefactorVisitor) {
	visitor = new(RefactorVisitor)
	visitor.sites = sites
	return
}

func (pw *RefactorVisitor) Visit(node interface{}) (ast.Visitor) {
	if node != nil {
		ident := pw.findIdentifier(node)
		if ident != "" {
			fmt.Printf("Visiting %v\n", ident)
		}
	}
	return pw
}

func (pw *RefactorVisitor) findIdentifier(node interface{}) string {
	switch n := node.(type) {
		case *ast.Ident:
			pw.sites.AddSite(n.String(), n.Pos())
		default:
			//fmt.Printf("Visiting %v (%v)\n", reflect.Typeof(n), n)
			//return fmt.Sprintf("[Ident: %v at '%v' offset:%v line:%v col:%v]", n.String(), n.Filename, n.Offset, n.Line, n.Column)
		/*case *ast.BasicLit:
			return fmt.Sprintf("[BasicLiteral: %v %v]", n.Kind, n.Value)
		case *ast.Object:
			return fmt.Sprintf("[Object: %v %v]", n.Kind, n.Name) */
	}
	return ""
}
