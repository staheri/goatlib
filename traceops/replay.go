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



func ReplayDispGMAP(tracePath, binaryPath string) {
	// obtain trace
	trc,err := ReadTrace(tracePath)
	check(err)

	parseRes,err := trace.ParseTrace(trc,binaryPath)
	check(err)
	_,gmap := GetGoroutineInfo(parseRes)
	GoroutineTable(gmap)
	StackTable(parseRes.Stacks)
}
