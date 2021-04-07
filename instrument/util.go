package instrument

import (
	"strings"
	"go/ast"
	"fmt"
)

func check(err error){
	if err != nil{
		panic(err)
	}
}

// If s contains e
func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

// Returns appName from goBench folders (omitting forbidden chars for database)
func GobenchAppNameFolder(path string) string{
  if !strings.HasSuffix(path,"/"){
    path = path + "/"
  }
  a := strings.Split(path,"/")
	ret := a[len(a)-4]+"_"+a[len(a)-3]+"_"+a[len(a)-2]
  b := strings.Split(ret,".")
	s := ""
	for i:=0;i<len(b);i++{
		if i == len(b) - 1{
			s = s + b[i]
		}else{
			s = s + b[i]+"_"
		}
	}
	ret = ""
	for _,b := range s{
		if string(b) == "-"{
			ret = ret + "_"
		} else{
			ret = ret + string(b)
		}
	}
	return ret
}


// Returns appName from long paths (omitting forbidden chars for database)
func appNameFolder(path string) string{
  if !strings.HasSuffix(path,"/"){
    path = path + "/"
  }
  a := strings.Split(path,"/")
	fmt.Println(a)
  b := strings.Split(a[len(a)-2],".")
	fmt.Println(b)
	s := ""
	ret := ""
	for i:=0;i<len(b);i++{
		if i == len(b) - 1{
			s = s + b[i]
		}else{
			s = s + b[i]+"_"
		}
	}
	for _,b := range s{
		if string(b) == "-"{
			ret = ret + "_"
		} else{
			ret = ret + string(b)
		}
	}
	return ret
}

// Returns appName from long paths (omitting forbidden chars for database)
func appNameSingleSource(app string) string{
	a := strings.Split(app,"/")
	b := strings.Split(a[len(a)-1],".")
	s := ""
	ret := ""
	for i:=0;i<len(b)-1;i++{
		if i == len(b) - 2{
			s = s + b[i]
		}else{
			s = s + b[i]+"_"
		}
	}
	for _,b := range s{
		if string(b) == "-"{
			ret = ret + "_"
		} else{
			ret = ret + string(b)
		}
	}
	return ret
}


// Tells if the ast file has main in it
func mainIn(root ast.Node) bool{
	ret := false
	ast.Inspect(root, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// find 'main' function
			if x.Name.Name == "main" && x.Recv == nil {
				ret = true
				return true
			}
		}
		return true
	})
	return ret
}


// Tells if the ast file has test in it
func testIn(root ast.Node) bool{
	ret := false
	ast.Inspect(root, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// find 'main' function
			if strings.HasPrefix(x.Name.Name,"Test") {
				ret = true
				return true
			}
		}
		return true
	})
	return ret
}
