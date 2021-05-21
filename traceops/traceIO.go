package traceops

import (
	"github.com/staheri/goatlib/trace"
	"os"
	"io/ioutil"
	"encoding/json"
	"io"
	"bytes"
)

func WriteTrace(traceBuff []byte, fileName string) int {
	out,err := os.Create(fileName)
	check(err)
	size,err := out.Write(traceBuff)
	check(err)
	out.Close()
	return size
}

func ReadTrace(fileName string) (io.Reader,int, error){
	buf,err := ioutil.ReadFile(fileName)
	return bytes.NewReader(buf),len(buf),err
}

func TraceToJSON(parseResult *trace.ParseResult, jsonPath string){
	// create one file with two keys: events, stacks
	rep,err := os.Create(jsonPath)
	check(err)
	newdat ,err := json.MarshalIndent(parseResult,"","    ")
	check(err)
	_,err = rep.WriteString(string(newdat))
	check(err)
	rep.Close()
}


func WriteTime(time string, fileName string) {
	err := ioutil.WriteFile(fileName, []byte(time), 0777)
	check(err)
}


func ReadTime(fileName string) string{
	data, err := ioutil.ReadFile(fileName)
  check(err)
  return string(data)
}


// read trace and parse
func ReadParseTrace(tracePath, binaryPath string) *trace.ParseResult{
	// obtain trace
	trc,_,err := ReadTrace(tracePath)
	check(err)

	parseRes,err := trace.ParseTrace(trc,binaryPath)
	check(err)

	return parseRes
}
