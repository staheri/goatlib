package traceops

//import ()


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


/*func filterSlash(s string) string {
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

func mat2dot(mat [][]string, header []string, withStack bool) string{

	width := "0"
	height := "0"
	fontsize := "0"

	if withStack{
		width = "5"
		height = "2"
		fontsize = "11"
	} else{
		width = "0.6"
		height = "0.3"
		fontsize = "6"
	}


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
		if withStack{
			tmp = tmp + "\n\t\tnode [margin=0 fontsize="+fontsize+" width="+width+" shape=box style=invis]"
		}else{
			tmp = tmp + "\n\t\tnode [margin=0 fontsize="+fontsize+" width="+width+" shape=circle style=invis]"
		}

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
}*/
