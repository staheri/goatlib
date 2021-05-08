package traceops

import (
	_"fmt"
	"github.com/staheri/goatlib/trace"
	_"path"
	_"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_"strconv"
	"os"
	_"strings"
	_"time"
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

func ReadTrace(fileName string) (io.Reader, error){
	buf,err := ioutil.ReadFile(fileName)
	return bytes.NewReader(buf),err
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
