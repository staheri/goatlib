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
	Conf                 loader.Config
	Root                 *App
	TO                   int // Timeout (seconds)
}


func newApp(name string,path string) *App{
	// Find all go files within the path
	paths,err := filepath.Glob(path+"/*.go")
  check(err)

	//Obtain Name
	appname := appNameFolder(path)

	var conf loader.Config
	if _, err := conf.FromArgs(paths, false); err != nil {
		panic(err)
	}
	return &App{Conf: conf, Name: appname}
}

// take the app conf file
// take the concurrency usage
// inject at those locations

func Instrument(path string){
  // create app
  app := newApp(path)
  concusage := Identify(app)
	for _,c := range(concusage){
		fmt.Println(c.String())
	}
  NewInstrumentedApp(app,concusage)
  // create a new app for instrumented
	//return newPath

}
