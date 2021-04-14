package instrument


func InstrumentCriticalPoints(origpath,newpath string) []*ConcurrencyUsage{
	criticalPoints := Identify(origpath)
	_ = rewrite_randomSched(origpath,newpath,criticalPoints)

	// to update source lines of critical points after rewrite
	criticalPoints = Identify(newpath)
	/*for _,c := range(criticalPoints){
		fmt.Println(c.String())
	}*/
	return criticalPoints
}


func InstrumentTraceOnly(origpath, newpath string) bool{
	return len(rewrite_traceOnly(origpath,newpath))>0
}
