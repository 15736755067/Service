//
//
package main

import (
	"errors"
	"sync"

	"Service/gracequit"
	//	"Service/log"
	//	"Service/util/cron"
)

type JobFunction func(gracequit.StopSigChan)

var jmx sync.Mutex
var allJobs = make([]JobFunction, 0)

func RegisterJob(f JobFunction) error {
	if f == nil {
		return errors.New("Job is nil")
	}
	jmx.Lock()
	defer jmx.Unlock()
	allJobs = append(allJobs, f)
	return nil
}

//func main() {

//	for _, f := range allJobs {
//		gracequit.StartGoroutine(f)
//	}
//}
