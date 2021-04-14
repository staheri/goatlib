package db

import (
	_"fmt"
	"github.com/staheri/goatlib/trace"
	_"path"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"log"
	"strings"
	"time"
)

type Sqldb struct{
	sql.DB
}


// initialize DB
func initDB(dbName string) (db *sql.DB) {
  // Initial Connecting to mysql driver
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/")
	if err != nil {
		panic(err)
	}else{
		log.Println("Store: Initial connection established")
	}

	// Creating new database for current experiment
	_,err = db.Exec("CREATE DATABASE "+dbName + ";")
	check(err)
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
	db.SetConnMaxLifetime(20*time.Second)

  return db

}

// Take sequence of events, create a new DB Schema and insert events into tables
func Store(parseResult *trace.ParseResult, dbName string) (db *sql.DB) {
  // Variables
	var err                  error
	var res                  sql.Result
  var eid                  int64
  var tkey                 uint64
	var events               []*trace.Event
	var stacks               map[uint64][]*trace.Frame
	var goCreateStackID      uint64

  // Data structures to store clocks and aux info
  // Init vector clocks
	msgs           := make(map[msgKey]eventPredecessor) // storing (to be) pred of a recv
	links          := make(map[int64]eventPredecessor) // storing (to be) pred of an event
	// Resource clocks
	localClock     := make(map[uint64]uint64) // vc[g]           = local clock
	chanClock      := make(map[uint64]uint64) // chansClock[cid] = channel clock
	wgClock        := make(map[uint64]uint64) // wgsClock[cid]   = wg clock
	mutexClock     := make(map[uint64]uint64) // mutexClock[cid] = mutex clock
  cvClock        := make(map[uint64]uint64) // cvClock[cid] = cv clock

	// decompose parseResults
	events = parseResult.Events
	stacks = parseResult.Stacks

	// init db
  db = initDB(dbName)

	// Create the triple tables (events, stackFrames, Args)
	createTables(db)

	// for the events with resources (channels, mutex, WaitingGroup)
	insertEventResourceStmt, err := db.Prepare("INSERT INTO Events (offset, type, vc , ts, g, p, linkoff, predG, predClk, rid, reid, rval, rclock, stack_id) values (? ,? ,? ,? ,? ,? ,? ,? ,? ,? ,? ,? ,? ,?);")
	check(err)

	insertStackStmt, err := db.Prepare("INSERT INTO StackFrames (stack_id, pc, func, file, line) values (?, ?, ?, ?, ?)")
	check(err)

	insertArgStmt, err   := db.Prepare("INSERT INTO Args (eventID, arg, value) values (?, ?, ?)")
	check(err)

  stmt := auxPrepareStmts(db)

	//cnt := 0

  // Iterate over events and stroe in relational databases

	for _,e := range events{
		if storeIgnore(e){
			continue
		}

		desc := EventDescriptions[e.Type]
		// fresh values for each event
		predG    := sql.NullInt64{}
		predClk  := sql.NullInt64{}
		rid      := sql.NullString{}
		rval     := sql.NullInt64{}
		reid     := sql.NullInt64{}
		rclock   := sql.NullInt64{}
		linkoff  := sql.NullInt64{}

		// Assign local logical clock
		if v,ok := localClock[e.G];ok{
			localClock[e.G] = v + 1
		} else{
			localClock[e.G] = 1
		}

		// Check category of events\
		// Assign resource clocks (channels, WaitingGroups, mutexes, CondVars)
		// Assign predG, predClk for ChRecv
		// Assign predG, predClk for Link
		// Assign rid, rval (if any), rclock for all resources
		if contains(ctgDescriptions[catMUTX].Members, "Ev"+desc.Name){
			// MUTX event
			// Assign mutexClock
			// Assign rid, rval=Null, rclock
			// predG, predClk: null
			tkey = e.Args[0] // muid - rwid
			rid = sql.NullString{Valid:true, String: "M"+strconv.FormatUint(tkey,10)} // muid
			if _,ok := mutexClock[tkey];ok{
				mutexClock[tkey] = mutexClock[tkey] + 1
			} else{
				mutexClock[tkey] = 1
			}
		} else if contains(ctgDescriptions[catCHNL].Members, "Ev"+desc.Name){
			// CHNL event
			// Assign chanClock
			// Assign rid, rval, rclock
			// ChSend? set predG , rval = value
			// ChRecv and MSG[key]? use predG,predClk, else: null,null
			// ChMake/Close? rval = null
			tkey = e.Args[0] // cid
			rid = sql.NullString{Valid:true, String: "C"+strconv.FormatUint(tkey,10)} // cid
			if vvc,ok := chanClock[tkey];ok{
				chanClock[tkey] = vvc + 1
			} else{
				chanClock[tkey] = 1
			}
			rclock = sql.NullInt64{Valid:true, Int64: int64(chanClock[tkey])}
			if desc.Name == "ChRecv"{
				// ignore if it is a blocked recv
				if e.Args[3] == 0{
					chanClock[tkey] = chanClock[tkey] - 1
					rclock = sql.NullInt64{Valid:true, Int64: int64(chanClock[tkey])}
				}
				rval = sql.NullInt64{Valid:true, Int64: int64(e.Args[2])} // message val
				reid = sql.NullInt64{Valid:true, Int64: int64(e.Args[1])} // message eid
				if vv,ok := msgs[msgKey{e.Args[0],e.Args[1],e.Args[2]}] ; ok{
					// A matching sent is found for the recv
					predG    = sql.NullInt64{Valid:true, Int64: int64(vv.g)}
					predClk  = sql.NullInt64{Valid:true, Int64: int64(vv.clock)}
				} //else{ // Receiver without a matching sender }
			}else{
				// ChMake, ChSend, ChClose
				if desc.Name == "ChSend"{
					// ignore if it is a blocked send
					if e.Args[3] == 0{
						chanClock[tkey] = chanClock[tkey] - 1
						rclock = sql.NullInt64{Valid:true, Int64: int64(chanClock[tkey])}
					}

					rval = sql.NullInt64{Valid:true, Int64: int64(e.Args[2])} // message val
					reid = sql.NullInt64{Valid:true, Int64: int64(e.Args[1])} // message eid
					// Set Predecessor for a receive (key to the event: {cid, eid, val})
					if _,ok := msgs[msgKey{e.Args[0],e.Args[1],e.Args[2]}] ; !ok{
						msgs[msgKey{e.Args[0],e.Args[1],e.Args[2]}] = eventPredecessor{e.G, localClock[e.G]}
					} //else{ // a send for this particular message has been stored before }
				}// else{  ChMake. ChClose }
			}
		} else if contains(ctgDescriptions[catWGCV].Members, "Ev"+desc.Name){
			// WGRP event
			// Assign wgsClock
			// Assign rid, rval=(add? val, else? Null), rclock
			// predG, predClk: null

			tkey= e.Args[0] // wgid/cvid
			if strings.HasPrefix(desc.Name,"Cv"){
				rid =  sql.NullString{Valid:true, String: "CV"+strconv.FormatUint(tkey,10)} // cvid
        if vvc,ok := cvClock[tkey];ok{
  				cvClock[tkey] = vvc + 1
  			} else{
  				cvClock[tkey] = 1
  			}
			} else{
				rid =  sql.NullString{Valid:true, String: "W"+strconv.FormatUint(tkey,10)} // wgid
        if vvc,ok := wgClock[tkey];ok{
  				wgClock[tkey] = vvc + 1
  			} else{
  				wgClock[tkey] = 1
  			}
			}

			if desc.Name == "WgAdd"{ // it has a val
				rval = sql.NullInt64{Valid:true, Int64: int64(e.Args[1])} // val
			}
			rclock = sql.NullInt64{Valid:true, Int64: int64(wgClock[tkey])}
			// All resource events are assigned a logical clock based on their id

		} else if e.Link != nil{
			// Set Predecessor for an event (key to the event: TS)
			linkoff = sql.NullInt64{Valid:true, Int64: int64(e.Link.Off)}
			if _,ok := links[int64(e.Link.Off)] ; !ok{
				links[int64(e.Link.Off)] = eventPredecessor{e.G, localClock[e.G]}
			} else{ // the link of current event has been linked to another event before
				panic("Previously linked to another event!")
			}
		}

		// So far, all predecessor values are set,
		// all resource values are set
		// if a recv has found a sender, it is all set
		// Now only check if the current event has a predecessor. If so: set predG, set predClk
		// otherise: everything is null
		if vv,ok := links[int64(e.Off)]; ok{
			// Is there a possibility that this event has resource other than G?
			// No. Events with predecessor links only have G resource
			predG    = sql.NullInt64{Valid:true, Int64: int64(vv.g)}
			predClk  = sql.NullInt64{Valid:true, Int64: int64(vv.clock)}
			if len(e.Args) > 0{
				// For events that has link (according to Go spec), they might
				// have an argument in Args which is the goroutie ID
				// (e.g GoUnblock has the id of goroutine that it unblocks)
				// So we want to save that under rid in the Events table
				tkey= e.Args[0] // g
				rid =  sql.NullString{Valid:true, String: "G"+strconv.FormatUint(tkey,10)} // g
			}
		}
		/*fmt.Printf("INSERT INTO Events (offset=%v, type=%v, vc=%v, ts=%v, g=%v, p=%v, linkoff=%v, predG=%v, predClk=%v, rid=%v, reid=%v, rval=%v, rclock=%v, stkID=%v, src=%v)\n",strconv.Itoa(e.Off),
																					 "Ev"+desc.Name,
																					 strconv.Itoa(int(localClock[e.G])),
																					 strconv.Itoa(int(e.Ts)),
																					 strconv.FormatUint(e.G,10),
																					 strconv.Itoa(e.P),
																					 linkoff,
																					 predG,
																					 predClk,
																					 rid,
																					 reid,
																					 rval,
																					 rclock,
																					 strconv.FormatUint(e.StkID,10))*/
		res,err = insertEventResourceStmt.Exec(strconv.Itoa(e.Off),
																					 "Ev"+desc.Name,
																					 strconv.Itoa(int(localClock[e.G])),
																					 strconv.Itoa(int(e.Ts)),
																					 strconv.FormatUint(e.G,10),
																					 strconv.Itoa(e.P),
																					 linkoff,
																					 predG,
																					 predClk,
																					 rid,
																					 reid,
																					 rval,
																					 rclock,
																					 strconv.FormatUint(e.StkID,10))

		check(err)
		eid, err = res.LastInsertId()
		check(err)

		// insert args
		//insertArgs(eid, e.Args, desc.Args, db)
		if len(e.Args) != 0{
			for i,a := range desc.Args{
				_,err = insertArgStmt.Exec(strconv.FormatInt(eid,10), a, strconv.FormatInt(int64(e.Args[i]),10))
				check(err)
			}
		}

		// Insert Goroutines and Channels
    // insert goroutines
    if desc.Name == "GoCreate" ||  desc.Name == "GoEnd"{
      res, err := stmt["grtnInitStmt"].Query(strconv.FormatUint(e.G,10))
      check(err)
      if res.Next() {
        // this goroutine already has been added
        // do other stuff with it
        if desc.Name == "GoCreate"{
          // this goroutine has been inserted and it creates another goroutine
          // insert child goroutine with (parent_id of current goroutine) (stack createLOC)
          gid := strconv.FormatInt(int64(e.Args[0]),10) // e.Args[0] for goCreate is "g"
          parent_id := e.G
          if len(e.Stk) != 0{
            goCreateStackID = e.StkID
          }else{
						goCreateStackID = 0
					}
          _,err := stmt["grtnInsertStmt"].Exec(gid,parent_id,goCreateStackID)
          check(err)
        } else if desc.Name == "GoEnd"{
          // this goroutine has been inserted before (with create)
          // Now we need to update its row with GoEnd eventID
          gid := e.G
          //q = fmt.Sprintf("UPDATE Goroutines SET ended=%v WHERE gid=%v;",eid,gid)
          //fmt.Printf(">>> Executing %s...\n",q)
          _,err := stmt["grtnUpdEndStmt"].Exec(eid,gid)
          check(err)
        }
      }else{
        if desc.Name == "GoCreate"{
          // this goroutine has not been inserted (no create) and it creates another goroutine
					// first inser parent
          gid := int(e.G) // current G
          parent_id := -1
					if len(e.Stk) != 0{
            goCreateStackID = e.StkID
          }else{
						goCreateStackID = 0
					}

          _,err := stmt["grtnInsertStmt"].Exec(gid,parent_id,goCreateStackID)
          check(err)

          // insert child goroutine with (parent_id of current goroutine)
					parent_id = gid
          //gid = strconv.FormatInt(int64(e.Args[0]),10) // e.Args[0] for goCreate is "g"
					gid = int(e.Args[0])// e.Args[0] for goCreate is "g"


					if len(e.Stk) != 0{
            goCreateStackID = e.StkID
						//createLoc = filterSlash(path.Base(e.Stk[len(e.Stk)-1].File)+":"+ e.Stk[len(e.Stk)-1].Fn + ":" + strconv.Itoa(e.Stk[len(e.Stk)-1].Line))
          }else{
						goCreateStackID = 0
					}

          _,err = stmt["grtnInsertStmt"].Exec(gid,parent_id,goCreateStackID)
          check(err)

        } else{
          // this goroutine has not been inserted before (no create) and started/ended out of nowhere
          panic("GoStart/End before creating...It is not possible!")
        }
      }
      err = res.Close()
      check(err)
    }
	}

	// Iterate over stacks and store
	for id,stk := range(stacks){
		for _,frm := range(stk){
			_,err := insertStackStmt.Exec(id,frm.PC,frm.Fn,frm.File,frm.Line)
			check(err)
		}
	}
	//closing auxiulary statements
	for _,v := range(stmt){
    err = v.Close()
		check(err)
  }

	err = insertEventResourceStmt.Close()
	check(err)
	err = insertArgStmt.Close()
	check(err)
	err = insertStackStmt.Close()
	check(err)
	return db
}

