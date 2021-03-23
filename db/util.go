package db

import (
	"strconv"
	"strings"
	_"github.com/jedib0t/go-pretty/table"
	"database/sql"
	"log"
)



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

	width := "2"
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
	tmp = tmp + "\n\t\tnode [margin=0 fontsize="+fontsize+" width="+width+" shape=box style=dashed]"
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
				if strings.Contains(el,"Mu") || strings.Contains(el,"RWM"){
					tmp = tmp + "[label=\""+el+"\",style=\"dotted,filled\", fillcolor=green]"
				}else if strings.Contains(el,"Wg"){
					tmp = tmp + "[label=\""+el+"\",style=\"dashed,filled\", fillcolor=gold]"
				}else if strings.Contains(el,"ChSend"){
					tmp = tmp + "[label=\""+el+"\",style=\"bold,filled\", fillcolor=cyan]"
				}else if strings.Contains(el,"ChRecv"){
					tmp = tmp + "[label=\""+el+"\",style=\"bold,filled\", fillcolor=violet]"
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
