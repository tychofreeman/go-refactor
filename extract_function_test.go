package refactor

import (
	"testing"
	"go/ast"
	//"go/parser"
	"go/token"
	"go/printer"
	"os"
	"fmt"
)

func ExtractFnFromExpr(name string, expr ast.Expr) (*ast.CallExpr, *ast.FuncDecl) {
	rtnType := "unknown"
	switch x := expr.(type) {
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				rtnType = "string"
			} else if x.Kind == token.INT {
				rtnType = "int"
			}
			break;
		case *ast.Ident:
			fmt.Printf("--------------------------\nIdent: %v\n", x)
			rtnType = x.Obj.Decl.(*ast.AssignStmt).Rhs[0].(*ast.CompositeLit).Type.(*ast.Ident).Name

	}
	return &ast.CallExpr{
		Fun: &ast.Ident{Name: name},
	}, 
	&ast.FuncDecl{
		Name: &ast.Ident{
			Name: name,
			Obj: &ast.Object{
				Kind: ast.Fun,
				Name: name,
			},
		},

		Type: &ast.FuncType{
			Params: &ast.FieldList{
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident {
							Name: rtnType,
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						expr,
					},
				},
			},
		},
	}
}

func TestExtractsLiteral(t *testing.T) {
	var root ast.BasicLit
	root.Kind = token.INT
	root.Value = "0"

	replace, fn := ExtractFnFromExpr("t", &root)
	
	if replace.Fun.(*ast.Ident).Name != "t" {
		t.Fail()
	}

	if fn.Recv != nil {
		t.Fail()
	}

	if fn.Name.Name != "t" {
		t.Fail()
	}

	if len(fn.Body.List) != 1 {
		t.Fail()
	}

	switch x := fn.Body.List[0].(type) {
		case *ast.ReturnStmt:
			if len(x.Results) != 1 {
				t.Fail()
			} else {
				if x.Results[0].(*ast.BasicLit).Kind != token.INT {
					t.Fail()
				}
				if x.Results[0].(*ast.BasicLit).Value != "0" {
					t.Fail()
				}
			}
		default:
			t.Fail()
	}

	printer.Fprint(os.Stdout, token.NewFileSet(), fn)
	printer.Fprint(os.Stdout, token.NewFileSet(), replace)
}

func TestExtractsOtherLiteral(t *testing.T) {
	root := &ast.BasicLit {
		Kind: token.STRING,
		Value: "test-string",
	}

	_, fn := ExtractFnFromExpr("t", root)
	if fn.Type.Results.List[0].Type.(*ast.Ident).Name != "string" {
		t.Errorf("Expected string, but was %s", fn.Type.Results.List[0].Type.(*ast.Ident).Name )
	}
}

func TestExtractsFuncFromIdent(t *testing.T) {
	root := &ast.Ident {
		Name: "a",
		Obj: &ast.Object {
			Kind: ast.Var,
			Name: "a",
			Decl: &ast.AssignStmt{
				Tok: token.DEFINE,
				Rhs: []ast.Expr {
					&ast.CompositeLit {
						Type: &ast.Ident{
							Name: "A",
						},
					},
				},
			},
		},
	}

	_, fn := ExtractFnFromExpr("t", root)
	if fn.Type.Results.List[0].Type.(*ast.Ident).Name != "A" {
		t.Error("Expected 'A', but was ", fn.Type.Results.List[0].Type)
	}
}
