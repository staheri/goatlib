package instrument

import (
	"go/ast"
	"go/token"
)

// returns "if Reschedule then Gosched()" line node
func astNode_sched() *ast.ExprStmt{
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "goat"},
				Sel: &ast.Ident{Name: "Sched_Handler"},
			},
		},
	}
}


// wrapper for new declrations
func astNode_goatMain() []ast.Stmt {
	ret := make([]ast.Stmt,3)
	ret[0] = astNode_goatStart()
	ret[1] = astNode_goatWatch()
	ret[2] = astNode_goatStop()
	//ret[3] = astNode_goatDoneAck()
	//ret[4] = astNode_goatStopTrace()
	return ret
}

//ast node for "GOAT_done <- 0"
func astNode_goatDone() *ast.SendStmt{
	return &ast.SendStmt{
		Chan: &ast.Ident{Name: "GOAT_done"},
		Value: &ast.Ident{Name: "true"},
	}
}

func astNode_goatDoneAck() *ast.ExprStmt{
	return &ast.ExprStmt{
		X: &ast.UnaryExpr{
			X: &ast.Ident{Name: "GOAT_done"},
			Op: token.ARROW,
		},

	}
}

func astNode_goatStopTrace() *ast.ExprStmt{
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "trace"},
				Sel: &ast.Ident{Name: "Stop"},
			},
		},
	}
}

// ast node for "GOAT_done := goat.Start()"
func astNode_goatStart() *ast.AssignStmt{
	return &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: "GOAT_done"},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "goat"},
					Sel: &ast.Ident{Name: "Start"},
				}, // args could be added here
			},
		},
	}
}

// ast node for "go goat.Finish(GOAT_done)"
func astNode_goatWatch() *ast.GoStmt{
	return &ast.GoStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "goat"},
				Sel: &ast.Ident{Name: "Watch"},
			},
			Args : []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: "GOAT_done"},
			},
		},
	}
}



func astNode_convertDefer(def *ast.DeferStmt) *ast.DeferStmt{
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						astNode_sched(),
						&ast.ExprStmt{
							X: def.Call,
						},
					},
				},
				Type: &ast.FuncType{Params: &ast.FieldList{}},
			},
		},
	}
}


func astNode_goatStop() *ast.DeferStmt{
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "goat"},
				Sel: &ast.Ident{Name: "Stop"},
			},
			Args : []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: "GOAT_done"},
			},
		},
	}
}
