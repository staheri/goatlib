package instrument

import(
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"strconv"
	"golang.org/x/tools/go/loader"
	"path/filepath"
)

type ConcurrencyUsage struct{
	Type              int                   `json:"type"`
	Location          *CodeLocation         `json:"location"`
}

func newConcurrencyUsage(typ int, codeloc *CodeLocation) *ConcurrencyUsage{
	return &ConcurrencyUsage{Type: typ, Location:codeloc}
}

func (cl *ConcurrencyUsage) String() string{
	//return concTypeDescription[cl.Type]+" - " +cl.OrigLoc.Filename+":"+strconv.Itoa(cl.OrigLoc.Line)
	return cl.Location.FileName+":"+strconv.Itoa(cl.Location.Line)+"("+ConcTypeDescription[cl.Type]+")"
}

type CodeLocation struct{
	FileName              string          `json:"fileName"`
	Function              string          `json:"function,omitempty"`   // will be empty in static instrumentation, will be updated duirng dynamic executions
	Line                  int             `json:"line"`
}

func newCodeLocation(file string, line int) *CodeLocation {
	return &CodeLocation{FileName: file, Line: line}
}

func (cl *CodeLocation) String() string{
	return cl.FileName+":"+strconv.Itoa(cl.Line)
}


// Identify Concurrency Usage locations
func Identify(path string) []*ConcurrencyUsage{
	// Vars
	var astfiles             []*ast.File
	var commclauses          []string
	var codeloc              *CodeLocation
	var concusage            []*ConcurrencyUsage
	var conf                 loader.Config

  // load program files
	paths,err := filepath.Glob(path+"/*.go")
	check(err)
	if _, err := conf.FromArgs(paths, false); err != nil {
		panic(err)
	}
  prog, err := conf.Load()

	for _,crt := range(prog.Created){
		for _,ast := range(crt.Files){
			astfiles = append(astfiles,ast)
		}
	}

	// Identify Concurrency Usage Locations
	for _,astFile := range(astfiles){
		astutil.Apply(astFile, func(cr *astutil.Cursor) bool{
			// Current Node
			n := cr.Node()
			if n != nil{
				codeloc = newCodeLocation(
					prog.Fset.Position(n.Pos()).Filename,
					prog.Fset.Position(n.Pos()).Line,
				)
				switch x:= n.(type){

				// Go keyword (GoCreate),
				case *ast.GoStmt:
					concusage = append(concusage,newConcurrencyUsage(GO,codeloc))
					return true// return false so it does not traverse childrens

				// Select statment (Select)
				case *ast.SelectStmt:
					ss := n.(*ast.SelectStmt)
					if len(ss.Body.List) > 2{
						concusage = append(concusage,newConcurrencyUsage(SELECT,codeloc))
						return true
					} else if len(ss.Body.List) == 2{
						cas0,ok := ss.Body.List[0].(*ast.CommClause)
						if !ok{
							return true
						}
						cas1,ok := ss.Body.List[1].(*ast.CommClause)
						if !ok{
							return true
						}

						if cas0.Comm == nil{
							//fmt.Printf("non-def case: %v\n",cas1.Pos())
							codeloc2 := newCodeLocation(
								prog.Fset.Position(cas1.Pos()).Filename,
								prog.Fset.Position(cas1.Pos()).Line,
							)
							concusage = append(concusage,newConcurrencyUsage(NBCASE,codeloc2))
							concusage = append(concusage,newConcurrencyUsage(NBSELECT,codeloc))
							return true
						}
						if cas1.Comm == nil{
							codeloc2 := newCodeLocation(
								prog.Fset.Position(cas0.Pos()).Filename,
								prog.Fset.Position(cas0.Pos()).Line,
							)
							concusage = append(concusage,newConcurrencyUsage(NBCASE,codeloc2))
							concusage = append(concusage,newConcurrencyUsage(NBSELECT,codeloc))
							return true
						}
						return true
					}
					return true


				// Selector (Mu(RW)Lock, Mu(RW)Unlock, (Cv/Wg)Wait, WgAdd, CvSignal, CvBroadcast)
				case *ast.Ident:
					_,ok := cr.Parent().(*ast.SelectorExpr)
					ident := n.(*ast.Ident)
					if ok && contains(selectorIdents,ident.String()){
						switch ident.String(){
						case "Lock":
							concusage = append(concusage,newConcurrencyUsage(LOCK,codeloc))
						case "Unlock":
							concusage = append(concusage,newConcurrencyUsage(UNLOCK,codeloc))
            case "RLock":
							concusage = append(concusage,newConcurrencyUsage(RLOCK,codeloc))
						case "RUnlock":
							concusage = append(concusage,newConcurrencyUsage(RUNLOCK,codeloc))
						case "Add":
							concusage = append(concusage,newConcurrencyUsage(ADD,codeloc))
						case "Done":
							if !contains(commclauses,codeloc.String()){ // for select
								concusage = append(concusage,newConcurrencyUsage(DONE,codeloc))
							}
						case "Signal":
							concusage = append(concusage,newConcurrencyUsage(SIGNAL,codeloc))
						case "Wait":
							concusage = append(concusage,newConcurrencyUsage(WAIT,codeloc))
						case "Broadcast":
							concusage = append(concusage,newConcurrencyUsage(BROADCAST,codeloc))
						}
						return false
					}

					// Close (ChClose)
					_,ok = cr.Parent().(*ast.CallExpr)
					if ok && ident.String() == "close"{
						concusage = append(concusage,newConcurrencyUsage(CLOSE,codeloc))
						return false
					}
					return true

				// ChSend
				case *ast.SendStmt:
					if cr.Index() >= 0{
						concusage = append(concusage,newConcurrencyUsage(SEND,codeloc))
						return false
					}
					return true

				// Store CommClause (select case) to ignore
				case *ast.CommClause:
					commclauses = append(commclauses,codeloc.String())
					return true

        case *ast.AssignStmt:
					as := n.(*ast.AssignStmt)
					asrhs := as.Rhs
          for _,expr := range(asrhs){
            switch y := expr.(type){
  						case *ast.UnaryExpr:
  							ux := expr.(*ast.UnaryExpr)
  							if !contains(commclauses,codeloc.String()) && ux.Op == token.ARROW{
  								concusage = append(concusage,newConcurrencyUsage(RECV,codeloc))
  								return false
  							}
  							return true
  						default:
  							_ = y
  							return true
  					}
          }

				// recvs in return
				case *ast.ReturnStmt:
					rs := n.(*ast.ReturnStmt)
					rsres := rs.Results
					for _,expr := range(rsres){
						switch y := expr.(type){
							case *ast.UnaryExpr:
								ux := expr.(*ast.UnaryExpr)
								if !contains(commclauses,codeloc.String()) && ux.Op == token.ARROW{
									concusage = append(concusage,newConcurrencyUsage(RECV,codeloc))
									return false
								}
								return true
							default:
								_ = y
								return true
						}
					}


				// ChRecv (with no assignment)
				case *ast.ExprStmt:
					es := n.(*ast.ExprStmt)
					esx := es.X
					switch y := esx.(type){
						case *ast.UnaryExpr:
							ux := esx.(*ast.UnaryExpr)
							if !contains(commclauses,codeloc.String()) && ux.Op == token.ARROW{
								concusage = append(concusage,newConcurrencyUsage(RECV,codeloc))
								return false
							}
							return true
						default:
							_ = y
							return true
					}

				// Range (ChRecv)
				case *ast.BlockStmt:
					p,ok := cr.Parent().(*ast.RangeStmt)
					if ok{
						// For position
						codeloc = newCodeLocation(
							prog.Fset.Position(p.For).Filename,
							prog.Fset.Position(p.For).Line,
						)
						concusage = append(concusage,newConcurrencyUsage(RANGE,codeloc))
						return true
					}
					return true

				// All other
				default:
					_ = x
					return true
				}
			}
			return true
		},nil)
	}
	return concusage
}


