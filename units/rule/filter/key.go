package filter

import (
	"fmt"
	"strings"
	"time"

	"Service/request"
)

const (
	VarLogicOr  = "or"
	VarLogicAnd = "and"
	VarLogicNot = "not"
)

type KeyFunction func(req request.Request) string

var keyFuncMap map[string]KeyFunction

//TODO 修改注释
// ip              调用方IP地址
// country         IP所属国家
// province        IP所属省
// city            IP所属市
// second          当前秒, 0-59
// minute          当前分, 0-59
// hour            当前小时, 0-23
// day             该月的第几天, 1-31
// month           当前月, 1-12
// year            当前年
// wday            当前星期的第几天，0:周日, 1:周一...
// time            当前时间，yyyy-mm-dd hh:mm:ss
// date            当前日期, yyyy-mm-dd
// url             页面的地址
// ua              UserAgent
// rand            1-100的随机数，每次使用都是新的随机数
// slotid		   广告位Id
// adtype		   广告类型(banner/native/video)
// advtcategory	   advt种类(dsp/s2s/union/direct)
// apptype		   应用类型(in-app/browser)
// devicetype	   设备类型(mobile/desktop)
// connectiontype  连接类型(有线连接/Wifi/移动网络)
// os			   操作系统
// osv			   操作系统版本
// width		   广告位宽
// height		   广告位高
// size			   广告位宽*高
// datacenter	   数据中心节点位置(US-WEST,US-EAST等)
//TODO↓↓↓
// get.sys.cid     用户cookie id，用于标志唯一用户
// get.sys.bid     用户分桶ID，0-f，共16个值，值可以使用函数f.bid()，这是一个用日期时间来计算0-f的值的函数
// get.sys.bid2    用户分桶二级ID，0-f，共16个值
// get.sys.si      当前广告位ID
// get.sys.sdk     当前请求的sdk
// get.sys.selector 当前广告来源selector
// get.xx          接口请求中的get参数xx
// get.xx[index]   接口请求中get参数xx的列表值（用逗号分隔）的索引（从0开始）为index的值
// get.xx{n1.n2}   接口请求中get参数为xx的JSON节点为node1/node2的值
// cookie.xx       Cookie中名为xx的值
// request.xx      xx为adid，或字母g为前缀后跟流量组id，可使用多个，使用逗号分隔。对象的总请求量
// d_request.xx    同requext.xx，但仅为当天的总量
// impression.xx   xx同reqest，对象的总展示总量
// d_impression.xx 同impression.xx，但仅为当天的总量
// click.xx        xx同request，对象的总点击量
// d_click.xx      同click.xx，但仅为当天的总量
// aclick.xx       xx同request，对象的弹出总量
// d_aclick.xx     同aclick.xx，当天总量
// f.xx            xx为频次ID，一般为adid
// F.xx            xx为IP频次ID，一般为adid
// user.xx         xx为用户属性
// ip.xx           基于ip的行为频次
func init() {
	keyFuncMap = map[string]KeyFunction{
		"brand": func(req request.Request) string {
			return strings.ToUpper(req.Brand())
		},
		"model": func(req request.Request) string {
			return strings.ToUpper(req.Brand() + " " + req.Model())
		},
		"browser": func(req request.Request) string {
			return strings.ToUpper(req.Browser() + " " + req.BrowserVersion())
		},
		"connection": func(req request.Request) string {
			return strings.ToUpper(req.ConnectionType())
		},
		"country": func(req request.Request) string {
			return strings.ToUpper(req.CountryCode())
		},
		"region": func(req request.Request) string {
			return strings.ToUpper(req.Region())
		},
		"city": func(req request.Request) string {
			return strings.ToUpper(req.City())
		},
		"timeOfDay": func(req request.Request) string { // 实际上并不会被使用
			return time.Now().UTC().Format("15:04")
		},
		"weekday": func(req request.Request) string { // 实际上并不会被使用
			return fmt.Sprintf("%d", time.Now().UTC().Weekday())
		},
		"day": func(req request.Request) string {
			return fmt.Sprintf("%d", time.Now().UTC().Day())
		},
		"month": func(req request.Request) string {
			return fmt.Sprintf("%d", time.Now().UTC().Month())
		},
		"year": func(req request.Request) string {
			return fmt.Sprintf("%d", time.Now().UTC().Year())
		},
		"time": func(req request.Request) string {
			return time.Now().UTC().Format("2006-01-02 15:04:05")
		},
		"date": func(req request.Request) string {
			return time.Now().UTC().Format("2006-01-02")
		},
		"device": func(req request.Request) string {
			return strings.ToUpper(req.DeviceType())
		},
		"iprange": func(req request.Request) string {
			return req.RemoteIp()
		},
		"isp": func(req request.Request) string {
			return strings.ToUpper(req.ISP())
		},
		"os": func(req request.Request) string {
			return strings.ToUpper(req.OS() + " " + req.OSVersion())
		},
		"language": func(req request.Request) string {
			return strings.ToUpper(req.Language())
		},
		"carrier": func(req request.Request) string {
			return strings.ToUpper(req.Carrier())
		},
		"referrer": func(req request.Request) string {
			return strings.ToUpper(req.Referrer())
		},
		"useragent": func(req request.Request) string {
			return strings.ToUpper(req.UserAgent())
		},
		"var1": func(req request.Request) string {
			return strings.ToUpper(req.Vars(0))
		},
		"var2": func(req request.Request) string {
			return strings.ToUpper(req.Vars(1))
		},
		"var3": func(req request.Request) string {
			return strings.ToUpper(req.Vars(2))
		},
		"var4": func(req request.Request) string {
			return strings.ToUpper(req.Vars(3))
		},
		"var5": func(req request.Request) string {
			return strings.ToUpper(req.Vars(4))
		},
		"var6": func(req request.Request) string {
			return strings.ToUpper(req.Vars(5))
		},
		"var7": func(req request.Request) string {
			return strings.ToUpper(req.Vars(6))
		},
		"var8": func(req request.Request) string {
			return strings.ToUpper(req.Vars(7))
		},
		"var9": func(req request.Request) string {
			return strings.ToUpper(req.Vars(8))
		},
		"var10": func(req request.Request) string {
			return strings.ToUpper(req.Vars(9))
		},
	}
}

func KF(key string) KeyFunction {
	return keyFuncMap[key]
}
