package db

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"os/exec"
	"strings"
	"github.com/staheri/goatlib/trace"
	"strconv"
	"path/filepath"
)

type Row struct{
	id                int
	event             string
	g                 uint64
	rid               string
	pos               int
	stack_id          uint64
}


func StackToString(frames []*trace.Frame) string{
	s := ""
	for i:= len(frames)-1 ; i>=0 ; i--{
		s = s + fmt.Sprintf("%v\\n",ToString(frames[i]))
	}
	return s
}

func ToString(f *trace.Frame) string {
	fu := strings.Split(f.Fn,"/")
	return fmt.Sprintf("%s @ %s:%d ",fu[len(fu)-1],filepath.Base(f.File),f.Line)
}

// Visualize execution
func ExecVis(db *sql.DB, resultpath,dbName string, stacks map[uint64][]*trace.Frame, withStack bool) {

	// Variables
	var g,stack_id          uint64
	var id                  int
	var event               string
	var rid                 sql.NullString
	var pos                 sql.NullInt64

	dtab := make(map[int]*Row)
	allids   := []int{}

	// load table into dtab
	// identify ids to ignore
	q:= `SELECT * from
				(SELECT t1.id,t1.type,t1.g,t1.rid,t1.stack_id,t2.value
				FROM Events t1
				INNER JOIN (select * from global.catBLCK union select * from global.catSCHD) t3
				ON t1.type=t3.eventName
				LEFT JOIN (select * from Args where arg="pos") t2
				ON t1.id=t2.eventid
				ORDER BY t1.id DESC LIMIT `+strconv.Itoa(EXECBOUND)+`) mt
			ORDER BY mt.id;`
	res,err := db.Query(q)
	check(err)
	for res.Next(){
		err = res.Scan(&id,&event,&g,&rid,&stack_id,&pos)
		check(err)
		if _,ok := dtab[id] ; !ok{
			newrow := &Row{id:id}
			newrow.event = strings.Split(event,"Ev")[1]
			newrow.g = g
			if rid.Valid{
				newrow.rid = rid.String
			}else{
				newrow.rid = ""
			}
			if pos.Valid{
				newrow.pos = int(pos.Int64)
			}else{
				newrow.pos = -1
			}
			newrow.stack_id = stack_id
			dtab[id]=newrow
			allids = append(allids,id)
		}
	}
	res.Close()

	// obtain keepids
	/*
	curIgnore := 0
	keepids := []int{}
	for _,idx := range(allids){
		if curIgnore == len(ignoreids){
			keepids = append(keepids,idx)
			continue
		}
		if idx < ignoreids[curIgnore]{
			// include it
			keepids = append(keepids,idx)
		} else{
				curIgnore++
			}
	}
	*/

	// create header row{}
	ginfo := GetGoroutineInfo(db)
	targetG := []*GInfo{}
	targetG = append(targetG,&GInfo{id:0})
	targetG = append(targetG,ginfo.main)
	targetG = append(targetG,ginfo.app[1:]...) // exclude watcher goroutine

	gmatHeader := make([]string,len(targetG))
	keepgs := []uint64{}
	for i := 0 ; i<len(targetG) ; i++{
		gi := targetG[i]
		if gi.createStack_id != 0 {
			if i == 0 {
				panic("wrong index")
			}
			if i == 1{
				gmatHeader[i] = fmt.Sprintf("G%d\\nMAIN",gi.id)
			} else{
				if withStack{
					gmatHeader[i] = fmt.Sprintf("G%d\\n%v",gi.id,StackToString(stacks[gi.createStack_id]))
				} else{
					gmatHeader[i] = fmt.Sprintf("G%d",gi.id)
				}

			}
		} else{
			gmatHeader[i] = fmt.Sprintf("G%d\\nROOT",gi.id)
		}
		keepgs = append(keepgs,gi.id)
	}


	// create gmat
	var gmat [][]string

	for _,idx := range(allids){
		if !containsUInt64(keepgs,dtab[idx].g){
			//ignore it
			continue
		} else{
			gmatRow := make([]string,len(keepgs))
			for i,kg := range(keepgs){
				if dtab[idx].g == kg {
					if !strings.HasPrefix(dtab[idx].rid,"G") && dtab[idx].rid != ""{
						if dtab[idx].pos == 0{
							if withStack{
								gmatRow[i] = fmt.Sprintf("%v.(pre)%v\\n%v",dtab[idx].rid,dtab[idx].event,StackToString(stacks[dtab[idx].stack_id]))
							}else{
								gmatRow[i] = fmt.Sprintf("%v.(pre)%v",dtab[idx].rid,dtab[idx].event)
							}

						}else{
							if withStack{
								gmatRow[i] = fmt.Sprintf("%v.%v\\n%v",dtab[idx].rid,dtab[idx].event,StackToString(stacks[dtab[idx].stack_id]))
							}else{
								gmatRow[i] = fmt.Sprintf("%v.%v",dtab[idx].rid,dtab[idx].event)
							}
						}
					}else{
						if withStack{
							gmatRow[i] = fmt.Sprintf("%v\\n%v",dtab[idx].event,StackToString(stacks[dtab[idx].stack_id]))
						}else{
							gmatRow[i] = fmt.Sprintf("%v",dtab[idx].event)
						}
					}
				}else{
					gmatRow[i] = "-"
				}
			}
			gmat = append(gmat,gmatRow)
		}
	}
	outdot := resultpath + "/" + dbName
	outpdf := resultpath + "/" + dbName
	if withStack{
		outdot = outdot + "_fullvis.dot"
		outpdf = outpdf + "_fullvis.pdf"
	}else{
		outdot = outdot + "_vis.dot"
		outpdf = outpdf + "_vis.pdf"
	}


	//write dot file
	f, err := os.Create(outdot)
	if err != nil {
		panic(err)
	}


	f.WriteString(mat2dot(gmat,gmatHeader,withStack))
	f.Close()

	// start cmd
	// Create pdf
	_cmd := "dot -Tpdf " + outdot + " -o " + outpdf
	cmd := exec.Command("dot", "-Tpdf", outdot, "-o", outpdf)
	log.Printf(">>> ExecVis: Executing %s...\n", _cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("Error creating pdf: %s - %s",outpdf,fmt.Sprint(err) + ": " + stderr.String())
		//panic(err)
		return
	}
	//log.Println(">>> ExecVis: Result: " + out.String())
	// end cmd
	fmt.Println("ExecVis: Generated visualization: ", outpdf)


	/*_cmd = "open " + outpdf
	cmd = exec.Command("open", outpdf)
	log.Printf(">>> ExecVis: Opening %s...\n", _cmd)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		panic(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	log.Println(">>> ExecVis: Result: " + out.String())*/

}
