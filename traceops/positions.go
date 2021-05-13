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
  1 : "sr", // send ready
  2 : "woken",
  3 : "buf", // buffered channel - directly from queue
  4 : "roc", // receive on close
}

var sendPos = map[int]string{
  0 : "blocked",
  1 : "rr",
  2 : "woken",
  3 : "buf",
}

var selectPos = map[int]string{
  0 : "blocked",
  1 : "avail",
  2 : "woken",
  3 : "nb",
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

  CatGRTN  = []string{"GoCreate","GoStart","GoEnd","GoStop","GoSched","GoPreempt","GoSleep","GoBlock","GoUnblock","GoBlockSend","GoBlockRecv","GoBlockSelect","GoBlockSync","GoBlockCond","GoBlockNet","GoWaiting","GoInSyscall","GoStartLocal","GoUnblockLocal","GoSysExitLocal","GoStartLabel","GoBlockGC"}
	CatBLCK  = []string{"GoCreate","GoStart","GoEnd","GoStop","GoSched","GoPreempt","GoSleep","GoBlock","GoUnblock","GoBlockSend","GoBlockRecv","GoBlockSelect","GoBlockSync","GoBlockCond","GoBlockNet","GoUnblockLocal","GoBlockGC"}
  CatPROC  = []string{"None","Batch","Frequency","Stack","Gomaxprocs","ProcStart","ProcStop"}
  CatGCMM  = []string{"GCStart","GCDone","GCSTWStart","GCSTWDone","GCSweepStart","GCSweepDone","HeapAlloc","NextGC","GCMarkAssistStart","GCMarkAssistDone"}
  CatSYSC  = []string{"GoSysCall","GoSysExit","GoSysBlock"}
  CatMISC  = []string{"UserTaskCreate","UserTaskEnd","UserRegion","UserLog","TimerGoroutine","FutileWakeup","String"}
  interestingEvents = [][]string{chEvents,muEvents,cvEvents,wgEvents,ssEvents,CatBLCK}
  iEvents  = []string{}
)

// returns event position description
func GetPositionDesc(e *trace.Event) string{
	ed := trace.EventDescriptions[e.Type]
	if contains(chEvents,ed.Name){
		if ed.Name == "ChRecv" {
			return recvPos[int(e.Args[1])] // args[3] for channel.send/recv is pos
		}else if ed.Name == "ChSend"{
			return sendPos[int(e.Args[1])] // args[3] for channel.send/recv is pos
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
