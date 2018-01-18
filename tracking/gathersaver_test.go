package tracking

import (
	"Service/gracequit"
	"Service/tracking/saver"
	"fmt"
	"testing"
	"time"
)

// 测试ip.go功能是否正常

type tableTest struct {
	key       interface{}
	insertSQL string
	deleteSQL string
	selectSQL string
	user      int
	table     string
}

// 每一个表都测试一下
var tableTests = []tableTest{
	{
		user: 1,
		key: IPStatisKey{
			UserID:      1,
			Timestamp:   14000000000,
			CampaignID:  2,
			IP:          "123.123.123.123",
			CountryName: "China",
		},
		deleteSQL: "DELETE FROM AdIPStatis WHERE UserID=?",
		selectSQL: `SELECT Visits, Clicks, Conversions, Cost, Revenue, Impressions, Clicks FROM AdIPStatis WHERE UserID=? and Timestamp=? and CampaignID=? and IP=? and CountryName=?`,
		insertSQL: ipStatisSQL,
		table:     "AdIPStatis",
	},
	{
		user: 2,
		key: ReferrerStatisKey{
			UserID:     2,
			Timestamp:  15000000000,
			CampaignID: 3,
			Referrer:   "http://www.baidu.com/",
		},
		deleteSQL: "DELETE FROM AdReferrerStatis WHERE UserID=?",
		selectSQL: `SELECT Visits, Clicks, Conversions, Cost, Revenue, Impressions, Clicks FROM AdReferrerStatis WHERE UserID=? and Timestamp=? and CampaignID=? and Referrer=?`,
		insertSQL: referrerStatisSQL,
		table:     "AdReferrerStatis",
	},
	{
		user: 3,
		key: ReferrerDomainStatisKey{
			UserID:         3,
			Timestamp:      15000000000,
			CampaignID:     3,
			ReferrerDomain: "http://www.baidu.com/",
		},
		deleteSQL: "DELETE FROM AdReferrerDomainStatis WHERE UserID=?",
		selectSQL: `SELECT Visits, Clicks, Conversions, Cost, Revenue, Impressions, Clicks FROM AdReferrerDomainStatis WHERE UserID=? and Timestamp=? and CampaignID=? and ReferrerDomain=?`,
		insertSQL: referrerDomainStatisSQL,
		table:     "AdReferrerDomainStatis",
	},
}

func TestAllStatis(t *testing.T) {
	for _, table := range tableTests {
		testTable(t, table)
	}
}

func testTable(t *testing.T, table tableTest) {
	var g gracequit.GraceQuit

	ip := newGatherSaver(&g, table.insertSQL, 10*time.Second)
	ip.Start(db)

	// 测试之前先清空此用户的数据
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE UserID=?", table.table), table.user)
	if err != nil {
		panic(err)
	}

	// 汇总值
	ip.AddVisit(table.key, 1)
	ip.AddConversion(table.key, 2)
	ip.AddCost(table.key, 3)
	ip.AddRevenue(table.key, 4)
	ip.AddImpression(table.key, 5)
	ip.AddClick(table.key, 6)

	// 停止
	g.StopAll()

	// 检查值
	row := db.QueryRow(table.selectSQL, saver.Args{}.AddFlatValues(table.key)...)

	v := StatisValue{}
	err = row.Scan(&v.Visits, &v.Clicks, &v.Conversions, &v.Cost, &v.Revenue, &v.Impressions, &v.Clicks)
	if err != nil {
		panic(err)
	}
	if v.Visits != 1 || v.Conversions != 2 || v.Cost != 3*MILLION || v.Revenue != 4*MILLION || v.Impressions != 5 || v.Clicks != 6 {
		t.Errorf("expected:%+v actual:%+v", "1,2,3,4,5,6", v)
	}

	t.Logf("table %+v ok", table)
}
