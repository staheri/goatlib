package traceops
import (
	"strings"
	_"github.com/jedib0t/go-pretty/table"
	"github.com/staheri/goatlib/trace"
	_"log"
	"fmt"
  "path/filepath"
	"log"
)

const(
  ROOT           =iota
  MAIN
  TRACE
  APP
  OTHER
)

var gtypes = []string{"ROOT","MAIN","TRACE","APP","OTHER"}


type GInfo struct{
  gid               uint64
	parent_id         uint64
  createStack_id    uint64
  createStack_frame []*trace.Frame
  ended             bool
  gtype             int
  lastEvent         *trace.Event
}

type GoroutineInfo struct{
	main          *GInfo
	trace        	*GInfo
	app           []*GInfo
  goat          []*GInfo
}

// returns goroutine info
func GetGoroutineInfo(parseResult *trace.ParseResult) (GoroutineInfo , map[uint64]*GInfo){
	var ret               GoroutineInfo
	var notAppGs          []uint64
	var appGs             []uint64
	//var allGs             []uint64
  //var goatGs            []uint64
//	var gid               uint64


  gmap := make(map[uint64]*GInfo)
  // iterate over events
  for _,e := range(parseResult.Events){
    desc := trace.EventDescriptions[e.Type]
		//fmt.Println("Event Id:",ii,desc.Name)
		//fmt.Println(e.String())
    // first store gmap info
    if gi,ok:=gmap[e.G];ok{
      // process event for gmap
			//fmt.Printf(">>> G %v is in the map\n",e.G)
      switch desc.Name {
      case "GoCreate":
				//fmt.Printf(">>> G %v is creating G %d\n",e.G,e.Args[0])
        if _,ok:=gmap[e.Args[0]];!ok{
          gmap[e.Args[0]] = &GInfo{gid:e.Args[0],
                                   parent_id:e.G,
                                   createStack_id:e.StkID,
                                   createStack_frame: e.Stk,
																 	 gtype:OTHER}
					//fmt.Printf(">>> Add G %v to gmap as OTHER\n",e.Args[0])

        } else{
          // the child that current e creates has already been existed in the map
          panic("child already in the map")
        }
      case "GoEnd":
				//fmt.Printf(">>> G %v has Ended - update\n",e.G)
        gi.ended = true
        gi.lastEvent = e
        gmap[e.G] = gi
        //if gii,ok:=gs[e.G];ok{
        //} else{
          // the child that current e creates has already been existed in the map
          //panic(fmt.Sprintf("unexplored goroutine %v ended\n",e.G))
        //}
      default:
				//fmt.Println("DEFAULT")
				//fmt.Printf(">>> G %v Default - update last event\n",e.G)
        gi.lastEvent = e
        gmap[e.G] = gi
      }

    } else{
      // a new goroutine without getting created (probably g=0)
			//fmt.Printf("##### G %v is not in the map\n",e.G)
      if desc.Name == "GoCreate"{
				//fmt.Printf("##### Add G %v as ROOT\n",e.G)
        gmap[e.G] = &GInfo{gid:e.G,gtype:ROOT}
				//fmt.Println("ADDED G0",gmap[e.G].String())
				//fmt.Printf("##### G %v is creating G %d\n",e.G,e.Args[0])
        if _,ok:=gmap[e.Args[0]];!ok{
          gmap[e.Args[0]] = &GInfo{gid:e.Args[0],
                                 parent_id:e.G,
                                 createStack_id:e.StkID,
                                 createStack_frame: e.Stk,
															 	 gtype:OTHER}
					//fmt.Printf("##### Add G %v to gmap as OTHER\n",e.Args[0])
        } else{
          // the child that current e creates has already been existed in the map
          panic("child already in the map")
        }
      }else{
        panic(fmt.Sprintf("unexplored goroutine %v captured %v\n",e.G,desc.Name))
      }
    }

		// it is guaranteed the refrenced ids are initialized
    // process GoroutineInfo

		// iterate over each events stack to see
    for _,frm := range(e.Stk){
      if strings.HasPrefix(frm.Fn,"github.com/staheri/goat/goat.Start"){
        // gid is not app-related
        notAppGs = append(notAppGs,e.Args[0])
      }
		}// end stack check
		if len(e.Stk) != 0{
			frm := e.Stk[0]
      if strings.HasPrefix(frm.Fn,"runtime/trace.Start") && ret.trace == nil{
				if desc.Name != "GoCreate"{
					panic("trace and main identified in a non-GoCreate event")
				}
				//fmt.Println(">>>> Trace has found",e.Args[0])
        rt,ok := gmap[e.Args[0]]
        if !ok{
          panic(fmt.Sprintf("%v key not exist in gmap",e.G))
        }
        ret.trace = rt

        ret.trace.gtype = TRACE
        gmap[e.Args[0]] = ret.trace

				//fmt.Println(">>>> Main Has Found",gmap[e.Args[0]].parent_id)
        ret.main, ok = gmap[gmap[e.Args[0]].parent_id]
        if !ok{
          panic(fmt.Sprintf("%v key (parent of trace: %v) not exist in gmap",gmap[e.Args[0]].parent_id,e.G))
        }
        ret.main.gtype = MAIN
        gmap[gmap[e.Args[0]].parent_id] = ret.main
      }
		}
  } // end gmap
  // anything that is not in notApp, then it is app
	//fmt.Println("NOT APP GS: ",notAppGs)

  for g,ginf := range(gmap){
		// make sure all gs have last events
		if ginf.lastEvent == nil{
			log.Printf("WARNING: G %v has no last event\n",g)
			fmt.Printf("WARNING: G %v has no last event\n",g)
		}
		isApp := true
		if len(parseResult.Stacks[ginf.createStack_id]) > 0 {
			for _,frm := range(parseResult.Stacks[ginf.createStack_id]){
				if strings.HasPrefix(frm.Fn,"github.com/staheri/goat/goat.Start"){
					// this is not app
					isApp = false
					break
				}
			}
		} else{
			// it is root
			isApp = false
		}

		if isApp{
			//if !containsUInt64(notAppGs,g){
			//fmt.Printf(">>> g%d is not in NotApps- it is app\n",g)
			if !containsUInt64(appGs,g){
				ginf.gtype = APP
				gmap[g] = ginf
	      ret.app = append(ret.app,ginf)
				appGs = append(appGs,g)
			}
		}
	}



	//StackTable(parseResult.Stacks)
	//GoroutineTable(gmap)
	return ret,gmap
}

