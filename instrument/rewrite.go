package instrument

import (
	"bytes"
	"go/ast"
	"go/printer"
	_"go/token"
	"io/ioutil"
	"path/filepath"
	"golang.org/x/tools/go/ast/astutil"
	_"log"
	"strconv"
	_"reflect"
	"fmt"
	"strings"
  "os"
)

func NewInstrumentedApp(app *App, concusage []*ConcurrencyUsage) (iapp *App){
  var newPath,newName string
  newName =app.Name+"_INST"
  // create placeholder folder for instrumented app
  if ws := os.Getenv("GOATWS");ws!="" {
    newPath = ws+"/"+newName
    err := os.MkdirAll(newPath,os.ModePerm)
  	check(err)
  }else{
    panic("GOATWS is not set!")
  }
  // we can rewrite differently each time
  app.rewrite_randomSched(newPath,concusage,2)

  iapp = newApp(newName,newPath)
  iapp.IsTest = app.IsTest
  return iapp
}

func (app *App) rewrite_randomSched(path string, concusage []*ConcurrencyUsage, depth int) []string{
  // Variables
  var astfiles    []*ast.File
  var ret         []string

  // extract aux data
  conclines := make(map[string]int) // extract concurrency lines
  _concfiles := make(map[string]int) // extract concurrency files
  var concfiles []string
  for _,c := range(concusage){
    conclines[c.OrigLoc.String()]=1
    _concfiles[c.OrigLoc.Filename] = 1
  }
  for k,_ := range(_concfiles){
    concfiles = append(concfiles,k)
  }

  // load program files
  prog, err := app.Conf.Load()
	check(err)
  for _,crt := range(prog.Created){
    for _,ast := range(crt.Files){
      astfiles = append(astfiles,ast)
    }
  }

  // for all ast files in the package
  //      add import github.com/staheri/goat/goat
  //      inject goat.Sched_Handler to concurrency lines (astutil.Apply)
  //      inject goat.Sched_Handler to range (astutil.Inspect)
  // for main/test:
  //      add (at the beginning) GOAT_done := goat.Start()
	//      add (at the beginning) go goat.Finish(GOAT_done,10)
  //      add (at the end) GOAT_done <- true
  //
  // then inject sched calls to concusage (through astutil.Apply - all files )
  // then inject sched calls to range (through astutil.Inspect - all files )
  // then inject imports (through astutil.AddImport - main/test)
  // then inject tracing and gomaxprocs (through ast.Inspect - main/test)
  for _,astFile := range(astfiles){

    // check if this file has concurrency usage
    if contains(concfiles,prog.Fset.Position(astFile.Package).Filename){ // add import
      astutil.AddImport(prog.Fset, astFile, "github.com/staheri/goat/goat")
    }
		// 
		// if mainIn(astFile) || testIn(astFile){
		// 	astutil.AddImport(prog.Fset, astFile, "runtime/trace")
		// }

    // add schedcalls wherever concusage
    astutil.Apply(astFile, func(cr *astutil.Cursor) bool{
      n := cr.Node()
      if n != nil{
        curloc := prog.Fset.Position(n.Pos()).Filename+":"+strconv.Itoa(prog.Fset.Position(n.Pos()).Line)
        if _,ok := conclines[curloc];ok{
          if conclines[curloc] != 1{
            return true
          }
          conclines[curloc] = 2
          // point of injection
          switch x:= n.(type){
          case *ast.DeferStmt:
            ds := n.(*ast.DeferStmt)
            cr.Replace(astNode_convertDefer(ds))
            _ = x
            return true
          }
          cr.InsertBefore(astNode_sched())
          return true
        }
      }
      return true
    },nil)

    // for range statement, all concusage
    ast.Inspect(astFile, func(n ast.Node) bool {
      switch x := n.(type){
      case *ast.RangeStmt:
        newCall := astNode_sched()
        x.Body.List = append(x.Body.List,newCall)
        return true
      }
      return true
    })

    // add other injections only to main/test file
    if mainIn(astFile) || testIn(astFile){
      if testIn(astFile){
        app.IsTest = true
      }

      // add gomaxprocs and trace start/stop code
    	ast.Inspect(astFile, func(n ast.Node) bool {
    		switch x := n.(type) {
    		case *ast.FuncDecl:
    			// find 'main' function
    			if x.Name.Name == "main" && x.Recv == nil {
            toAdd := astNode_goatMain()
            stmts := []ast.Stmt{toAdd[0]}
            stmts = append(stmts,toAdd[1])
            stmts = append(stmts,toAdd[2])
						stmts = append(stmts,x.Body.List...)
            //stmts = append(stmts,toAdd[3])
						//stmts = append(stmts,toAdd[4])
            x.Body.List = stmts
    				return true
    			}else if strings.HasPrefix(x.Name.Name,"Test") && x.Recv == nil{
            toAdd := astNode_goatMain()
            stmts := []ast.Stmt{toAdd[0]}
            stmts = append(stmts,toAdd[1])
            stmts = append(stmts,toAdd[2])
						stmts = append(stmts,x.Body.List...)
            //stmts = append(stmts,toAdd[3])
						//stmts = append(stmts,toAdd[4])
            x.Body.List = stmts
    				return true
          }
    		}
    		return true
    	})
    } // end for main

    // write files
    var buf bytes.Buffer
  	err := printer.Fprint(&buf, prog.Fset, astFile)
  	check(err)
    filename := filepath.Join(path, strings.Split(filepath.Base(prog.Fset.Position(astFile.Pos()).Filename),".")[0]+".go")
    fmt.Println("AST Name",filename)
    fmt.Println("App Name",app.Name)
    err = ioutil.WriteFile(filename, buf.Bytes(), 0666)
    check(err)
    ret = append(ret,filename)
  }
  return ret

	/*
	}*/
}
