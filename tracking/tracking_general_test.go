package tracking

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"Service/gracequit"
)

var db *sql.DB

func init() {
	// 初始化数据库
	user := "root"
	pass := "123456"
	host := "localhost"
	port := 3306
	dbname := "AdClickTool"

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, pass, host, port, dbname)
	var err error
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}

}

func TestConversions(t *testing.T) {
	// 启动Conversion保存
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		SavingConversions(db, c)
	})
	start := time.Now()

	//
	count := 100000
	for i := 0; i < count; i++ {
		SaveConversion(randConversion())
	}

	step1 := time.Now()
	gracequit.StopAll()
	step2 := time.Now()

	t.Logf("call SaveConversion %v times take:%v", count, step1.Sub(start))
	t.Logf("wait all sql write done take:%v", step2.Sub(step1))
}

func TestTracking(t *testing.T) {
	// 启动保存
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		Saving(db, c)
	})

	// 启动汇总
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		Gathering(c, 10*time.Second)
	})

	// 调用一些API
	apis1 := []func(key AdStatisKey, count int){
		AddVisit,
		AddClick,
		AddConversion,
		AddImpression,
	}

	apis2 := []func(key AdStatisKey, count float64){
		AddCost,
		AddPayout,
	}

	begin := time.Now()
	count := 100000
	for i := 0; i < count; i++ {
		key := randStatisKey()

		f1 := apis1[rand.Int()%len(apis1)]
		f2 := apis2[rand.Int()%len(apis2)]

		f1(key, 1)
		f2(key, 1.0)
	}
	end := time.Now()
	t.Logf("flush %v rows take:%v", count, end.Sub(begin))

	gracequit.StopAll()

	//if savedCount != count {
	//	t.Errorf("savedCount:%v expected:%v", savedCount, count)
	//}
	//t.Logf("savedCount:%v", savedCount)
}
