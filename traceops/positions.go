package traceops
import (
	_"strings"
	_"github.com/jedib0t/go-pretty/table"
	"github.com/staheri/goatlib/trace"
	_"log"
	_"fmt"
  _"path/filepath"
	_"log"
)

var recvPos = map[int]string{
  0 : "blocked",
  1 : "roc", // receive on closed
  2 : "buf-dir", // buffered channel - directly from queue
  3 : "woken-up",
  4 : "sr-buf",
  5 : "sel-sr-rs?", // select, receive ready, case sent is selected
  6 : "sel-roc?", // select, receive ready, case sent is selected
}

var sendPos = map[int]string{
  0 : "blocked",
  1 : "none",
  2 : "woken-up",
  3 : "rr",
  4 : "sel-rr-ss?", // select, receive ready, case sent is selected
}

var selectPos = map[int]string{
  0 : "all",
  1 : "nb-send",
  2 : "nb-recv",
  3 : "nb-recv2",
}

var muPos = map[int]string{
  0 : "locked",
  1 : "free",
  2 : "woken-up",
}

var wgPos = map[int]string{
  0 : "blocked",
  1 : "free",
  2 : "woken-up",
}

var cvPos = map[int]string{
  0 : "none",
  1 : "sig",
  2 : "brdcst",
}

var (
  chEvents = []string{"ChMake","ChClose","ChSend","ChRecv"}
  muEvents = []string{"MuLock","MuUnlock"}
  cvEvents = []string{"CvWait","CvSig"}
  wgEvents = []string{"WgWait","WgAdd"}
  ssEvents = []string{"Select","Sched"}

  catGRTN  = []string{"GoCreate","GoStart","GoEnd","GoStop","GoSched","GoPreempt","GoSleep","GoBlock","GoUnblock","GoBlockSend","GoBlockRecv","GoBlockSelect","GoBlockSync","GoBlockCond","GoBlockNet","GoWaiting","GoInSyscall","GoStartLocal","GoUnblockLocal","GoSysExitLocal","GoStartLabel","GoBlockGC"}
	catBLCK  = []string{"GoCreate","GoStart","GoEnd","GoStop","GoSched","GoPreempt","GoSleep","GoBlock","GoUnblock","GoBlockSend","GoBlockRecv","GoBlockSelect","GoBlockSync","GoBlockCond","GoBlockNet","GoUnblockLocal","GoBlockGC"}
  catPROC  = []string{"None","Batch","Frequency","Stack","Gomaxprocs","ProcStart","ProcStop"}
  catGCMM  = []string{"GCStart","GCDone","GCSTWStart","GCSTWDone","GCSweepStart","GCSweepDone","HeapAlloc","NextGC","GCMarkAssistStart","GCMarkAssistDone"}
  catSYSC  = []string{"GoSysCall","GoSysExit","GoSysBlock"}
  catMISC  = []string{"UserTaskCreate","UserTaskEnd","UserRegion","UserLog","TimerGoroutine","FutileWakeup","String"}
  interestingEvents = [][]string{chEvents,muEvents,cvEvents,wgEvents,ssEvents,catBLCK}
  iEvents  = []string{}
)

// returns event position description
func GetPositionDesc(e *trace.Event) string{
	ed := trace.EventDescriptions[e.Type]
	if contains(chEvents,ed.Name){
		if ed.Name == "ChRecv" {
			return recvPos[int(e.Args[3])] // args[3] for channel.send/recv is pos
		}else if ed.Name == "ChSend"{
			return sendPos[int(e.Args[3])] // args[3] for channel.send/recv is pos
		}
	}else if contains(muEvents,ed.Name){
		if ed.Name == "MuLock" {
			return muPos[int(e.Args[1])]
		}
	}else if contains(cvEvents,ed.Name){
		if ed.Name == "CvSig" {
			return cvPos[int(e.Args[1])]
		}
	}else if contains(wgEvents,ed.Name){
		if ed.Name != "WgAdd" {
			return wgPos[int(e.Args[1])]
		}
	} else if contains(ssEvents,ed.Name){
		return selectPos[int(e.Args[0])]
	}
	return ""
}
