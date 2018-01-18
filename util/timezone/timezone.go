package timezone

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"Service/log"
)

var osreg = regexp.MustCompile(`^(\+|\-)((0[0-9]|10|11)\:[0-5][0-9]|12\:00)$`)

func offsetValid(offset string) bool {
	return osreg.Match([]byte(offset))
}

var hmreg = regexp.MustCompile(`^((0[0-9]|1[0-9]|2[0-3])\:[0-5][0-9])$`)

func hmValid(hm string) bool {
	return hmreg.Match([]byte(hm))
}

// offset:+08:00/+05:45/-08:00
// return:seconds east of UTC
func zoneOffset(offset string) int {
	ts := strings.Split(offset[1:], ":")
	var h, m int64
	var err error
	if h, err = strconv.ParseInt(ts[0], 10, 64); err != nil {
		log.Errorf("zoneOffset(%s) hour error:%s\n", offset, err.Error())
		return 0
	}
	if m, err = strconv.ParseInt(ts[1], 10, 64); err != nil {
		log.Errorf("zoneOffset(%s) minute error:%s\n", offset, err.Error())
		return 0
	}
	os := (h * 60 + m) * 60
	if strings.HasPrefix(offset, "-") {
		return int(os * -1)
	}
	return int(os)
}

func TimeInZone(t time.Time, zoneName, offset string) (zt time.Time) {
	if !offsetValid(offset) {
		return t
	}

	return t.In(time.FixedZone(zoneName, zoneOffset(offset)))
}

func TimeLocation(zoneName, offset string) *time.Location {
	if !offsetValid(offset) {
		return time.UTC
	}
	return time.FixedZone(zoneName, zoneOffset(offset))
}

func IsZoneTimeBetween(t time.Time, offset, lhm, rhm string) bool {
	zt := TimeInZone(t, "", offset)
	if !hmValid(lhm) || !hmValid(rhm) {
		return false
	}

	zhm := zt.Format("15:04")
	return zhm >= lhm && zhm <= rhm
}
