package instrument

import (
	"bytes"
	"errors"
	"fmt"
  "io"
	"io/ioutil"
	"os"
	"os/exec"
	"log"
	"time"
  "github.com/staheri/goatlib/trace"
)

// - Compile and executes the modified source
// - Parse collected trace
func (app *App)ExecuteTrace() ([]*trace.Event, error){
  var cmd *exec.Cmd
	// create binary file holder
	log.Println("ExecuteTrace: Create tempdir ")
	tmpBinary, err := ioutil.TempFile("", "GOAT")
  if err != nil {
		fmt.Println("create temp file error")
		return nil, err
	}

	// remove it after done
	defer os.Remove(tmpBinary.Name())

	// build binary
  if app.IsTest{
    cmd = exec.Command("go", "test", "-c", "-o", tmpBinary.Name())
  } else{
    cmd = exec.Command("go", "build", "-o", tmpBinary.Name())
  }

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Dir = app.Path

	// timing start
	start := time.Now()

	err = cmd.Run()
	if err != nil {
		fmt.Println("go build error", stderr.String())
		return nil, err
	}

	// timing end
	end := time.Since(start)
	log.Printf("[TIME %v: %v]\n","ExecTrace Build",end)
	if TIMING{
		fmt.Printf("[TIME %v: %v]\n","ExecTrace Build",end)
	}


	// run
	log.Println("ExecuteTrace: Run ",tmpBinary.Name())
	stderr.Reset()
	cmd = exec.Command(tmpBinary.Name())
	cmd.Stderr = &stderr

	// timing start
	start = time.Now()

	if err = cmd.Run(); err != nil {
		fmt.Println("modified program failed:", err, stderr.String())
		return nil, err
	}
	// timing end
	end = time.Since(start)
	log.Printf("[TIME %v: %v]\n","ExecTrace Run",end)
	if TIMING{
		fmt.Printf("[TIME %v: %v]\n","ExecTrace Run",end)
	}


	// check length of stderr
	if stderr.Len() == 0 {
		return nil, errors.New("empty trace")
	}

	// parse
	log.Println("ExecuteTrace: Redirect stderr to ParseTrace ")

	// timing start
	start = time.Now()
	// command
	r,e := parseTrace(&stderr, tmpBinary.Name())
	// timing end
	end = time.Since(start)
	log.Printf("[TIME %v: %v]\n","Parse Trace",end)
	if TIMING{
		fmt.Printf("[TIME %v: %v]\n","Parse Trace",end)
	}
	return r,e
}

// removes dir
func removeDir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Println("Cannot remove temp dir:", err)
	}
}

// reads trace from stderr (io.reader) and parse
func parseTrace(r io.Reader, binary string) ([]*trace.Event, error) {
	parseResult, err := trace.Parse(r,binary)
	if err != nil {
		return nil, err
	}

	err = trace.Symbolize(parseResult.Events, binary)

	return parseResult.Events, err
}