// we need to find these

// x.Lock()      ExprStmt CallExpr SelectorExpr ident ident
// x.Unlock()
// x.Wait()
// x.Add()
// x.Signal()
// x.Broadcast()
// ch <- x      = SendStmt Ident Ident
// x := <- ch    = AssignStmt Ident UnaryExpr Ident
//   <- x        = ExprStmt UnaryExpr Ident
// select

// make(chan int)   = ast.AssigStmt (Ident(channelName),CallExpr,Ident(make),ChanType,ident (int))
// close(ch)        = ExprStmt CallExpr Ident Ident


// go func(){}...     =    GoStmt CallExpr FuncLit FuncType FieldList BlockStmt{}
// go new()           =  GoStmt CallExpr Ident


// var m1 sync.Mutex    =     DeclStmt , GenDecl, ValueSpec, Ident SelectorExpr Ident Ident


// x.x.lock()       = ExprStmt CallExpr SelectorExpr SelectorExpr Ident Ident Ident


// select (multiple cases)
// = SelectStmt {BlockStmt}  case1 : CommClause(case) ExprStmt UnaryExpr Ident
//                           case2 : CommClause(case) SendStmt Ident BasicLit
//                           default: CommClause


// range channel
// *ast.RangeStmt *ast.Ident  *ast.Ident  *ast.BlockStmt

// range list
// *ast.RangeStmt *ast.Ident *ast.Ident *ast.ParenExpr *ast.Ident *ast.BlockStmt

// for
// *ast.ForStmt *ast.AssignStmt *ast.Ident *ast.BasicLit *ast.BinaryExpr *ast.Ident *ast.BasicLit *ast.IncDecStmt *ast.Ident  *ast.BlockStmt
