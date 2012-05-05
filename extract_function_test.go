package refactor

import (
	"testing"
	"go/ast"
	//"go/parser"
	"go/token"
	"go/printer"
	"os"
)

func basicLitTypeString(x *ast.BasicLit) ast.Expr {
	name := "unknown"
	if x.Kind == token.STRING {
		name = "string"
	} else if x.Kind == token.INT {
		name = "int"
	} else if x.Kind == token.FLOAT {
		name = "float"
	}
	return &ast.Ident {
		Name: name,
	}
}

func typeString(expr ast.Node) ast.Expr {
	switch x := expr.(type) {
		case *ast.FuncLit:
			return x.Type
		case *ast.BasicLit:
			return basicLitTypeString(x)
		case *ast.AssignStmt:
			return typeString(x.Rhs[0])
		case *ast.CompositeLit:
			return typeString(x.Type)
		case *ast.Ident:
			if x.Obj != nil && x.Obj.Decl != nil {
				return typeString(x.Obj.Decl.(ast.Node))
			} else {
				return &ast.Ident { Name: x.Name }
			}
		case *ast.BinaryExpr:
			return typeString(x.X)
	}
	return nil
}

func ExtractFnFromStmt(name string, stmt ast.Stmt) (ast.Stmt, *ast.FuncDecl) {
    return &ast.AssignStmt{}, 
        &ast.FuncDecl{
            Type: &ast.FuncType {
                Params: nil,
                Results: &ast.FieldList {
                    List: []*ast.Field {
                        &ast.Field {
                            Type: stmt.(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type,
                        },
                    },
                },
            },
        }
}

func ExtractFnFromExpr(name string, expr ast.Expr) (*ast.CallExpr, *ast.FuncDecl) {
	rtnType := typeString(expr)
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
						Type: rtnType, 
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

func StandAloneLiteral(value string, kind token.Token) (ast.Expr) {
	return &ast.BasicLit {
		Kind: kind,
		Value: value,
	}
}
func StandAloneIdent(name, value string, kind token.Token) (*ast.Ident) {
	return  &ast.Ident {
				Name: name,
				Obj: &ast.Object {
					Kind: ast.Var,
					Name: name,
					Decl: &ast.AssignStmt {
						Tok: token.DEFINE,
						Rhs: []ast.Expr {
							&ast.BasicLit {
								Kind: kind,
								Value: value,
							},
						},
					},
				},
			}
}

func TestExtractsFuncFromIntAddExpr(t *testing.T) {
	root := &ast.BinaryExpr {
			X: StandAloneLiteral("1", token.INT),
			Op: token.ADD,
			Y: StandAloneLiteral("2", token.INT),
	}

	_, fn := ExtractFnFromExpr("t", root)
	if fn.Type.Results.List[0].Type.(*ast.Ident).Name != "int" {
		t.Errorf("Expected int, but was %v", fn.Type.Results.List[0].Type.(*ast.Ident).Name)
	}
}

func TestExtractsFuncFromDoubleAddExpr(t *testing.T) {
	root := &ast.BinaryExpr {
			X: StandAloneLiteral("1.0", token.FLOAT),
			Op: token.ADD,
			Y: StandAloneLiteral("2.0", token.FLOAT),
	}

	_, fn := ExtractFnFromExpr("t", root)
	if fn.Type.Results.List[0].Type.(*ast.Ident).Name != "float" {
		t.Errorf("Expected float, but was %v", fn.Type.Results.List[0].Type.(*ast.Ident).Name)
	}
}

func TestExtractsFuncFromDoubleAddIdentExpr(t *testing.T) {
	root := &ast.BinaryExpr {
		X: StandAloneIdent("a", "1", token.FLOAT),
		Op: token.ADD,
		Y: StandAloneIdent("b", "2", token.FLOAT),
	}

	_, fn := ExtractFnFromExpr("t", root)
	if fn.Type.Results.List[0].Type.(*ast.Ident).Name != "float" {
		t.Errorf("Expected float, but was %v", fn.Type.Results.List[0].Type.(*ast.Ident).Name)
	}
}

func TestExtractsFuncFromFuncLit(t *testing.T) {
	root := &ast.FuncLit {
		Type: &ast.FuncType { 
			Params: &ast.FieldList {},
			Results: &ast.FieldList {},
		},
	}

	_, fn := ExtractFnFromExpr("t", root)
	if fn.Type.Results.List[0].Type.(*ast.FuncType) != root.Type {
		t.Errorf("Expected function, but was %v", fn.Type.Results.List[0].Type)
	}
}

func TestExtractFuncFromDecl(t *testing.T) {
    ident := StandAloneIdent("a", "1", token.INT)
	root := &ast.DeclStmt{
        Decl: &ast.GenDecl {
            Tok: token.VAR,
            Specs: []ast.Spec {
                &ast.ValueSpec {
                    Names: []*ast.Ident { ident },
                    Type: ident.Obj.Decl.(*ast.AssignStmt).Rhs[0],
                    Values: []ast.Expr { StandAloneLiteral("2", token.INT) },
                },
            },
        },
	}

	_, fn := ExtractFnFromStmt("t", root)
	if fn.Type.Results.List[0].Type.(*ast.BasicLit) == nil {
		t.Errorf("Expected function, but was %v", fn.Type.Results.List[0].Type)
	}
}
