package db

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"os"
	"strings"
	"github.com/jedib0t/go-pretty/table"
	"path/filepath"
)

var MessageFormat=`
=============================== GOAT ===============================
Total Leaked: %v
--------------------------------------------------------------------
%v
====================================================================
`

var MessagePerLeaked =`
Leaked #%v
--------------------------------------------------------------------
Goroutine ID %v
Created in %v
Last event: %v
StackTrace:
%v
`
type Report struct{
	GlobalDL      bool
	Leaked        int
	Message       string
	TotalG        int
}

// a function to query the database
// select * from events where type = gocreate
// for res.Next()
//      if ! mainGoroutine && res.linkoff != null:
//              maingoroutine = (select g from events where offset=res.linkoff)
//							continue
//      if ! traceGoutine && res.linkoff != null:
//             traceGoroutine = (select g from events where offset=res.linkoff)
//      if ! watchGoutine && res.linkoff != null:
//             watchGoroutine = (select g from events where offset=res.linkoff)

//
func Checker(db *sql.DB, long bool) Report{

	// Variables
	var isGlobalDL      bool
	var suspicious      []uint64
	var stack_id        uint64
	var createStack_id  uint64
	var event           string
	var file,funct,line string


	lastEventMap := make(map[uint64]string)
	stk_idLastEventMap := make(map[uint64]uint64)
	stackTraceMap := make(map[uint64]string)
	createLocMap := make(map[uint64]string)


	// Prepare statements
	lastEventStmt,err := db.Prepare("SELECT type,stack_id FROM Events WHERE g=? ORDER BY id DESC LIMIT 1")
	check(err)

	stackTraceStmt,err := db.Prepare("SELECT file,func,line FROM StackFrames WHERE stack_id=? ORDER BY id")
	check(err)

	createLocStmt,err := db.Prepare("SELECT createStack_id FROM Goroutines WHERE gid=?")
	check(err)

	// get goroutines information
	gs := GetGoroutineInfo(db)
	//fmt.Println(gs.String())


	// check for global deadlock
	res,err := lastEventStmt.Query(gs.main.id)
	check(err)
	isGlobalDL = false
	if res.Next(){
		err = res.Scan(&event,&stack_id)
		check(err)
		lastEventMap[gs.main.id]=event
		stk_idLastEventMap[gs.main.id]=stack_id
		if event != "EvGoSched"{
			isGlobalDL = true
		}
	}
	res.Close()

	if !isGlobalDL{
		q := `SELECT file,func,line FROM StackFrames WHERE stack_id=`+strconv.FormatUint(stack_id,10)+` ORDER BY id LIMIT 1`
		res,err = db.Query(q)
		check(err)
		if res.Next(){
			err = res.Scan(&file,&funct,&line)
			check(err)
			if !strings.HasPrefix(funct,"runtime.StopTrace"){
				isGlobalDL = true
			}
		}
		res.Close()
	}



	// check for partial deadlock
	totalg := len(gs.app)+1
	for _,gi := range(gs.app){
		// Last event
		res,err = lastEventStmt.Query(gi.id)
		check(err)
		for res.Next(){
			err = res.Scan(&event,&stack_id)
			check(err)
			if event != "EvGoEnd" {
				suspicious = append(suspicious,gi.id)
				lastEventMap[gi.id]=event
				stk_idLastEventMap[gi.id]=stack_id
			}
		}
		res.Close()
	}
	res.Close()
	lastEventStmt.Close()

	for _,gi := range(suspicious){
		res,err = stackTraceStmt.Query(stk_idLastEventMap[gi])
		check(err)
		stackElements := []string{}
		for res.Next(){
			err = res.Scan(&file,&funct,&line)
			check(err)
			stackElem := filepath.Base(funct)+"@"+filepath.Base(file)+":"+line
			stackElements = append(stackElements,stackElem)
		}
		res.Close()
		stackTraceMap[gi]=strings.Join(stackElements,"\n")

		res,err = createLocStmt.Query(gi)
		check(err)
		if res.Next(){
			err = res.Scan(&createStack_id)
			check(err)
		}
		res.Close()
		if createStack_id != 0{
			res,err = stackTraceStmt.Query(createStack_id)
			check(err)
			stackElements := []string{}
			for res.Next(){
				err = res.Scan(&file,&funct,&line)
				check(err)
				stackElem := []string{funct,file,line}
				stackElements = append(stackElements,strings.Join(stackElem,":"))
			}
			res.Close()
			createLocMap[gi]=strings.Join(stackElements,"\n")
		}
	}
	res.Close()
	stackTraceStmt.Close()
	createLocStmt.Close()

	// ****************
	// Generate Message

	msg := "=============================== GOAT ===============================\n"
	if isGlobalDL {
		msg = msg + fmt.Sprintf("Total leaked: %v\n",len(suspicious)+1)
	}else{
		msg = msg + fmt.Sprintf("Total leaked: %v\n",len(suspicious))
	}
	msg = msg + "--------------------------------------------------------------------"

	for i,g := range(suspicious){
		msg = msg + fmt.Sprintf(MessagePerLeaked,i,g,createLocMap[g],lastEventMap[g],stackTraceMap[g])
	}
	msg = msg + "===================================================================="


	if isGlobalDL{
		return Report{GlobalDL: true,Leaked:0, Message:msg, TotalG:totalg}
	} else if len(suspicious) != 0{
		//fmt.Println(string(colorRed),"Fail (partial deadlock)",string(colorReset))
		return Report{GlobalDL: false,Leaked:len(suspicious), Message:msg, TotalG:totalg}
	}
	//fmt.Println(string(colorGreen),"Pass",string(colorReset))
	return Report{GlobalDL: false,Leaked:0, Message:msg, TotalG:totalg}
}

