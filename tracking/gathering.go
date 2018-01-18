package tracking

import "time"

type adStatisValues struct {
	Visits      int
	Clicks      int
	Conversions int
	Cost        int64
	Revenue     int64
	Impressions int
}

// adStaticTableFields AdStatic表所有的字段
type adStaticTableFields struct {
	AdStatisKey
	adStatisValues
}

type events struct {
	keyMd5    string
	keyFields AdStatisKey
	action    func(d *adStatisValues)
}

var userStatis map[string]*adStaticTableFields
var toSave chan map[string]*adStaticTableFields
var gatherChan chan events
var flushEvent chan struct{}

func init() {
	gatherChan = make(chan events, 1000000)
	userStatis = make(map[string]*adStaticTableFields)
	toSave = make(chan map[string]*adStaticTableFields)
	flushEvent = make(chan struct{})
}

func getData(keyMD5 string, keyFields AdStatisKey) *adStaticTableFields {
	d, ok := userStatis[keyMD5]
	if !ok {
		d = &adStaticTableFields{}
		d.AdStatisKey = keyFields
		userStatis[keyMD5] = d
	}
	return d
}

// 此函数只能在下面的gathering里面调用
func flush() {
	toSave <- userStatis
	userStatis = make(map[string]*adStaticTableFields)
}

// Gathering 统计信息汇总
func Gathering(stop chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case a := <-gatherChan:
			d := getData(a.keyMd5, a.keyFields)
			a.action(&d.adStatisValues)

		case <-ticker.C:
			flush()

		case <-flushEvent:
			flush()

		case <-stop:
			// 把已经有的收完
			for {
				select {
				case a := <-gatherChan:
					d := getData(a.keyMd5, a.keyFields)
					a.action(&d.adStatisValues)
				default:
					// 没有多余的数据了
					goto allreceived
				}
			}
		allreceived:
			flush()
			return
		}
	}
}
