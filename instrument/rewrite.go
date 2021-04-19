package instrument

import (
	"bytes"
	"go/ast"
	"go/printer"
	_"go/token"
	"io/ioutil"
	"path/filepath"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/loader"
	"strconv"
	"strings"
)

const MAXPROCS = 4

// add both tracing and delays
func rewrite_randomSched(origpath,newpath string, criticalPoints []*ConcurrencyUsage) []string{
  // Variables
  var astfiles    []*ast.File
  var ret         []string
	var conf        loader.Config
	var concfiles   []string

  // extract aux data
  conclines := make(map[string]int) // extract concurrency lines
  _concfiles := make(map[string]int) // extract concurrency files

  for _,c := range(criticalPoints){
    conclines[c.OrigLoc.String()]=1
    _concfiles[c.OrigLoc.Filename] = 1
  }
  for k,_ := range(_concfiles){
    concfiles = append(concfiles,k)
  }

  // load program files
	paths,err := filepath.Glob(origpath+"/*.go")
	check(err)
	if _, err := conf.FromArgs(paths, false); err != nil {
		panic(err)
	}
  prog, err := conf.Load()
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

  for _,astFile := range(astfiles){

    // check if this file has concurrency usage
    if contains(concfiles,prog.Fset.Position(astFile.Package).Filename){ // add import
      astutil.AddImport(prog.Fset, astFile, "github.com/staheri/goat/goat")
    }
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

      // add gomaxprocs and trace start/stop code
    	ast.Inspect(astFile, func(n ast.Node) bool {
    		switch x := n.(type) {
    		case *ast.FuncDecl:
    			// find 'main' function
    			if x.Name.Name == "main" && x.Recv == nil {
            toAdd := astNode_goatMain(1)
            stmts := []ast.Stmt{toAdd[0]}
            stmts = append(stmts,toAdd[1])
            stmts = append(stmts,toAdd[2])
						stmts = append(stmts,x.Body.List...)
            x.Body.List = stmts
    				return true
    			}else if strings.HasPrefix(x.Name.Name,"Test") && x.Recv == nil{
            toAdd := astNode_goatMain(1)
            stmts := []ast.Stmt{toAdd[0]}
            stmts = append(stmts,toAdd[1])
            stmts = append(stmts,toAdd[2])
						stmts = append(stmts,x.Body.List...)
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
    filename := filepath.Join(newpath, strings.Split(filepath.Base(prog.Fset.Position(astFile.Pos()).Filename),".")[0]+".go")
    err = ioutil.WriteFile(filename, buf.Bytes(), 0666)
    check(err)
    ret = append(ret,filename)
  }
  return ret
}

// add only tracing mechanism to the main function
func rewrite_traceOnly(origpath,newpath string) []string{
  // Variables
  var astfiles    []*ast.File
  var ret         []string
	var conf        loader.Config

  // load program files
	paths,err := filepath.Glob(origpath+"/*.go")
	check(err)
	if _, err := conf.FromArgs(paths, false); err != nil {
		panic(err)
	}
  prog, err := conf.Load()
	check(err)
  for _,crt := range(prog.Created){
    for _,ast := range(crt.Files){
      astfiles = append(astfiles,ast)
    }
  }

  // for main/test:
  //      add (at the beginning) GOAT_done := goat.Start()
	//      add (at the beginning) go goat.Finish(GOAT_done,10)
  //      add (at the end) GOAT_done <- true
  for _,astFile := range(astfiles){
    if mainIn(astFile) || testIn(astFile){
			// add import
			astutil.AddImport(prog.Fset, astFile, "github.com/staheri/goat/goat")
      // add goat start, stop, watch
    	ast.Inspect(astFile, func(n ast.Node) bool {
    		switch x := n.(type) {
    		case *ast.FuncDecl:
    			// find 'main' function
    			if x.Name.Name == "main" && x.Recv == nil {
            toAdd := astNode_goatMain(MAXPROCS)
            stmts := []ast.Stmt{toAdd[0]}
            stmts = append(stmts,toAdd[1])
            stmts = append(stmts,toAdd[2])
						stmts = append(stmts,x.Body.List...)
            x.Body.List = stmts
    				return true
    			}else if strings.HasPrefix(x.Name.Name,"Test") && x.Recv == nil{
            toAdd := astNode_goatMain(MAXPROCS)
            stmts := []ast.Stmt{toAdd[0]}
            stmts = append(stmts,toAdd[1])
            stmts = append(stmts,toAdd[2])
						stmts = append(stmts,x.Body.List...)
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
    filename := filepath.Join(newpath, strings.Split(filepath.Base(prog.Fset.Position(astFile.Pos()).Filename),".")[0]+".go")
    err = ioutil.WriteFile(filename, buf.Bytes(), 0666)
    check(err)
    ret = append(ret,filename)
  }
  return ret
}

// add only delays before critical points (no tracing)
func rewrite_randomSchedOnly(origpath,newpath string, criticalPoints []*ConcurrencyUsage) []string{
  // Variables
  var astfiles    []*ast.File
  var ret         []string
	var conf        loader.Config
	var concfiles   []string

  // extract aux data
  conclines := make(map[string]int) // extract concurrency lines
  _concfiles := make(map[string]int) // extract concurrency files

  for _,c := range(criticalPoints){
    conclines[c.OrigLoc.String()]=1
    _concfiles[c.OrigLoc.Filename] = 1
  }
  for k,_ := range(_concfiles){
    concfiles = append(concfiles,k)
  }

  // load program files
	paths,err := filepath.Glob(origpath+"/*.go")
	check(err)
	if _, err := conf.FromArgs(paths, false); err != nil {
		panic(err)
	}
  prog, err := conf.Load()
	check(err)
  for _,crt := range(prog.Created){
    for _,ast := range(crt.Files){
      astfiles = append(astfiles,ast)
    }
  }

  // for all ast files in the package
  //      add import github.com/staheri/goat/goat
  //      inject goat.Sched_Handler to concurrency lines (astutil.Apply)
  for _,astFile := range(astfiles){

    // check if this file has concurrency usage
    if contains(concfiles,prog.Fset.Position(astFile.Package).Filename){ // add import
      astutil.AddImport(prog.Fset, astFile, "github.com/staheri/goat/goat")
    }

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

      // add gomaxprocs and trace start/stop code
    	ast.Inspect(astFile, func(n ast.Node) bool {
    		switch x := n.(type) {
    		case *ast.FuncDecl:
    			// find 'main' function
    			if x.Name.Name == "main" && x.Recv == nil {
            toAdd := astNode_goatRaceMain(1)
            stmts := []ast.Stmt{toAdd[0]}
            //stmts = append(stmts,toAdd[1])
            //stmts = append(stmts,toAdd[2])
						stmts = append(stmts,x.Body.List...)
            x.Body.List = stmts
    				return true
    			}else if strings.HasPrefix(x.Name.Name,"Test") && x.Recv == nil{
						toAdd := astNode_goatRaceMain(1)
            stmts := []ast.Stmt{toAdd[0]}
						stmts = append(stmts,x.Body.List...)
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
    filename := filepath.Join(newpath, strings.Split(filepath.Base(prog.Fset.Position(astFile.Pos()).Filename),".")[0]+".go")
    err = ioutil.WriteFile(filename, buf.Bytes(), 0666)
    check(err)
    ret = append(ret,filename)
  }
  return ret
}
