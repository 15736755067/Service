package main

import (
	"Service/common"
	"Service/config"
	"Service/log"
	"Service/servehttp"
	"Service/util/countrycode"
	"Service/util/ip2location"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
)

func main() {
	defer func() {
		log.Alertf("%d Quit main()\n", os.Getpid())
		log.Alert(string(debug.Stack()))
	}()
	help := flag.Bool("help", false, "show help")
	port := flag.Int("port", 5050, "port")
	ipPath := flag.String("ippath", "DB24.BIN", "ip path")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	if err := config.LoadConfig(true); err != nil {
		panic(err.Error())
	}

	//log
	var logAdapter, logConfig string
	var logAsync bool
	//logAdapter = config.String("LOG", "adapter")
	//logConfig = config.String("LOG", "jsonconfig")
	//logAsync = config.Bool("LOG", "async")
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

	//load ip2location
	log.Infof("ip path:%v", *ipPath)
	ip2location.Open(*ipPath)

	//http server
	mux := http.DefaultServeMux
	http.HandleFunc("/", ipHander)
	addr := fmt.Sprintf(":%d", *port)
	reqServer := &http.Server{Addr: addr, Handler: mux}
	log.Infof("server start at port:%d", *port)
	err := servehttp.Serve(reqServer)
	if err != nil {
		log.Errorf("start servehttp fail: %v", err)
	}

}

type IpResp struct {
	Response    string `json:"response"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
	Region      string `json:"region"`
	City        string `json:"city"`
	Zipcode     string `json:"zipcode"`
	Areacode    string `json:"areacode"`
}

func ipHander(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ip := q.Get("ip")
	location := ip2location.Get_all(ip)
	log.Infof("[ipHander] location for ip %v: %+v  ", ip, location)
	IpResp := IpResp{
		Response:    "OK",
		CountryCode: countrycode.CountryCode2To3(location.Country_short),
		CountryName: location.Country_long,
		Region:      location.Region,
		City:        location.City,
		Zipcode:     location.Zipcode,
		Areacode:    location.Areacode,
	}

	bs, err := json.Marshal(IpResp)
	if err != nil {
		fmt.Fprintf(w, "%s", `{"response": "ERR"}`)
		return
	}
	fmt.Fprintf(w, "%s", string(bs))
}
