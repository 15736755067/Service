package glog

// wrap glog as LoggerInterface in logs of beego
//
//    logs In Beego      V.S.  glog
//    LevelEmergency     =>    fatalLog
//    LevelAlert         =>    fatalLog
//    LevelCritical      =>    fatalLog
//    LevelError         =>    errorLog
//    LevelWarning       =>    warningLog
//    LevelNotice        =>    warningLog
//    LevelInformational =>    infoLog
//    LevelDebug         =>    infoLog

import (
	"encoding/json"
	"errors"
	"fmt"

	"Service/log/logs"
)

type GlogLogWriter struct {
	Dir   string `json:"dir"` // the directory where log files are stored in
	Level int    `json:"level"`
}

func (glw *GlogLogWriter) Init(jsonconfig string) error {
	err := json.Unmarshal([]byte(jsonconfig), glw)
	if err != nil {
		return err
	}
	if len(glw.Dir) == 0 {
		return errors.New("jsonconfig must have dir name")
	}
	SetLogDir(glw.Dir)

	return err
}

func (glw *GlogLogWriter) WriteMsg(msg string, level int) error {
	//fmt.Printf("GlogLogWriter.WriteMsg\tglw.Level[%d]\tmsg[%v]\tlevel[%d]\n", glw.Level, msg, level)

	if level > glw.Level {
		// ignore
		return nil
	}

	switch level {
	case logs.LevelEmergency:
		fallthrough
	case logs.LevelAlert:
		fallthrough
	case logs.LevelCritical:
		// fatalLog
		fmt.Println("glog;fatal log")
		Fatal(msg)
	case logs.LevelError:
		// errorLog
		fmt.Println("glog;error log")
		Error(msg)
	case logs.LevelWarning:
		fallthrough
	case logs.LevelNotice:
		// warningLog
		fmt.Println("glog;warning log")
		Warning(msg)
	case logs.LevelDetail:
		fallthrough
	case logs.LevelDebug:
		fallthrough
	case logs.LevelInfo:
		// infoLog
		fmt.Println("glog;info log")
		Info(msg)
	}
	return nil
}

func (glw *GlogLogWriter) Destroy() {
	Exitln("Exit")
}

func (glw *GlogLogWriter) Flush() {
	Flush()
}

func NewGlogLogWriter() logs.LoggerInterface {
	glw := &GlogLogWriter{
		Dir:   "",
		Level: logs.LevelTrace,
	}

	return glw
}

func init() {
	logs.Register("glog", NewGlogLogWriter)
}
