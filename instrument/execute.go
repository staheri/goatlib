package instrument

import (
	"bytes"
	"errors"
	"fmt"
  "io"
	_"strings"
	"os"
	"os/exec"
	"log"
  "github.com/staheri/goatlib/trace"
	"io/ioutil"
)

// build from the files in sourceDir
func BuildCommand (sourceDir,exMode,mode string, race bool) string{
	//create binary file
	tmpBinary, err := ioutil.TempFile(sourceDir,"*"+exMode)
  if err != nil {
    fmt.Println("create temp file error")
    panic(err)
  }

	var cmd *exec.Cmd
	if mode == "test"{
		if race{
			cmd = exec.Command("go","test","-race","-c","-o",tmpBinary.Name())
		} else{
			cmd = exec.Command("go","test","-c","-o",tmpBinary.Name())
		}
	}else { // mode = main
		if race{
			cmd = exec.Command("go","build","-race","-o",tmpBinary.Name())
		} else{
			cmd = exec.Command("go","build","-o",tmpBinary.Name())
		}
	}
  var stderr bytes.Buffer
  cmd.Stderr = &stderr
  cmd.Dir = sourceDir

  err = cmd.Run()
  if err != nil {
    fmt.Println("go build error", stderr.String())
    panic(err)
  }
	return tmpBinary.Name()
}


// - executes the instrumented binary
// - Parses collected trace
func ExecuteTrace(binary string) (*trace.ParseResult, error){
	var stderr bytes.Buffer
	var stdout bytes.Buffer
  var cmd *exec.Cmd
	// run
	log.Println("ExecuteTrace: Run ",binary)
	cmd = exec.Command(binary)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout


	if err := cmd.Run(); err != nil {
		fmt.Printf("modified program failed\nErr: %v\nStderr: %v\nStdout: %v\n", err, stderr.String(),stdout.String())
		err1 := errors.New(fmt.Sprintf("%v",err))
		err1 = fmt.Errorf("%v:%v",err,stderr.String())
		return nil, err1
	}

	// check length of stderr
	if stderr.Len() == 0 {
		return nil, errors.New("empty trace")
	}

	// parse
	log.Println("ExecuteTrace: Redirect stderr to ParseTrace ")
	return parseTrace(&stderr, binary)
}

// removes dir
func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Println("Cannot remove temp dir:", err)
	}
}

// reads trace from stderr (io.reader) and parse
func parseTrace(r io.Reader, binary string) (*trace.ParseResult, error) {
	parseResult, err := trace.Parse(r,binary)
	if err != nil {
		return nil, err
	}

	err = trace.Symbolize(parseResult.Events, binary)

	return &parseResult, err
}
