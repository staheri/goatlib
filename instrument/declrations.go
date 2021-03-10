package instrument

import (
	"go/ast"
	"go/token"
	"strconv"
)

// returns trace statements trace statments
func astDecl_traceStmts(timeout int) []ast.Stmt {
	ret := make([]ast.Stmt, 2)

	// trace.Start(os.Stderr)
	ret[0] = &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "trace"},
				Sel: &ast.Ident{Name: "Start"},
			},
			Args: []ast.Expr{
				&ast.SelectorExpr{
					X:   &ast.Ident{Name: "os"},
					Sel: &ast.Ident{Name: "Stderr"},
				},
			},
		},
	}
	if timeout > 0{
		// go func(){ <-time.After(5 * time.Second) trace.Stop() os.Exit(1) }()
		ret[1] = &ast.GoStmt{
			Call: &ast.CallExpr{
				Fun: &ast.FuncLit{
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "time"},
										Sel: &ast.Ident{Name: "Sleep"},
									},
									Args: []ast.Expr{
										&ast.BinaryExpr{
											X:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(timeout)},
											Op: token.MUL,
											Y: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "time"},
												Sel: &ast.Ident{Name: "Second"},
											},
										},
									},
								},
							},
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "trace"},
										Sel: &ast.Ident{Name: "Stop"},
									},
								},
							},
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "os"},
										Sel: &ast.Ident{Name: "Exit"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.INT, Value: "0"},
									},
								},
							},
						},
					},
					Type: &ast.FuncType{Params: &ast.FieldList{}},
				},
			},
		}
	} else{
		// defer func(){ time.Sleep(50*time.Millisecond; trace.Stop() }()
		ret[1] = &ast.DeferStmt{
			Call: &ast.CallExpr{
				Fun: &ast.FuncLit{
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "time"},
										Sel: &ast.Ident{Name: "Sleep"},
									},
									Args: []ast.Expr{
										&ast.BinaryExpr{
											X:  &ast.BasicLit{Kind: token.INT, Value: "50"},
											Op: token.MUL,
											Y: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "time"},
												Sel: &ast.Ident{Name: "Millisecond"},
											},
										},
									},
								},
							},
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "trace"},
										Sel: &ast.Ident{Name: "Stop"},
									},
								},
							},
						},
					},
					Type: &ast.FuncType{Params: &ast.FieldList{}},
				},
			},
		}
	}

	return ret
}

// returns a general declration representing a constant node
func astDecl_constNode(name, value string) *ast.GenDecl {
	return &ast.GenDecl{
		Tok:token.CONST,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{
					&ast.Ident{Name: name},
				},
				Values: []ast.Expr{
					&ast.BasicLit{Kind: token.INT, Value: value},
				},
			},
		},
	}
}

// returns sharedInt type structure node
func astDecl_structNode() *ast.GenDecl {
	return &ast.GenDecl{
		Tok:token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{ Name: "goatRS_sharedInt",},
				Type: &ast.StructType{
					Fields:&ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{&ast.Ident{Name: "n"}},
								Type: &ast.Ident{Name: "int"},
							},
							&ast.Field{
								Type: &ast.SelectorExpr{
									X: &ast.Ident{Name: "sync"},
									Sel: &ast.Ident{Name: "Mutex"},
								},
							},
						},
					},
				},
			},
		},
	}
}

// returns a global instance of sharedInt node
func astDecl_globalCount() *ast.GenDecl{
	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{&ast.Ident{Name: "goatRS_cnt"}},
				Type: &ast.Ident{Name: "goatRS_sharedInt"},
			},
		},
	}
}


// returns GOMAXPROCS line node
func astDecl_goMaxProcs(max int) ast.Stmt{
	//ret := make([]ast.Stmt, 1)
	ret := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "runtime"},
				Sel: &ast.Ident{Name: "GOMAXPROCS"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(max)},
			},
		},
	}
	return ret
}

// returns "if Reschedule then Gosched()" line node
func astDecl_callFuncSched() *ast.IfStmt{
	return &ast.IfStmt{
		Cond: &ast.CallExpr{
			Fun: &ast.Ident{Name: "Reschedule"},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "runtime"},
							Sel: &ast.Ident{Name: "Gosched"},
						},
					},
				},
			},
		},
	}
}

// returns Reschedule() delration node
func astDecl_declFuncSched() *ast.FuncDecl{
	return &ast.FuncDecl{
		Name: &ast.Ident{Name: "Reschedule"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident{Name: "bool"},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{ // random seed generator
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "rand"},
							Sel: &ast.Ident{Name: "Seed"},
						},
						Args: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "time"},
											Sel: &ast.Ident{Name: "Now"},
										},
									},
									Sel: &ast.Ident{Name: "UnixNano"},
								},
							},
						},
					},
				},
				&ast.IfStmt{ // main if
					Cond: &ast.BinaryExpr{ // coint toss
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "rand"},
								Sel: &ast.Ident{Name: "Intn"},
							},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.INT, Value: "2"},
							},
						},
						Y: &ast.BasicLit{Kind: token.INT, Value: "1"},
						Op: token.EQL,
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{ // lock
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "goatRS_cnt"},
										Sel: &ast.Ident{Name: "Lock"},
									},
								},
							},
							&ast.DeferStmt{// defer unlock
								Call: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "goatRS_cnt"},
										Sel: &ast.Ident{Name: "Unlock"},
									},
								},
							},
							&ast.IfStmt{// if
								Init: &ast.AssignStmt{
									Tok: token.DEFINE,
									Lhs: []ast.Expr{
										&ast.Ident{Name: "d"},
										&ast.Ident{Name: "err"},
									},
									Rhs:[]ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "strconv"},
												Sel: &ast.Ident{Name: "Atoi"},
											},
											Args: []ast.Expr{
												&ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "os"},
														Sel: &ast.Ident{Name: "Getenv"},
													},
													Args: []ast.Expr{
														&ast.BasicLit{Kind: token.STRING, Value: "\"GOATRSBOUND\""},
													},
												},
											},
										},
									},
								},
								Cond: &ast.BinaryExpr{
									X: &ast.BinaryExpr{
										X: &ast.Ident{Name: "err"},
										Y: &ast.Ident{Name: "nil"},
										Op: token.EQL,
									},
									Y: &ast.BinaryExpr{
										X: &ast.SelectorExpr{
											X: &ast.Ident{Name: "goatRS_cnt"},
											Sel: &ast.Ident{Name: "n"},
										},
										Y:&ast.Ident{Name: "d"},
										Op: token.LSS,
									},
									Op: token.LAND,
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.IncDecStmt{
											X: &ast.SelectorExpr{
												X: &ast.Ident{Name: "goatRS_cnt"},
												Sel: &ast.Ident{Name: "n"},
											},
											Tok: token.INC,
										},
										&ast.ReturnStmt{
											Results: []ast.Expr{
												&ast.Ident{Name: "true"},
											},
										},
									},
								},
								Else: &ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.Ident{Name: "false"},
									},
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{Name: "false"},
					},
				},
			},
		},
	}
}

// wrapper for new declrations
func astDecl_newDecls(depth int) []ast.Decl {
	ret := make([]ast.Decl,3)
	//ret[0] = astDecl_constNode("depth",strconv.Itoa(depth))
	ret[0] = astDecl_structNode()
	ret[1] = astDecl_globalCount()
	ret[2] = astDecl_declFuncSched()
	return ret
}

func astDecl_convertDefer(def *ast.DeferStmt) *ast.DeferStmt{
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						astDecl_callFuncSched(),
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
