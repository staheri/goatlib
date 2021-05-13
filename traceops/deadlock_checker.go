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

	gs,gmap,_ := GetGoroutineInfo(parseResult)
	//gs := GetGoroutineInfo(parseResult)
	//fmt.Println(gs.StringDetail())
	GoroutineTable(gmap)

	leakedMsg := ""

	// ToStringIsViz
	isViz := false

	// check for global deadlock
	// check the last event of Main
	if len(gs.Main.Events) == 0{
		rep.GlobalDL = true
		leaked = append(leaked,gs.Main.Gid)
		leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gs.Main.Gid,"ROOT","NULL","NULL")
	} else if trace.EventDescriptions[gs.Main.Events[len(gs.Main.Events)-1].Type].Name != "GoSched"{
		rep.GlobalDL = true
		leaked = append(leaked,gs.Main.Gid)
		leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gs.Main.Gid,"ROOT",trace.EventDescriptions[gs.Main.Events[len(gs.Main.Events)-1].Type].Name,stackToString(gs.Main.Events[len(gs.Main.Events)-1].Stk,isViz))
	}
	if !rep.GlobalDL{
		if !strings.HasPrefix(gs.Main.Events[len(gs.Main.Events)-1].Stk[0].Fn,"runtime.StopTrace"){
			rep.GlobalDL = true
			leaked = append(leaked,gs.Main.Gid)
			leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gs.Main.Gid,"ROOT",trace.EventDescriptions[gs.Main.Events[len(gs.Main.Events)-1].Type].Name,stackToString(gs.Main.Events[len(gs.Main.Events)-1].Stk,isViz))
		}
	}

	for _,gi := range(gs.App){
		//fmt.Println("TEST")
		//fmt.Println(gi.lastEvent.String())
		//fmt.Println("APP G LastEVENT Type",gi.lastEvent.Type)
		if len(gi.Events) == 0{
			leaked = append(leaked,gi.Gid)
			leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gi.Gid,stackToString(parseResult.Stacks[gi.CreateStack_id],isViz),"NULL","NULL")
		}else if trace.EventDescriptions[gi.Events[len(gi.Events)-1].Type].Name != "GoEnd"{
			leaked = append(leaked,gi.Gid)
			leakedMsg = leakedMsg + fmt.Sprintf(MessagePerLeaked,gi.Gid,stackToString(parseResult.Stacks[gi.CreateStack_id],isViz),trace.EventDescriptions[gi.Events[len(gi.Events)-1].Type].Name,stackToString(gi.Events[len(gi.Events)-1].Stk,isViz))
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
