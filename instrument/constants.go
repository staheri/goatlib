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
  RWLOCK
  RWUNLOCK
  SEND
  RECV
  CLOSE
  GO
  WAIT
  ADD
  SIGNAL
  BROADCAST
  SELECT
  RANGE1
  RANGE2
)

var selectorIdents = []string{"Lock", "Unlock", "Wait", "Add", "Signal", "Broadcast","Done"}
