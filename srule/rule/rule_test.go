package rule

import (
	"Service/config"
	"Service/log"
	"flag"
	"testing"
)

func init() {
	help := flag.Bool("help", false, "show help")
	flag.Parse()
	flag.Set("config", "/Users/robin/Program/gopath/src/AdClickService/src/Service/srule/srule.ini")
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
}

func TestProcessRuleBySchedule(t *testing.T) {
	schedule := "0 0 10 9 5 *"
	ProcessRuleBySchedule(schedule)
}

func TestProcessRuleByOneTime(t *testing.T) {
	oneTime := "2017-05-09T11"
	ProcessRuleByOneTime(oneTime)
}
