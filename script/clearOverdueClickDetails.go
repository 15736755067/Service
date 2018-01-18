package main

import (
	"Service/config"
	"Service/db"
	"Service/log"
	"flag"
	"fmt"
	_ "runtime/debug"
	"strconv"
	"strings"
	"time"
)

//var deleteClickDetailSql = `DELETE FROM Cache.ReqCache WHERE unix_timestamp(createdAt) <= unix_timestamp(date_sub(NOW(),interval 1 month))`

/*
usage:
go run clearClickDetails2.go -config "/home/albert/project/newbidder/AdClickService/src/Service/config/production2.ini" -beginDate "2017-03-01" -endDate "2017-04-01"

*/

var queryLimit = 100
var getOldClickSql = `SELECT clickId FROM Cache.ReqCache WHERE createdAt <= date_sub(NOW(),interval 2 week) ORDER BY createdAt asc LIMIT ` + strconv.Itoa(queryLimit)
var deleteClickSql = `DELETE FROM Cache.ReqCache WHERE clickId in (?)`

func makeDeleteClickSqlTemplate(length int) string {
	return "DELETE FROM Cache.ReqCache WHERE clickId in (?" + strings.Repeat(",?", length-1) + ")"
}

func clearOutdatedClickDetails() {

	var isContinue bool
	flag.BoolVar(&isContinue, "isContinue", false, "whether process continue when no row meet criter") // for limited period use false; from one point to future use true

	var beginDate, endDate string
	flag.StringVar(&beginDate, "beginDate", "", "empty or 2017-03-01")
	flag.StringVar(&endDate, "endDate", "", "empty or 2017-04-30")
	flag.Parse()

	timeTemp := " STR_TO_DATE('%v 01:00am', '%%Y-%%m-%%d %%h:%%i%%p') "
	getOldSql1 := "SELECT clickId FROM Cache.ReqCache WHERE createdAt <= date_sub(NOW(),interval 2 week) "

	var selectSlice []string
	selectSlice = append(selectSlice, getOldSql1)
	if beginDate != "" {
		selectSlice = append(selectSlice, " and createdAt >= "+fmt.Sprintf(timeTemp, beginDate))
	}
	if endDate != "" {
		selectSlice = append(selectSlice, " and createdAt <= "+fmt.Sprintf(timeTemp, endDate))
	}

	selectSlice = append(selectSlice, " ORDER BY createdAt asc LIMIT "+strconv.Itoa(queryLimit))

	getOldClickSql = strings.Join(selectSlice, "")

	fmt.Println("getOldClickSql", getOldClickSql)

	//fmt.Println("---", time.Now())
	/*
		defer func() {
			if x := recover(); x != nil {
				fmt.Println(x)
				log.Alert(x, string(debug.Stack()))
			}
		}()
	*/

	if err := config.LoadConfig(true); err != nil {
		panic(err.Error())
	}

	logAdapter := config.String("LOG", "adapter")
	logConfig := config.String("LOG", "jsonconfig")
	logAsync := config.Bool("LOG", "async")
	if logAdapter == "" {
		logAdapter = "console"
	}
	if logConfig == "" {
		logConfig = `{"level":7}`
	}
	log.Init(logAdapter, logConfig, logAsync)
	defer func() {
		log.Flush()
	}()

	dbClient := db.GetDB("REMOTEREQCACHE")

	queryStmt, err := dbClient.Prepare(getOldClickSql)
	if err != nil {
		fmt.Println(err)
	}

	deleteStmt, err := dbClient.Prepare(makeDeleteClickSqlTemplate(queryLimit))

	rowCount := 0
	for {
		fmt.Println("111")
		for {
			fmt.Println("222")
			rowCount = 0
			//rows, err := dbClient.Query(getOldClickSql)
			rows, err := queryStmt.Query()
			if err != nil {
				fmt.Println(err)
				time.Sleep(1 * time.Second)
				continue
			}

			var clickIdArray []interface{}
			fmt.Println("333")

			for rows.Next() {
				fmt.Println("---", rowCount)
				var clickId string
				if err := rows.Scan(&clickId); err != nil {
					fmt.Println(err)
					//log.Errorf(err)
					continue
				}
				rowCount += 1
				clickIdArray = append(clickIdArray, clickId)
			}

			fmt.Println("---", rowCount)
			fmt.Println("---", clickIdArray)

			if rowCount == 0 {
				if isContinue == false {
					return
				}
				break
			}

			//idArrayStr := strings.Join(clickIdArray, ",")

			//---
			//fmt.Println("-----", rowCount, idArrayStr)
			//fmt.Println(idArrayStr)

			//result, err := dbClient.Exec(deleteClickSql, idArrayStr)
			if rowCount == queryLimit {
				result, err := deleteStmt.Exec(clickIdArray...)
				if err != nil {
					fmt.Println("-----err", err)
				}
				fmt.Println("-----result", result)
			} else {
				deleteStmt2, err := dbClient.Prepare(makeDeleteClickSqlTemplate(rowCount))
				fmt.Println("---err", err)
				result, err := deleteStmt2.Exec(clickIdArray...)
				if err != nil {
					fmt.Println("-----err", err)
				}
				fmt.Println("-----result", result)
			}
		}
		time.Sleep(60 * time.Second)
	}
}

func main() {
	clearOutdatedClickDetails()
}
