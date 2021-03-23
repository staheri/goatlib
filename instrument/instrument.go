package instrument

import(
	_"bytes"
	"golang.org/x/tools/go/loader"
	"path/filepath"
	"fmt"
)

// Single Source Target
type App struct{
	Name                 string
	Path                 string
	Conf                 loader.Config
	Root                 *App
	TO                   int // Timeout (seconds)
	IsTest               bool

}


func newApp(appname,path string) *App{
	// Find all go files within the path
	paths,err := filepath.Glob(path+"/*.go")
  check(err)

	var conf loader.Config
	if _, err := conf.FromArgs(paths, false); err != nil {
		panic(err)
	}
	return &App{Path: path, Conf: conf, Name: appname}
}

func Instrument(path string, traceOnly bool) *App{
  // Create app
	// Obtain Name
	//appname := appNameFolder(path)
	appname := gobenchAppNameFolder(path)

  app := newApp(appname,path)
  concusage := Identify(app)
	for _,c := range(concusage){
		fmt.Println(c.String())
	}
	iapp := NewInstrumentedApp(app,concusage,traceOnly)
	iapp.Root = app
  return iapp
}
