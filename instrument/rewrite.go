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

  return iapp
}

func (app *App) rewrite_randomSched(path string, concusage []*ConcurrencyUsage, depth int) []string{
  // Variables
  var astfiles    []*ast.File
  var ret         []string
  // extract concurrency lines
  conclines := make(map[string]int)
  for _,c := range(concusage){
    conclines[c.OrigLoc.String()]=1
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
  for _,astFile := range(astfiles){
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
          cr.InsertBefore(astDecl_callFuncSched())
          return true
        }else{
          return true
        }
      } else{
        return true
      }

    },nil)

    // add other injections only to main/test file
    if mainIn(astFile) || testIn(astFile){
      astutil.AddImport(prog.Fset, astFile, "os")
    	astutil.AddImport(prog.Fset, astFile, "runtime/trace")
    	astutil.AddImport(prog.Fset, astFile, "time")
    	if app.TO > 0{
    		astutil.AddNamedImport(prog.Fset, astFile, "_", "net")
    	}
      astutil.AddImport(prog.Fset, astFile, "sync")
    	astutil.AddImport(prog.Fset, astFile, "runtime")
    	astutil.AddImport(prog.Fset, astFile, "math/rand")
      // add constant, struct type, global counter, function declration
      ast.Inspect(astFile, func(n ast.Node) bool {
  			switch x := n.(type) {
  			case *ast.File:
  				// add constant, struct type, global counter, function declration
  				decls := astDecl_newDecls(depth)
  				decls2 := x.Decls
  				decls = append(decls2,decls...)
  				x.Decls = decls
  				return true
  			}
  			return true
  		})
      // add gomaxprocs and trace start/stop code
    	ast.Inspect(astFile, func(n ast.Node) bool {
    		switch x := n.(type) {
    		case *ast.FuncDecl:
    			// find 'main' function
    			if x.Name.Name == "main" && x.Recv == nil {
    				stmts := astDecl_traceStmts(app.TO)
            stmts = append(stmts,astDecl_goMaxProcs())
    				stmts = append(stmts, x.Body.List...)
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
    filename := filepath.Join(path, strings.Split(filepath.Base(prog.Fset.Position(astFile.Pos()).Filename),".")[0]+"_inst.go")
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
