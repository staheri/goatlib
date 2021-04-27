package traceops

import (
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_"strconv"
  "github.com/staheri/goatlib/trace"
	_"path/filepath"

)


var MessagePerLeaked =`
--------------------------------------------------------------------
[*] Leaked Goroutine: %v
[*] Created at:
%v
[*] Last event: %v
[*] StackTrace:
%v
`

type DLReport struct{
	GlobalDL      bool
	Leaked        int
	Message       string
	TotalG        int
}


func DeadlockChecker(parseResult *trace.ParseResult, long bool) DLReport{

	// Variables
	rep := DLReport{}
	var leaked []uint64

	gs,gmap := GetGoroutineInfo(parseResult)
	//gs := GetGoroutineInfo(parseResult)
	//fmt.Println(gs.StringDetail())
	//GoroutineTable(gmap)

	leakedMsg := ""

	// check for global deadlock
	// check the last event of main
	if gs.main.lastEvent == nil{
		rep.GlobalDL = true
		leaked = append(leaked,gs.main.gid)
		leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gs.main.gid,"ROOT","NULL","NULL")
	} else if trace.EventDescriptions[gs.main.lastEvent.Type].Name != "GoSched"{
		rep.GlobalDL = true
		leaked = append(leaked,gs.main.gid)
		leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gs.main.gid,"ROOT",trace.EventDescriptions[gs.main.lastEvent.Type].Name,stackToString(gs.main.lastEvent.Stk))
	}
	if !rep.GlobalDL{
		if !strings.HasPrefix(gs.main.lastEvent.Stk[0].Fn,"runtime.StopTrace"){
			rep.GlobalDL = true
			leaked = append(leaked,gs.main.gid)
			leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gs.main.gid,"ROOT",trace.EventDescriptions[gs.main.lastEvent.Type].Name,stackToString(gs.main.lastEvent.Stk))
		}
	}

	for _,gi := range(gs.app){
		//fmt.Println("TEST")
		//fmt.Println(gi.lastEvent.String())
		//fmt.Println("APP G LastEVENT Type",gi.lastEvent.Type)
		if gi.lastEvent == nil{
			leaked = append(leaked,gi.gid)
			leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gi.gid,stackToString(parseResult.Stacks[gi.createStack_id]),"NULL","NULL")
		}else if trace.EventDescriptions[gi.lastEvent.Type].Name != "GoEnd"{
			leaked = append(leaked,gi.gid)
			leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gi.gid,stackToString(parseResult.Stacks[gi.createStack_id]),trace.EventDescriptions[gi.lastEvent.Type].Name,stackToString(gi.lastEvent.Stk))
		}
	}

	// Message generation
	msg := "=============================== GOAT ===============================\n"
	if rep.GlobalDL {
		msg = msg + fmt.Sprintf("Total leaked: %v (Global Deadlock, leaked: main + %d app goroutines)\n",len(leaked),len(leaked)-1)
	}else{
		msg = msg + fmt.Sprintf("Total leaked: %v (Partial Deadlock)\n",len(leaked),len(leaked)-1)
	}
	msg = msg + leakedMsg

	rep.Leaked = len(leaked)
	rep.Message = msg
	rep.TotalG = len(gmap)

  return rep
}
