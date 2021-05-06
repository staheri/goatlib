package traceops

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
*/
