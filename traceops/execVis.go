package traceops

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"github.com/staheri/goatlib/trace"
	"strconv"
	//"path/filepath"
)

type Row struct{
	id                int
	event             string
	g                 uint64
	rid               string
	pos               int
	stack_id          uint64
}




// Visualize execution
func ExecVis(tracePath, binaryPath,resultPathName string, withStack bool) {
  // Gather interesting events
  for _,evs := range(interestingEvents){
    iEvents = append(iEvents,evs...)
  }
  // obtain trace
  trc,err := ReadTrace(tracePath)
  check(err)

  parseRes,err := trace.ParseTrace(trc,binaryPath)
  check(err)
  ginfo,gmap,_ := GetGoroutineInfo(parseRes)
  GoroutineTable(gmap)
  //StackTable(parseRes.Stacks)

  targetG := []*GInfo{}
	targetG = append(targetG,gmap[0])
	targetG = append(targetG,ginfo.Main)
	targetG = append(targetG,ginfo.App...) // ignoe goat_watch
  targetGuint := []uint64{}

  gmatHeader := make([]string,len(targetG))
  for i := 0 ; i<len(targetG) ; i++{
    gi := targetG[i]
    targetGuint = append(targetGuint,gi.Gid)
    if gi.CreateStack_id != 0 {
      if i == 0 {
        panic("wrong index")
      }
      if i == 1{
        gmatHeader[i] = fmt.Sprintf("G%d\\nMAIN",gi.Gid)
      } else{
        if withStack{
          gmatHeader[i] = fmt.Sprintf("G%d\\n%v",gi.Gid,stackToString(parseRes.Stacks[gi.CreateStack_id],true)) // isViz: true
        } else{
          gmatHeader[i] = fmt.Sprintf("G%d",gi.Gid)
        }
      }
    } else{
      gmatHeader[i] = fmt.Sprintf("G%d\\nROOT",gi.Gid)
    }
  }

  // for _,h := range(gmatHeader){
  //   fmt.Println(h)
  // }


  // create gmat
  var gmat [][]string


  for _,e := range(parseRes.Events){
    //fmt.Println(e.String())
    // init
    var rid,_pos        string
    var wgVal           int
    pos := -1           // default for pos
    ed := trace.EventDescriptions[e.Type]

    // ignore uninteresting Gs
    if !containsUInt64(targetGuint,e.G){
      continue
    }

    // ignore uninteresting events
    if !contains(iEvents,ed.Name){
      continue
    }

    // ignore GOAT events
    if isGoatFunction(e.Stk) && ed.Name != "GoSched"{
      continue
    }

    if contains(chEvents,ed.Name){
      rid = fmt.Sprintf("Ch(%v)",e.Args[0]) // args[0] for channel is cid
      if ed.Name == "ChRecv" {
        pos = int(e.Args[3])
        _pos = recvPos[pos] // args[3] for channel.send/recv is pos
        rid = fmt.Sprintf("Ch(%v)[%v]",e.Args[0],_pos) // include pos in rid
      }else if ed.Name == "ChSend"{
        pos = int(e.Args[3])
        _pos = sendPos[pos] // args[3] for channel.send/recv is pos
        rid = fmt.Sprintf("Ch(%v)[%v]",e.Args[0],_pos) // include pos in rid
      }
    }else if contains(muEvents,ed.Name){
      rid = fmt.Sprintf("Mu(%v)",e.Args[0]) // args[0] for mutex is mu id
      if ed.Name == "MuLock" {
        pos = int(e.Args[1]) // args[1] for mutex.lock is pos
        _pos = muPos[pos]
        rid = fmt.Sprintf("Mu(%v)[%v]",e.Args[0],_pos) // args[0] for mutex is mu id
      }
    }else if contains(cvEvents,ed.Name){
      rid = fmt.Sprintf("Cv(%v)",e.Args[0]) // args[0] for cond var is cv id
      if ed.Name == "CvSig" {
        pos = int(e.Args[1]) // args[1] for cond var is pos
        _pos = cvPos[pos]
        rid = fmt.Sprintf("Cv(%v)[%v]",e.Args[0],_pos) // args[0] for cond var is cv id
      }
    }else if contains(wgEvents,ed.Name){
      rid = fmt.Sprintf("Wg(%v)",e.Args[0]) // args[0] for WaitGroup is wg id
      if ed.Name == "WgAdd" {
        wgVal = int(e.Args[1]) // args[1] for wgAdd is val
        rid = fmt.Sprintf("Wg(%v)[val:%v]",e.Args[0],wgVal) // args[0] for WaitGroup is wg id
      }else{
        pos = int(e.Args[1]) // args[1] for wgWait is pos
        _pos = wgPos[pos]
        rid = fmt.Sprintf("Wg(%v)[%v]",e.Args[0],_pos) // args[0] for cond var is cv id
      }
    } else if contains(ssEvents,ed.Name){
      pos = int(e.Args[0]) // args[0] for ss is pos
      _pos = selectPos[pos]
      rid = fmt.Sprintf("SS(%v)[%v]",e.Args[0],_pos) // args[0] for SS is pos
    }

    // ignore based on E.stack (ignore goat events)
    gmatRow := make([]string,len(targetGuint))
    for i,kg := range(targetGuint){
      if e.G == kg { // check if this is the goroutine that we want
        if rid != ""{
          if pos == 0{
            if withStack{
              gmatRow[i] = fmt.Sprintf("%v.(pre)%v\\n%v",rid,ed.Name,stackToString(e.Stk,true))
            }else{
              gmatRow[i] = fmt.Sprintf("%v.(pre)%v",rid,ed.Name)
            }
          }else{
            if withStack{
              gmatRow[i] = fmt.Sprintf("%v.%v\\n%v",rid,ed.Name,stackToString(e.Stk,true))
            }else{
              gmatRow[i] = fmt.Sprintf("%v.%v",rid,ed.Name)
            }
          }
        }else{
          if withStack{
            gmatRow[i] = fmt.Sprintf("%v\\n%v",ed.Name,stackToString(e.Stk,true))
          }else{
            gmatRow[i] = fmt.Sprintf("%v",ed.Name)
          }
        }
      }else{
        gmatRow[i] = "-"
      }
    }
    gmat = append(gmat,gmatRow)
  }

  // for _,row := range(gmat){
  //   fmt.Println(row)
  // }

	outdot := resultPathName
	outpdf := resultPathName
	if withStack{
		outdot = outdot + "_fullVis.dot"
		outpdf = outpdf + "_fullVis.pdf"
	}else{
		outdot = outdot + "_minVis.dot"
		outpdf = outpdf + "_minVis.pdf"
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


}

func isGoatFunction(stack []*trace.Frame) bool{
  for _,frm := range(stack){
    if strings.HasPrefix(frm.Fn,"github.com/staheri/goat/goat"){
      return true
    }
  }
  return false
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
}
