package traceops

import (
	_ "github.com/go-sql-driver/mysql"
	_"strconv"
  "github.com/staheri/goatlib/trace"
	_"path/filepath"
)


func ReplayDeadlockChecker(tracePath, binaryPath string) DLReport{
	// obtain trace
	trc,err := ReadTrace(tracePath)
	check(err)

	parseRes,err := trace.ParseTrace(trc,binaryPath)
	check(err)

	return DeadlockChecker(parseRes,false)
}
