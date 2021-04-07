package db

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	_"log"
	_"os"
	"strings"
	_"github.com/jedib0t/go-pretty/table"
	"github.com/staheri/goatlib/instrument"
	"path/filepath"
	"sort"
)

// Dynamic Concurrency Usage
func DynamicConcurrencyUsage(db *sql.DB, ct map[string][]bool,ctlist []string) (map[string][]bool) {

	// Variables
	var event,file,fun         string
	var eid,line,pos         int
	var covered                 bool
	//var _rid                    sql.NullString
	eventMap := make(map[int]string)
	posMap := make(map[int]int)


	// get goroutines information
	//gs := GetGoroutineInfo(db)

	q := "select t1.id,t1.type  from events t1 inner join global.catSCHD t2 on t1.type=t2.eventName;"
	res,err := db.Query(q)
	check(err)
	for res.Next(){
		pos = -1
		err = res.Scan(&eid,&event)
		check(err)
		eventMap[eid] = event
		/*if _rid.Valid{
			rid = _rid.String
		} else{
			rid = ""
		}*/
	}
	res.Close()

	for k,_ := range(eventMap){
		// find pos
		res2,err2 := db.Query("SELECT value FROM args WHERE eventid="+strconv.Itoa(k)+" and arg=\"pos\";")
		check(err2)
		if res2.Next(){
			err2 = res2.Scan(&pos)
			check(err2)
			posMap[k]=pos
		}
		res2.Close()
	}
	for k,_ := range(eventMap){
		// find stackframes
		covered = false
		res3,err3 := db.Query("SELECT file,func,line FROM stackframes WHERE eventid="+strconv.Itoa(k)+";")
		check(err3)
		s := ""
		for res3.Next(){
			err3 = res3.Scan(&file,&fun,&line)
			s = fmt.Sprintf("%s/%s/%s:%d",filepath.Dir(fun),strings.Split(filepath.Base(fun),".")[0],file,line)
			//fmt.Println(event," >> " , s)
			if contains(ctlist,s){
				//fmt.Println("HITTTT")
				covered = true
				ct[s][0]=true
				res3.Close()
				break
			}
		}
		res3.Close()
		if !covered {
			continue
		}
		if posMap[k] == 0{
			ct[s][2] = true
		} else if posMap[k] == 1{
			ct[s][1] = true
		}
		//eventMap[g] = append(eventMap[g],rid+":"+event+"<>"+strconv.Itoa(pos)+"<>"+s)
	}
	//res.Close()
	//res2.Close()
	//res3.Close()
	//stackStmt.Close()
	//argStmt.Close()
	return ct
}


func InitiCoverageTable(concusage []*instrument.ConcurrencyUsage) (ct map[string][]bool, ctlist []string){
	ct = make(map[string][]bool)
	for _,cu := range(concusage){
		ct["_"+cu.String()]=[]bool{false,false,false}
		ctlist = append(ctlist,"_"+cu.String())
	}
	fmt.Println(ct)
	fmt.Println(ctlist)
	sort.Strings(ctlist)
	return ct,ctlist
}
