package db

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"log"
	"os"
	"strings"
	"github.com/jedib0t/go-pretty/table"

)

type GoroutineInfo struct{
	main          int
	trace         int
	watcher       int
	app           []int
}

func gids(dbName string) (gs GoroutineInfo){
	// Variables
	var   linkoff  sql.NullInt32
	var   g        int

	// Establish connection to DB
 db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/"+dbName)
 check(err)
 log.Println("Cheker(short): Connected to ",dbName)
 defer db.Close()
 // END DB

 q := `Select linkoff from Events where type="EvGoCreate";`
 res, err := db.Query(q)
 check(err)
 for res.Next(){
	 err = res.Scan(&linkoff)
	 if linkoff.Valid {
		 //
		 q2 := `Select g from Events where offset=`+strconv.Itoa(int(linkoff.Int32))+`;`
		 res2, err2 := db.Query(q2)
		 check(err2)
		 if res2.Next(){
			 err2 = res2.Scan(&g)
			 if gs.main == 0{
				 gs.main = g
				 continue
			 }
			 if gs.trace == 0{
				 gs.trace = g
				 // check
				 continue
			 }
			 if gs.watcher == 0{
				 gs.watcher = g
				 // check
				 continue
			 }
			 gs.app = append(gs.app,g)
		 }
		 res2.Close()
	 }
 }
 res.Close()
 return gs
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
func Checker(dbName string, long bool) bool{
	gs := gids(dbName)
	// fmt.Println("Main:",gs.main)
	// fmt.Println("Watcher:",gs.watcher)
	// fmt.Println("Trace:",gs.trace)
	// fmt.Println("App:",gs.app)
  if long {
		return longLeakReport(dbName,gs)
	}
	return shortLeakReport(dbName,gs)
}

func shortLeakReport(dbName string,gs GoroutineInfo) bool{
	// Variables
	var   isGlobalDL   bool
	var suspicious []int
	var event string

	// Establish connection to DB
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		panic(err)
	}else{
		log.Println("Cheker(short): Connected to ",dbName)
	}
	defer db.Close()
	// END DB

	lastEventStmt,err := db.Prepare("SELECT type FROM Events WHERE g=? ORDER BY id DESC LIMIT 1")
	check(err)
	defer lastEventStmt.Close()

	// check for global deadlock
	res,err := lastEventStmt.Query(gs.main)
	check(err)
	if res.Next(){
		err = res.Scan(&event)
		check(err)
		if event != "EvGoSched"{
			isGlobalDL = true
		}
	}
	res.Close()

	// check for partial deadlock
	for _,gi := range(gs.app){
		// Last event
		res,err = lastEventStmt.Query(gi)
		check(err)
		for res.Next(){
			err = res.Scan(&event)
			check(err)
			if event != "EvGoEnd" {
				suspicious = append(suspicious,gi)
			}
		}
		res.Close()
	}


	// ****************
	// Generate report
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	if isGlobalDL{
		fmt.Println(string(colorRed),"Fail (global deadlock)",string(colorReset))
		return false
	} else if len(suspicious) != 0{
		fmt.Println(string(colorRed),"Fail (partial deadlock - leak)",string(colorReset))
		return false
	}
	fmt.Println(string(colorGreen),"Pass",string(colorReset))
	return true
}

func longLeakReport(dbName string, gs GoroutineInfo) bool{
	// Establish connection to DB
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		panic(err)
	}else{
		log.Println("Cheker(long): Connected to ",dbName)
	}
	defer db.Close()
	// END DB

	var event,rid string

	// last events stroe every last event of goroutines
	lastEvents := make(map[int]string)


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
		row = append(row,gi)

		// Last event
		res,err := lastEventStmt.Query(gi)
		check(err)
		for res.Next(){
			err = res.Scan(&event)
			check(err)
			lastEvents[gi]=event
			//gs = append(gs,g) // append g to gs
			row = append(row,event)
		}

		// Resources
		resMap := make(map[string]int)
		var resources []interface{}
		var otherg []interface{}
		res,err = resStmt.Query(gi)
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

func textReport(lastEvents map[int]string,gs GoroutineInfo) bool{
	//writer := tabwriter.NewWriter(os.Stdout,0 , 16, 1, '\t', tabwriter.AlignRight)
	var  suspicious []int
	var   isGlobalDL   bool

	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorGreen := "\033[32m"

	if lastEvents[gs.main] != "EvGoSched"{
		isGlobalDL = true
	}
	for k,v := range(lastEvents){
		if v != "EvGoEnd"{
			suspicious = append(suspicious,k)
		}
	}

	if isGlobalDL{
		fmt.Println("Global Deadlock:",string(colorRed),"TRUE",string(colorReset))
		return false
	} else{
		fmt.Println("Global Deadlock:",string(colorGreen),"FALSE",string(colorReset))
	}

	if len(suspicious) != 0{
		temp := ""
		for _,i := range(suspicious){
			temp = temp + strconv.Itoa(i) + " "
		}
		fmt.Println("Leaked Goroutines:",string(colorRed),temp,string(colorReset))
		return false
	}
	fmt.Println("Leaked Goroutines:",string(colorGreen),"NONE",string(colorReset))
	return true
}
