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
	"strings"
)

type ExecuteResult struct{
	TraceBuffer     *bytes.Buffer
	RaceMessage     string
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
  var stderr   bytes.Buffer
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
	var stderr         bytes.Buffer
	var stderr_back    bytes.Buffer
	var stdout         bytes.Buffer
	var traceBuf       bytes.Buffer
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
		//fmt.Printf("modified program failed\nErr: %v\nStderr: %v\nStdout: %v\n", err, stderr.String(),stdout.String())
		err1 := fmt.Errorf("%v:%v",err,stderr.String())

		raceSeparator := "=================="
		stderr_back = stderr
		fi := strings.Index(stderr_back.String(), raceSeparator)
		li := strings.LastIndex(stderr_back.String(), raceSeparator)
		fmt.Printf("fi: %v - li %v\n",fi,li)
		if fi != -1 && li > 0 && fi != li{ // there is a race
			raceOut := stderr_back.Bytes()[fi:li+len(raceSeparator)]
			raceMsg := bytes.NewBuffer(raceOut).String()
			//raceBuf = *bytes.NewBuffer(raceOut)
			traceOut := stderr.Bytes()[:fi]
			//fmt.Printf("TRACE _ 1 : %v\n",bytes.NewBuffer(traceOut).String())
			//fmt.Printf("RACE: %v\n",raceMsg)
			traceOut = append(traceOut,stderr.Bytes()[li+len(raceSeparator)+1:]...)
			//fmt.Printf("TRACE _ 1 : %v\n",bytes.NewBuffer(traceOut).String())
			traceBuf = *bytes.NewBuffer(traceOut)

			// if there is any runtime error: it would be in traceBuf

			ret := &ExecuteResult{TraceBuffer:&traceBuf,RaceMessage:raceMsg,ExecTime:et}
			// check length of stderr
			//if bytes.NewBuffer(traceOut).Len() == 0 {
			if traceBuf.Len() == 0 {
				return nil, errors.New("empty trace")
			}
			return ret,nil
		}
		ret := &ExecuteResult{TraceBuffer:&stderr,ExecTime:et}
		return ret, err1
	}
	end := time.Now()
	et := end.Sub(start)
	raceSeparator := "=================="
	stderr_back = stderr
	fi := strings.Index(stderr_back.String(), raceSeparator)
	li := strings.LastIndex(stderr_back.String(), raceSeparator)
	fmt.Printf("fi: %v - li %v\n",fi,li)
	if fi != -1 && li > 0 && fi != li{
		raceOut := stderr_back.Bytes()[fi:li+len(raceSeparator)]
		raceMsg := bytes.NewBuffer(raceOut).String()
		//raceBuf = *bytes.NewBuffer(raceOut)
		traceOut := stderr.Bytes()[:fi]
		//fmt.Printf("TRACE _ 1 : %v\n",bytes.NewBuffer(traceOut).String())
		//fmt.Printf("RACE: %v\n",raceMsg)
		traceOut = append(traceOut,stderr.Bytes()[li+len(raceSeparator)+1:]...)
		traceBuf = *bytes.NewBuffer(traceOut)

		//fmt.Printf("(W) TRACE _ 1 : %v\n",traceBuf.String())
		//fmt.Printf("(W) RACE: %v\n",raceBuf.String())
		ret := &ExecuteResult{TraceBuffer:&traceBuf,RaceMessage:raceMsg,ExecTime:et}
		// check length of stderr
		if bytes.NewBuffer(traceOut).Len() == 0 {
			return nil, errors.New("empty trace")
		}
		return ret,nil
	}
	ret := &ExecuteResult{TraceBuffer:&stderr,ExecTime:et}
	return ret,nil
}

// removes dir
func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Println("Cannot remove temp dir:", err)
	}
}
