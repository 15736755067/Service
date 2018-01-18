package config

import (
	"flag"
	"os"
	//"path/filepath"
	"fmt"
	"sync"

	"github.com/robfig/config"

	"Service/log"
	"time"
)

var instance *config.Config
var configLock sync.RWMutex

var configfile = "development"

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "..\\config\\development.ini", "specify the configuration file")
}

func LoadConfig(isFirstTime bool) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if isFirstTime {
				panic(e)
			}
			log.Errorf("load config fail: %v\n", e)
		}
	}()

	if configPath != "" {
		// 优先使用参数中指定的config文件
		configfile = configPath
	} /* else {
		if !filepath.IsAbs(*rootPath) {
			absPath, err := filepath.Abs(*rootPath)
			if err != nil {
				panic("Convert root path to abs path failed")
			}
			*rootPath = absPath
		}

		configLock.Lock()
		defer configLock.Unlock()

		//环境变量加载不同配置
		mode := os.Getenv("adbund_mode")
		if mode != "" {
			configfile = mode
		}

		wd, _ := os.Getwd()

		if file := wd + "/config/" + configfile + ".ini"; isExist(file) {
			configfile = file
		}

		if file := *rootPath + "/config/" + configfile + ".ini"; isExist(file) {
			configfile = file
		}

		if file := os.Getenv("adbund_file"); file != "" && isExist(file) {
			configfile = file
		}
	}*/

	if !isExist(configfile) {
		log.Errorf("Config file not found %s\n", configfile)
		return
	}

	fmt.Println("Config file:", configfile)
	configLock.Lock()
	instance, err = config.ReadDefault(configfile)
	configLock.Unlock()
	if err != nil {
		log.Errorf("Config file ReadDefault error:%s\n", err.Error())
		return
	}
	log.Infof("load config file: %s\n", configfile)

	InitVars()
	return
}

func isExist(filePath string) bool {
	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func ReloadConfig() {
	LoadConfig(false)
}

//String get config string value
func String(section, key string) string {
	if instance == nil {
		LoadConfig(false)
	}
	configLock.RLock()
	defer configLock.RUnlock()

	v, err := instance.String(section, key)

	if err != nil {
		log.Errorf("Read config string value for section:%v key:%v error:%v\n", section, key, err)
	}

	return v
}

//Bool  get config bool value
func Bool(section, key string) bool {
	if instance == nil {
		LoadConfig(false)
	}
	configLock.RLock()
	defer configLock.RUnlock()

	v, err := instance.Bool(section, key)
	if err != nil {
		log.Errorf("Read config bool value for section:%v key:%v error:%v", section, key, err)
	}

	return v
}

//Int get config int value
func Int(section, key string) int {
	if instance == nil {
		LoadConfig(false)
	}
	configLock.RLock()
	defer configLock.RUnlock()

	v, err := instance.Int(section, key)
	if err != nil {
		log.Errorf("Read config int value for section:%v, key:%v error:%v\n", section, key, err)
	}
	return v
}

//Float  get config float value
func Float(section, key string) float64 {
	if instance == nil {
		LoadConfig(false)
	}
	configLock.RLock()
	defer configLock.RUnlock()

	v, err := instance.Float(section, key)
	if err != nil {
		log.Errorf("Read config float value for section:%v key:%v error:%v\n", section, key, err)
	}

	return v
}

func GetEnginePort() string {
	return String("DEFAULT", "engineport")
}

func GetPostbackPort() string {
	return String("DEFAULT", "postbackport")
}

func GetSvrUniqueStr() string {
	return String("DEFAULT", "unique")
}

// InitVars 初始化vars.go里面的变量
func InitVars() {
	reqcachetime := Int("[DEFAULT", "reqcachetime")
	if reqcachetime > 0 {
		ReqCacheTime = time.Duration(reqcachetime) * time.Second
	}
}
