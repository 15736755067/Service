package filter

import (
	"fmt"
	"net"
	"strings"
	"time"

	"Service/request"
	"Service/util/timezone"
)

type OperationFunction func(key string, expr []string, req request.Request) bool

var logicOpFunctions map[string]OperationFunction

var actionOpFunctions map[string]OperationFunction

func LOF(op string) OperationFunction {
	return logicOpFunctions[op]
}

func AOF(op string) OperationFunction {
	return actionOpFunctions[op]
}

// 目前逻辑操作符支持以下几种
// in,not in,(不区分大小写)
// time between,time not between,
// weekday in,weekday not in,
// contain,not contain,(不区分大小写)
func init() {
	logicOpFunctions = map[string]OperationFunction{
		"in": func(value string, expr []string, req request.Request) bool {
			// #All,#SameAsCampaign指令也处理了
			for _, s := range expr {
				s = strings.ToUpper(s)
				if s == "#SAMEASCAMPAIGN" {
					s = req.CampaignCountry()
					if s == "ZZZ" { // CampaignCountry为Global
						//return true // 永远为真
						continue // 忽略
					}
				}
				if strings.HasPrefix(s, "#ALL ") {
					ts := strings.TrimPrefix(s, "#ALL ")
					if ts == "" {
						continue
					}
					if strings.HasPrefix(value, ts+" ") {
						return true
					}
					continue
				}
				if value == s {
					return true
				}
			}
			return false
		},
		"not in": func(value string, expr []string, req request.Request) bool {
			// #All,#SameAsCampaign指令也处理了
			for _, s := range expr {
				s = strings.ToUpper(s)
				if s == "#SAMEASCAMPAIGN" {
					s = req.CampaignCountry()
					if s == "ZZZ" { // CampaignCountry为Global
						//return false // 永远为假
						continue // 忽略
					}
				}
				if strings.HasPrefix(s, "#ALL ") {
					ts := strings.TrimPrefix(s, "#ALL ")
					if ts == "" {
						continue
					}
					if strings.HasPrefix(value, ts+" ") {
						return false
					}
					continue
				}
				if value == s {
					return false
				}
			}
			return true
		},
		"time between": func(value string, expr []string, req request.Request) bool {
			if len(expr) < 3 {
				return false
			}
			return timezone.IsZoneTimeBetween(time.Now(), expr[0], expr[1], expr[2])
		},
		"time not between": func(value string, expr []string, req request.Request) bool {
			if len(expr) < 3 {
				return false
			}
			return !timezone.IsZoneTimeBetween(time.Now(), expr[0], expr[1], expr[2])
		},
		"weekday in": func(value string, expr []string, req request.Request) bool {
			if len(expr) < 2 {
				return false
			}
			value = fmt.Sprintf("%d", timezone.TimeInZone(time.Now(), "", expr[0]).Weekday())
			for _, wday := range expr[1:] {
				if value == wday {
					return true
				}
			}
			return false
		},
		"weekday not in": func(value string, expr []string, req request.Request) bool {
			if len(expr) < 2 {
				return true
			}
			value = fmt.Sprintf("%d", timezone.TimeInZone(time.Now(), "", expr[0]).Weekday())
			for _, wday := range expr[1:] {
				if value == wday {
					return false
				}
			}
			return true
		},
		"contain": func(value string, expr []string, req request.Request) bool {
			for _, e := range expr {
				if strings.Contains(value, strings.ToUpper(e)) {
					return true
				}
			}
			return false
		},
		"not contain": func(value string, expr []string, req request.Request) bool {
			for _, e := range expr {
				if strings.Contains(value, strings.ToUpper(e)) {
					return false
				}
			}
			return true
		},
		"ip in": func(value string, expr []string, req request.Request) bool {
			//TODO expr传入ipNetwork的slice进来
			ip := net.ParseIP(value)
			for _, ipRange := range expr {
				if !strings.Contains(ipRange, "/") {
					if value == ipRange {
						return true
					}
					continue
				}
				_, ipNetwork, _ := net.ParseCIDR(ipRange)
				if ipNetwork == nil {
					continue
				}
				if ipNetwork.Contains(ip) {
					return true
				}
			}
			return false
		},
		"ip not in": func(value string, expr []string, req request.Request) bool {
			ip := net.ParseIP(value)
			for _, ipRange := range expr {
				if !strings.Contains(ipRange, "/") {
					if value == ipRange {
						return false
					}
					continue
				}
				_, ipNetwork, _ := net.ParseCIDR(ipRange)
				if ipNetwork == nil {
					continue
				}
				if ipNetwork.Contains(ip) {
					return false
				}
			}
			return true
		},
	}
}