// convert stack trace ([]stack frames) to string
func stackToString (frames []*trace.Frame, isViz bool) string{
	s := ""
	for i:= len(frames)-1 ; i>=0 ; i--{
		if isViz{
			s = s + fmt.Sprintf("%v\n",ToStringViz(frames[i]))
		} else{
			s = s + fmt.Sprintf("%v\n",ToString(frames[i]))
		}

	}
	return s
}

// convert stack frame to string (for execViz)
func ToStringViz(f *trace.Frame) string {
	fu := strings.Split(f.Fn,"/")
	return fmt.Sprintf("%s @ %s:%d ",fu[len(fu)-1],filepath.Base(f.File),f.Line)
}

// convert stack frame to string
func ToString(f *trace.Frame) string {
	return fmt.Sprintf("%s\n\t%s:%d ",f.Fn,f.File,f.Line)
}

// returns individual ginfo string
func (ginf *GInfo) String() string{
  s := fmt.Sprintf("G: %v\n",ginf.gid)
  s = s + fmt.Sprintf("Parent: %v\n",ginf.parent_id)
  s = s + fmt.Sprintf("Ended: %v\n",ginf.ended)
	s = s + fmt.Sprintf("Type: %v\n",gtypes[ginf.gtype])
  s = s + fmt.Sprintf("Create StackFrame:\n%v\n",stackToString(ginf.createStack_frame,false))
	if ginf.lastEvent != nil{
		s = s + fmt.Sprintf("Last Event: %v\n",trace.EventDescriptions[ginf.lastEvent.Type])
	}
  return s
}

// returns a detail report of execution goroutine structure
func (ginf *GoroutineInfo) StringDetail() string{
	s := fmt.Sprintf("Main: \n%v\n",ginf.main.String())
	s = s +  fmt.Sprintf("Trace: \n%v\n",ginf.trace.String())
	for _,gi := range(ginf.app){
		s = s +  fmt.Sprintf("App: \n%v\n---\n",gi.String())
	}
  // for _,gi := range(ginf.goat){
	// 	s = s +  fmt.Sprintf("GOAT: \n%v\n---\n",gi.String())
	// }
	return s
}

// returns a short report of execution goroutine structure
func (ginf *GoroutineInfo) String() string{
	s := fmt.Sprintf("Main: %v\n",ginf.main.gid)
	s = s +  fmt.Sprintf("Trace: %v\n",ginf.trace.gid)
	for _,gi := range(ginf.app){
		s = s +  fmt.Sprintf("App: %v\n",gi.gid)
	}
  for _,gi := range(ginf.goat){
		s = s +  fmt.Sprintf("GOAT: %v\n",gi.gid)
	}
	return s
}
