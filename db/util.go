package db

import (
	"strconv"
	"strings"
	_"github.com/jedib0t/go-pretty/table"
	"database/sql"
	"github.com/staheri/goatlib/trace"
	"log"
	"fmt"
)

type GInfo struct{
	id                uint64
	createStack_id    uint64
}

type GoroutineInfo struct{
	main          *GInfo
	trace        	*GInfo
	watcher       *GInfo
	app           []*GInfo
}

func (ginf *GoroutineInfo) String() string{
	s := fmt.Sprintf("Main: %v\n",ginf.main.id)
	s = s +  fmt.Sprintf("Trace: %v\n",ginf.trace.id)
	s = s +  fmt.Sprintf("Watcher: %v\n",ginf.watcher.id)
	for _,gi := range(ginf.app){
		s = s +  fmt.Sprintf("App: %v\n",gi.id)
	}
	return s
}

// blacklist events
func storeIgnore(e *trace.Event) bool{
	desc := EventDescriptions[e.Type]

	// we do not want to ignore GoSched, no matter what
	if desc.Name == "GoSched" || desc.Name == "GoCreate"{
		return false
	}

	for _,frm := range(e.Stk){
		if strings.HasPrefix(frm.Fn,"github.com/staheri/goat/goat."){
			return true
		}
	}
	return false
}


func GetGoroutineInfo(db *sql.DB) GoroutineInfo{
	var gs                GoroutineInfo
	var linkoff           sql.NullInt32
	var createStack_id    uint64
	var offsets           []int
	var g                 uint64

	findGStmt,err := db.Prepare("SELECT t1.g,t2.createStack_id FROM Events t1 inner join goroutines t2 on t1.g=t2.gid WHERE t1.offset=?")
	check(err)


	// Find G information
	q := `Select linkoff from Events where type="EvGoCreate";`
	res, err := db.Query(q)
	check(err)
	for res.Next(){
		err = res.Scan(&linkoff)
		if linkoff.Valid{
			offsets = append(offsets,int(linkoff.Int32))
		}
	}
	res.Close()

	for _,off := range(offsets){
		res2, err2 := findGStmt.Query(off)
		check(err2)
		if res2.Next(){
			err2 = res2.Scan(&g,&createStack_id)
			if gs.main == nil{
				gs.main = &GInfo{id:g,createStack_id : createStack_id}
				res2.Close()
				continue
			}
			if gs.trace == nil{
				gs.trace = &GInfo{id:g,createStack_id : createStack_id}
				res2.Close()
				continue
			}
			if gs.watcher == nil{
				gs.watcher = &GInfo{id:g,createStack_id : createStack_id}
				res2.Close()
				continue
			}
			appn := &GInfo{id:g,createStack_id : createStack_id}
			gs.app = append(gs.app,appn)
		}
		res2.Close()
	}

	findGStmt.Close()
	return gs
}

func DBExist(dbName string) (db *sql.DB,exist bool) {
	// Initial Connecting to mysql driver
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/")
	if err != nil {
		panic(err)
	}else{
		log.Println("Store: Initial connection established")
	}

	// Creating new database for current experiment
	res,err := db.Query("SHOW DATABASES LIKE '"+dbName + "';")
	check(err)
	if !res.Next(){
		// databases does not exist
		res.Close()
		return nil,false
	}
	res.Close()
	// Close conncection to re-establish it again with proper DBname
	err = db.Close()
	check(err)

	// Re-establish
	//dbName = "dinphilX18"
	db, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		panic(err)
	}else{
		log.Println("Store: Connected to ",dbName)
	}
	db.SetMaxOpenConns(50000)
	db.SetMaxIdleConns(40000)
	db.SetConnMaxLifetime(0)

  return db,true

}


// Operations on db
func Clean() {
	// Vars
	var dbs,q string
	// Connecting to mysql driver
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/")
	defer db.Close()
	check(err)

	log.Println("Clean: Clean all")
	res,err := db.Query("SHOW DATABASES;")
	check(err)
	for res.Next(){
		err := res.Scan(&dbs)
		check(err)
		//fmt.Printf("DB: %s \n",dbs)
		if dbs[len(dbs)-1] >= '0' && dbs[len(dbs)-1] <= '9'{
			q = "DROP DATABASE "+dbs+";"
			_,err2 := db.Exec(q)
			check(err2)
			log.Println("Clean: DROP ",dbs)
		}
	}
	err=res.Close()
	check(err)
}

func check(err error){
	if err != nil{
		panic(err)
	}
}

// If s contains e
func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func containsInt(l []int, b int) bool{
	for _, a := range l {
			if a == b {
					return true
			}
	}
	return false
}

func containsUInt64(l []uint64, b uint64) bool{
	for _, a := range l {
			if a == b {
					return true
			}
	}
	return false
}


func filterSlash(s string) string {
	ret := ""
	for _,b := range s{
		if string(b) == "/"{
			ret = ret + "\\"
		} else{
			ret = ret + string(b)
		}
	}
	return ret
}

