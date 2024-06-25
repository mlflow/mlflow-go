package main

import (
	"go/ast"
	"go/token"
)

func mkImportSpec(value string) *ast.ImportSpec {
	return &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: value}}
}

func mkImportStatements(importStatements ...string) ast.Decl {
	specs := make([]ast.Spec, 0, len(importStatements))

	for _, importStatement := range importStatements {
		specs = append(specs, mkImportSpec(importStatement))
	}

	return &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: specs,
	}
}

func mkStarExpr(e ast.Expr) *ast.StarExpr {
	return &ast.StarExpr{
		X: e,
	}
}

func mkSelectorExpr(x, sel string) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: ast.NewIdent(x), Sel: ast.NewIdent(sel)}
}

func mkNamedField(name string, typ ast.Expr) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  typ,
	}
}

func mkField(typ ast.Expr) *ast.Field {
	return &ast.Field{
		Type: typ,
	}
}

// fun(arg1, arg2, ...)
func mkCallExpr(fun ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  fun,
		Args: args,
	}
}

// Shorthand for creating &expr.
func mkAmpExpr(expr ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{
		Op: token.AND,
		X:  expr,
	}
}

// err != nil.
var errNotEqualNil = &ast.BinaryExpr{
	X:  ast.NewIdent("err"),
	Op: token.NEQ,
	Y:  ast.NewIdent("nil"),
}

// return err.
var returnErr = &ast.ReturnStmt{
	Results: []ast.Expr{ast.NewIdent("err")},
}

func mkBlockStmt(stmts ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: stmts,
	}
}

func mkIfStmt(init ast.Stmt, cond ast.Expr, body *ast.BlockStmt) *ast.IfStmt {
	return &ast.IfStmt{
		Init: init,
		Cond: cond,
		Body: body,
	}
}

func mkAssignStmt(lhs, rhs []ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: lhs,
		Tok: token.DEFINE,
		Rhs: rhs,
	}
}

func mkReturnStmt(results ...ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: results,
	}
}
