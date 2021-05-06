package traceops

import (
  "github.com/staheri/goatlib/trace"
)


// read trace and check if it deadlocks
func ReplayDeadlockChecker(tracePath, binaryPath string) DLReport{
	// obtain trace
	trc,err := ReadTrace(tracePath)
	check(err)

	parseRes,err := trace.ParseTrace(trc,binaryPath)
	check(err)

	return DeadlockChecker(parseRes,false)
}

// read trace and display its goroutines and stacks
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
