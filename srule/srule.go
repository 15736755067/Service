package main

import (
	"Service/bnotify"
	"Service/common"
	"Service/config"
	"Service/log"
	"Service/servehttp"
	"Service/srule/rule"
	"Service/util/cron"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var c *cron.Cron

type jobManager struct {
	name   string
	jobMap map[string]interface{}
	mux    sync.RWMutex
}

const CSchedule = "schedule"
const COnetime = "onetime"

var scheduleJob *jobManager = NewJobManager(CSchedule)
var onetimeJob *jobManager = NewJobManager(COnetime)

func NewJobManager(name string) *jobManager {
	return &jobManager{
		name:   name,
		jobMap: make(map[string]interface{}),
	}
}

func (m *jobManager) listJobSpec() []string {
	m.mux.RLock()
	defer m.mux.RUnlock()
	ret := make([]string, 0)
	for spec := range m.jobMap {
		ret = append(ret, spec)
	}
	return ret
}

func (m *jobManager) addJob(spec string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.jobMap[spec]; ok {
		return
	} else {
		var jobid string
		var err error
		if m.name == CSchedule {
			jobid, err = c.AddFunc(spec, ProcessSchedule(spec))
		}

		if err != nil {
			log.Infof("add job fails %v", spec)
		} else if jobid != "" {
			m.jobMap[spec] = jobid
			log.Infof("add job %v", spec)
		}
		return
	}
}

func (m *jobManager) removeJob(spec string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if jobid, ok := m.jobMap[spec]; ok {
		if m.name == CSchedule {
			delete(m.jobMap, spec)
			c.RemoveJob(jobid.(string))
			log.Infof("remove job %v", spec)
		}
		return
	}
	return
}

func ProcessSchedule(schedule string) func() {
	schedule2 := schedule

	ret := func(schedule string) func() {
		return func() {
			rule.ProcessRuleBySchedule(schedule2)
		}
	}(schedule)
	return ret
}

func ProcessOneTime(oneTime string) func() {
	ret := func(oneTime string) func() {
		return func() {
			rule.ProcessRuleByOneTime(oneTime)
		}
	}(oneTime)
	return ret
}

func intCron() {
	c = cron.New()
	c.Start()
}

func updateCron() {
	// update schedule
	scheduleList := rule.GetScheduleList()
	if len(scheduleList) > 0 {
		for _, spec := range scheduleList {
			if _, err := cron.Parse(spec); err == nil {
				schedule2 := spec
				scheduleJob.addJob(schedule2)
			} else {
				log.Infof("spec %v,err %v", spec, err)
				continue
			}
		}

	}

	//TODO: 删除job
	if len(scheduleJob.jobMap) == 0 {
		return
	}

	for _, spec := range scheduleJob.listJobSpec() {
		exist := false
		for _, dbSpec := range scheduleList {
			if spec == dbSpec {
				exist = true
				break
			}
		}

		if !exist {
			scheduleJob.removeJob(spec)
		}
	}

	//entries := c.Entries()
	//for _, entry := range entries {
	//	fmt.Printf("Entry %+v,%+v\n", entry.Schedule, entry.Job)
	//}

}

func updateOneTime() {
	oneTimeList := rule.GetOneTimeList()
	if len(oneTimeList) == 0 {
		return
	}
	onetimeJob.mux.Lock()
	for _, oneTime := range oneTimeList {
		if _, ok := onetimeJob.jobMap[oneTime]; ok {
			continue
		}

		f := ProcessOneTime(oneTime)
		dur := rule.GetTime(oneTime).Sub(time.Now())
		t := time.AfterFunc(dur, f)

		onetimeJob.jobMap[oneTime] = t
		log.Infof("add oneTime %v", oneTime)
	}
	onetimeJob.mux.Unlock()

	//remove onetime
	if len(onetimeJob.jobMap) != 0 {
		for _, jobSpec := range onetimeJob.listJobSpec() {
			exist := false
			for _, dbSpec := range oneTimeList {
				if jobSpec == dbSpec {
					exist = true
					break
				}
			}

			if !exist {
				job := onetimeJob.jobMap[jobSpec]
				job.(*time.Timer).Stop()
				delete(onetimeJob.jobMap, jobSpec)
				log.Infof("remove oneTime %v", jobSpec)
			}
		}
	}

}

func getTimerDur() time.Duration {
	now := time.Now()
	timerMarkMinute := config.Int("DEFAULT", "timer_mark_minute")
	timerMark := time.Duration(timerMarkMinute) * time.Minute
	next := now.Truncate(timerMark).Add(timerMark).Add(-time.Duration(5) * time.Second)
	var timerDur time.Duration
	if next.Before(now) {
		timerDur = time.Second
	} else {
		timerDur = next.Sub(now)
	}

	return timerDur
}
func main() {
	help := flag.Bool("help", false, "show help")
	port := flag.Int("port", 5050, "port")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	if err := config.LoadConfig(true); err != nil {
		panic(err.Error())
	}

	//log
	logAdapter := config.String("LOG", "adapter")
	logConfig := config.String("LOG", "jsonconfig")
	logAsync := config.Bool("LOG", "async")
	if logAdapter == "" {
		logAdapter = "console"
	}
	if logConfig == "" {
		logConfig = `{"level":7}`
	}
	logConfig = `{"level":7}`
	log.Init(logAdapter, logConfig, logAsync)
	defer func() {
		log.Flush()
	}()

	common.WritePidFile()
	bnotify.Start()
	// check db
	if rule.InitDB() != nil {
		panic("init db fails")
	}

	// start cron
	intCron()
	defer c.Stop()

	go func() {
		updateCron()
		updateOneTime()
		mysqlTickMinute := config.Int("DEFAULT", "mysql_tick_minute")
		ticker := time.NewTicker(time.Duration(mysqlTickMinute) * time.Minute)

		timer := time.NewTimer(getTimerDur())
		for {
			select {
			case <-ticker.C:
				updateCron()
				updateOneTime()
			case <-timer.C:
				updateCron()
				updateOneTime()
				timer.Reset(getTimerDur())
			}
		}
	}()

	mux := http.DefaultServeMux
	http.HandleFunc("/status", status)
	http.HandleFunc("/job/update", updateJob)
	addr := fmt.Sprintf(":%d", *port)
	reqServer := &http.Server{Addr: addr, Handler: mux}
	log.Infof("server start at port:%d", *port)
	err := servehttp.Serve(reqServer)
	if err != nil {
		log.Errorf("start servehttp fail: %v", err)
	}

}

func updateJob(w http.ResponseWriter, r *http.Request) {
	updateCron()
	updateOneTime()
}

func status(w http.ResponseWriter, r *http.Request) {

	resp := ""
	scheduleList := scheduleJob.listJobSpec()

	resp = resp + fmt.Sprintf("----schedule----\n")
	for _, schedule := range scheduleList {
		resp = resp + fmt.Sprintf("%v\n", schedule)
	}

	onetimeList := onetimeJob.listJobSpec()
	resp = resp + fmt.Sprintf("\n----onetime----\n")
	for _, onetime := range onetimeList {
		resp = resp + fmt.Sprintf("%v\n", onetime)
	}
	fmt.Fprint(w, resp)
}