func mat2dot(mat [][]string, header []string) string{

	width := "5"
	height := "2"
	fontsize := "11"

	if len(mat) < 1{
		panic("Mat is empty")
	}
	if len(mat[0]) < 1{
		panic("Mat row is empty")
	}

	tmp := ""
	dot := "digraph G{\n\trankdir=TB"

	//subgraph G labels (-1)
	tmp = "\n\tsubgraph{"
	tmp = tmp + "\n\t\tnode [margin=0 fontsize="+fontsize+" width="+width+" height="+height+" shape=box style=dashed fixedsize=true]"
	tmp = tmp + "\n\t\trank=same;"
	tmp = tmp + "\n\t\trankdir=LR"
	for i,g := range(header){
		tmp=tmp+"\n\t\t\"-1,"+strconv.Itoa(i)+"\" [label=\""+g+"\"]"
	}
	tmp = tmp + "\n\n\t\tedge [dir=none, style=invis]"

	for i:=0;i<len(mat[0])-1;i++{
		tmp = tmp + "\n\t\t\"-1,"+strconv.Itoa(i)+"\" -> \"-1,"+strconv.Itoa(i+1)+"\""
	}
	tmp = tmp+"\t}"
	dot = dot + tmp
	dot = dot + "\n"
	// For loop for all the subgraphs
	for i,row := range(mat){
		tmp = "\n\tsubgraph{"
		tmp = tmp + "\n\t\tnode [margin=0 fontsize="+fontsize+" width="+width+" shape=box style=invis]"
		tmp = tmp + "\n\t\trank=same;"
		tmp = tmp + "\n\t\trankdir=LR\n"
		for j,el := range(row){
			tmp = tmp + "\n\t\t\""+strconv.Itoa(i)+","+strconv.Itoa(j)+"\" "
			if el != "-"{
				if strings.Contains(el,"Block"){
					tmp = tmp + "[label=\""+el+"\",style=\"bold,filled\", fillcolor=firebrick1]"
				}else if strings.Contains(el,"(pre)"){
					tmp = tmp + "[label=\""+el+"\",style=\"dotted,filled\", fillcolor=gold]"
				}else if strings.Contains(el,"MuLock") || strings.Contains(el,"MuUnlock"){
					tmp = tmp + "[label=\""+el+"\",style=\"filled\", fillcolor=green2]"
				}else if strings.Contains(el,"Wg"){
					tmp = tmp + "[label=\""+el+"\",style=\"dashed,filled\", fillcolor=aqua]"
				}else if strings.Contains(el,"ChSend") || strings.Contains(el,"ChRecv"){
					tmp = tmp + "[label=\""+el+"\",style=\"filled\", fillcolor=green2]"
				}else{
					tmp = tmp + "[label=\""+el+"\",style=filled]"
				}
			}
		}

		tmp = tmp + "\n\n\t\tedge [dir=none, style=invis]"

		for j:=0;j<len(row)-1;j++{
			tmp = tmp + "\n\t\t\""+strconv.Itoa(i)+","+strconv.Itoa(j)+"\" -> \""+strconv.Itoa(i)+","+strconv.Itoa(j+1)+"\""
		}
		tmp = tmp+"\t}"
		dot = dot + tmp
		dot = dot + "\n"
	}


	//subgraph X
	tmp = "\n\tsubgraph{"
	tmp = tmp + "\n\t\tnode [margin=0 fontsize="+fontsize+" width="+width+" shape=box style=invis]"
	tmp = tmp + "\n\t\trank=same;"
	tmp = tmp + "\n\t\trankdir=LR"
	for i,_ := range(mat[0]){
		tmp=tmp+"\n\t\t\"x,"+strconv.Itoa(i)+"\""
	}
	tmp = tmp + "\n\n\t\tedge [dir=none, style=invis]"

	for i:=0;i<len(mat[0])-1;i++{
		tmp = tmp + "\n\t\t\"x,"+strconv.Itoa(i)+"\" -> \"x,"+strconv.Itoa(i+1)+"\""
	}
	tmp = tmp+"\t}"
	dot = dot + tmp
	dot = dot + "\n"
	// Edges
	dot = dot + "\n\tedge [dir=none, color=gray88]"
	for j := 0; j<len(mat[0]) ; j++{
		for i:= -1; i<len(mat) ; i++{
			if i == len(mat)-1{
				dot = dot + "\n\t\""+strconv.Itoa(i)+","+strconv.Itoa(j)+"\" -> \"x,"+strconv.Itoa(j)+"\""
			}else{
				dot = dot + "\n\t\""+strconv.Itoa(i)+","+strconv.Itoa(j)+"\" -> \""+strconv.Itoa(i+1)+","+strconv.Itoa(j)+"\""
			}
			dot = dot + "\n"
		}
	}
	dot = dot + "\n}"


	return dot
}
