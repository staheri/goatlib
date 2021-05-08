package instrument

var DEBUG bool
var VERBOSE bool
var TIMING bool

const (
  LOGPREFIX = "INSTRUMENT:"
)

const(
  LOCK     = iota
  UNLOCK
  RLOCK
  RUNLOCK
  SEND
  RECV
  CLOSE
  GO
  WAIT
  ADD
  DONE
  SIGNAL
  BROADCAST
  SELECT
  RANGE
  COUNT
)

var selectorIdents = []string{
  "Lock",
  "Unlock",
  "RLock",
  "RUnlock",
  "Wait",
  "Add",
  "Done",
  "Signal",
  "Broadcast",
  "Done"}


var ConcTypeDescription = [COUNT]string{
  "LOCK",
  "UNLOCK",
  "RLOCK",
  "RUNLOCK",
  "SEND",
  "RECV",
  "CLOSE",
  "GO",
  "WAIT",
  "ADD",
  "DONE",
  "SIGNAL",
  "BROADCAST",
  "SELECT",
  "RANGE",
}
