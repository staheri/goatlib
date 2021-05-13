package traceops

import (
	"fmt"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	_"strconv"
  "github.com/staheri/goatlib/trace"
	_"path/filepath"
	"github.com/jedib0t/go-pretty/table"
	"github.com/staheri/goatlib/instrument"
	"os"

)


func MeasureCoverage(parseResult *trace.ParseResult, concUsage []*instrument.ConcurrencyUsage) {
	// Variables
	concStackIDs := []uint64{}
	concStackTable := make(map[uint64]*instrument.ConcurrencyUsage)

	// check
	fmt.Println("conc Usage >>>>>>>>>>>")
	for _,cu := range(concUsage){
		fmt.Println(cu.String())
	}
// we have lines of codes
// find concurrency usages
// iterate over stack traces
	for stack_id,frms := range(parseResult.Stacks){
		// iterate over frames

		for _,frm := range(frms){
			// iterate concUsage
			for _,cu := range(concUsage){
				// check file and line
				//fmt.Printf("CHECK OK\nCU File:%s\nStack file:%s\n",cu.OrigLoc.Filename,frm.File)
				if cu.OrigLoc.Filename == frm.File {
					fmt.Println("file ok")
					if cu.OrigLoc.Line == frm.Line{
						fmt.Println("line ok")
						if _,ok := concStackTable[stack_id] ; !ok{
							cu.OrigLoc.Function = frm.Fn
							concStackTable[stack_id] = cu
							concStackIDs = append(concStackIDs,stack_id)
						}
					}
				}
			}
		}
	}

	fmt.Println(concStackIDs)
	t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  //t.AppendHeader(table.Row{"Stack ID","File","Function","Line","Event","CU Event","Position"})
	t.AppendHeader(table.Row{"Stack ID","Function","Line","Event","G","CU Event","Position"})

	for _,e := range(parseResult.Events){
		fmt.Println(e.String())
		ed := trace.EventDescriptions[e.Type]
		// check for HB unblock
		// check for concurrency usage
		if containsUInt64(concStackIDs,e.StkID){
			fmt.Printf("***CONC***\n-------\n")
			switch concStackTable[e.StkID].Type{
			case instrument.LOCK, instrument.UNLOCK, instrument.RUNLOCK, instrument.RLOCK:
				if !strings.HasPrefix(ed.Name,"Mu") && !contains(CatBLCK,ed.Name){
					continue
				}
			case instrument.SEND, instrument.RECV, instrument.CLOSE:
				if !strings.HasPrefix(ed.Name,"Ch") && !contains(CatBLCK,ed.Name){
					continue
				}
			case instrument.SELECT:
				if !strings.HasPrefix(ed.Name,"Select") && !contains(CatBLCK,ed.Name){
					continue
				}
			case instrument.GO:
				if !strings.HasPrefix(ed.Name,"Go") && !contains(CatBLCK,ed.Name){
					continue
				}
			}
			var row []interface{}
			row = append(row,e.StkID)
			//row = append(row,concStackTable[e.StkID].OrigLoc.Filename)
			row = append(row,concStackTable[e.StkID].OrigLoc.Function)
			row = append(row,concStackTable[e.StkID].OrigLoc.Line)
			row = append(row,ed.Name)
			row = append(row,e.G)
			row = append(row,instrument.ConcTypeDescription[concStackTable[e.StkID].Type])
			row = append(row,GetPositionDesc(e))
			t.AppendRow(row)
		}
	}
	t.Render()



	t = table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"File","Line","Type"})

	for _,cu := range(concUsage){
		var row []interface{}
		row = append(row,cu.OrigLoc.Filename)
		row = append(row,cu.OrigLoc.Line)
		row = append(row,instrument.ConcTypeDescription[cu.Type])
		t.AppendRow(row)
	}
	t.Render()
}
