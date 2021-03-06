package traceops
import (
	_"strings"
	"github.com/jedib0t/go-pretty/table"
	"github.com/staheri/goatlib/trace"
	_"log"
	_"fmt"
  _"path/filepath"
  "os"
  "sort"
  "strings"
  "strconv"
)

// print stack table
func StackTable (stackFrames map[uint64][]*trace.Frame) {
  var allIDs []uint64

  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"ID","File","function","Line"})

  for k,_:=range(stackFrames){
    allIDs = append(allIDs,k)
  }
  sort.Slice(allIDs, func(i,j int) bool {
    return allIDs[i] < allIDs[j]
  })
  for _,id := range(allIDs){
    var row []interface{}
    files := []string{}
    funcs := []string{}
    lines := []string{}
    row = append(row,id)
    for _,frm := range(stackFrames[id]){
      files = append(files,frm.File)
      funcs = append(funcs,frm.Fn)
      lines = append(lines,strconv.Itoa(frm.Line))
    }
    row = append(row,strings.Join(files,"\n"))
    row = append(row,strings.Join(funcs,"\n"))
    row = append(row,strings.Join(lines,"\n"))
    t.AppendRow(row)
    t.AppendSeparator()
  }
  t.Render()
}

// print goroutine table
func GoroutineTable (gmap map[uint64]*GInfo) {
  var allGs []uint64

  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"G","Parent","CreateStack ID","Type","Last Event","Ended"})

  for k,_:=range(gmap){
    allGs = append(allGs,k)
  }
  sort.Slice(allGs, func(i,j int) bool {
    return allGs[i] < allGs[j]
  })
  for _,gid := range(allGs){
    var row []interface{}
    row = append(row,gid)
    row = append(row,gmap[gid].Parent_id)
    row = append(row,gmap[gid].CreateStack_id)
    row = append(row,gtypes[gmap[gid].Gtype])
    if len(gmap[gid].Events) != 0{
      row = append(row,trace.EventDescriptions[gmap[gid].Events[len(gmap[gid].Events)-1].Type].Name)
    }else{
      row = append(row,"NULL")
    }

    row = append(row,gmap[gid].Ended)
    t.AppendRow(row)
  }
  t.Render()
}
