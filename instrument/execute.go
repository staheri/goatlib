package instrument

import (
	"bytes"
	"errors"
	"fmt"
	"time"
	"os"
	"os/exec"
	"log"
	"io/ioutil"
	"path/filepath"
)

type ExecuteResult struct{
	TraceBuffer     *bytes.Buffer
	ExecTime        time.Duration
}

// build from the files in sourceDir
func BuildCommand (sourceDir,destDir,exMode,mode string, race bool) string{
	//create binary file
	tmpBinary, err := ioutil.TempFile(destDir,"*"+exMode)
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
	return filepath.Base(tmpBinary.Name())
}


// - executes the instrumented binary
// - Parses collected trace
func ExecuteTrace(binary string, args ...string) (*ExecuteResult, error){
	var stderr bytes.Buffer
	var stdout bytes.Buffer
  var cmd *exec.Cmd
	// run
	log.Println("ExecuteTrace: Run ",binary)
	cmd = exec.Command(binary,args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	start := time.Now()
	if err := cmd.Run(); err != nil {
		end := time.Now()
		et := end.Sub(start)
		fmt.Printf("modified program failed\nErr: %v\nStderr: %v\nStdout: %v\n", err, stderr.String(),stdout.String())
		err1 := fmt.Errorf("%v:%v",err,stderr.String())
		ret := &ExecuteResult{&stderr,et}
		return ret, err1
	}
	end := time.Now()
	et := end.Sub(start)

	// check length of stderr
	if stderr.Len() == 0 {
		return nil, errors.New("empty trace")
	}

	ret := &ExecuteResult{&stderr,et}
	return ret,nil
}

// removes dir
func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Println("Cannot remove temp dir:", err)
	}
}