func longLeakReport(db *sql.DB, gs GoroutineInfo) Report{

	var event,rid string

	// last events stroe every last event of goroutines
	lastEvents := make(map[uint64]string)


	lastEventStmt,err := db.Prepare("SELECT type FROM Events WHERE g=? ORDER BY id DESC LIMIT 1")
	check(err)
	defer lastEventStmt.Close()

	resStmt,err := db.Prepare("SELECT type,rid FROM Events WHERE rid IS NOT NULL AND g=?")
	check(err)
	defer resStmt.Close()


	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Goroutine","Last Event","Resources","Goroutines"})

	for _,gi := range(append(gs.app,gs.main)){
		// New row
		var row []interface{}
		row = append(row,gi.id)

		// Last event
		res,err := lastEventStmt.Query(gi.id)
		check(err)
		for res.Next(){
			err = res.Scan(&event)
			check(err)
			lastEvents[gi.id]=event
			//gs = append(gs,g) // append g to gs
			row = append(row,event)
		}

		// Resources
		resMap := make(map[string]int)
		var resources []interface{}
		var otherg []interface{}
		res,err = resStmt.Query(gi.id)
		check(err)
		for res.Next(){
			err = res.Scan(&event,&rid)
			check(err)
			if _,ok := resMap[rid]; !ok{
				resMap[rid] = 1
				if strings.HasPrefix(rid,"G"){
					otherg = append(otherg,rid)
				}else{
					resources = append(resources,rid)
				}

			}
		}
		row = append(row,resources)
		row = append(row,otherg)

		t.AppendRow(row)
		res.Close()
	}

	t.Render()
	//fmt.Println(lastEvents)
	return textReport(lastEvents,gs)
}

func textReport(lastEvents map[uint64]string,gs GoroutineInfo) Report{
	//writer := tabwriter.NewWriter(os.Stdout,0 , 16, 1, '\t', tabwriter.AlignRight)
	var  suspicious []uint64
	var   isGlobalDL   bool

	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorGreen := "\033[32m"

	if lastEvents[gs.main.id] != "EvGoSched"{
		isGlobalDL = true
	}
	for k,v := range(lastEvents){
		if v != "EvGoEnd" && k != gs.main.id{
			suspicious = append(suspicious,k)
		}
	}

	if isGlobalDL{
		fmt.Println("Global Deadlock:",string(colorRed),"TRUE",string(colorReset))
		return Report{GlobalDL: true,Leaked:0}
	} else{
		fmt.Println("Global Deadlock:",string(colorGreen),"FALSE",string(colorReset))
	}

	if len(suspicious) != 0{
		temp := ""
		for _,i := range(suspicious){
			temp = temp + strconv.FormatUint(i,10) + " "
		}
		fmt.Println("Leaked Goroutines:",string(colorRed),temp,string(colorReset))
		return Report{GlobalDL: false,Leaked:len(suspicious)}
	}
	fmt.Println("Leaked Goroutines:",string(colorGreen),"NONE",string(colorReset))
	return Report{GlobalDL: false,Leaked:0}
}