// Create tables for newly created schema db
func createTables(db *sql.DB){
	eventsCreateStmt := `CREATE TABLE Events (
    									id int NOT NULL AUTO_INCREMENT,
    									offset int NOT NULL,
    									type varchar(255) NOT NULL,
											vc int NOT NULL,
    									ts bigint NOT NULL,
    									g int NOT NULL,
    									p int NOT NULL,
											linkoff int,
											predG int,
											predClk int,
											rid varchar(255),
											reid int,
											rval bigint,
											rclock int,
    									stack_id int,
											PRIMARY KEY (id)
											);`
	stkFrmCreateStmt := `CREATE TABLE StackFrames (
    									id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
							    		stack_id int NOT NULL,
							    		pc int NOT NULL,
							    		func varchar(255) NOT NULL,
							    		file varchar(255) NOT NULL,
							    		line int NOT NULL
											);`
	argsCreateStmt :=   `CREATE TABLE Args (
											id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    									eventID int NOT NULL,
    									arg varchar(255) NOT NULL,
    									value bigint NOT NULL);`
	grtnCreateStmt  :=  `CREATE TABLE Goroutines (
    									id int NOT NULL AUTO_INCREMENT,
    									gid int NOT NULL,
											createStack_id int NOT NULL,
    									parent_id int NOT NULL,
											ended int DEFAULT -1,
    									PRIMARY KEY (id)
											);`

	createTable(eventsCreateStmt,"Events",db)
	createTable(stkFrmCreateStmt,"StackFrames",db)
	createTable(argsCreateStmt,"Args",db)
	createTable(grtnCreateStmt,"Goroutines",db)
}

// Create individual tables for schema db
func createTable(stmt , name string, db *sql.DB) () {
	_,err := db.Exec(stmt)
	check(err)
}

// Create a map of goroutine/channel table statement
func auxPrepareStmts(dbs *sql.DB) map[string]*sql.Stmt{
  ret := make(map[string]*sql.Stmt)
  grtnInitStmt, err       := dbs.Prepare("SELECT * FROM Goroutines WHERE gid=?")
	check(err)
  ret["grtnInitStmt"]=grtnInitStmt

	grtnUpdEndStmt, err     := dbs.Prepare("UPDATE Goroutines SET ended=? WHERE gid=?")
	check(err)
	ret["grtnUpdEndStmt"]=grtnUpdEndStmt

	grtnInsertStmt, err     := dbs.Prepare("INSERT INTO Goroutines (gid, parent_id, createStack_id) VALUES (?, ?, ?)")
	check(err)
	ret["grtnInsertStmt"]=grtnInsertStmt

  return ret
}
