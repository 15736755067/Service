package main

import (
	"Service/db"
	"Service/log"
	"runtime/debug"
)

var deleteClickDetailSql = `DELETE FROM Cache.ReqCache WHERE unix_timestamp(createdAt) <= unix_timestamp(date_sub(NOW(),interval 1 month))`

func clearOutdatedClickDetails() {
	defer func() {
		if x := recover(); x != nil {
			log.Alert(x, string(debug.Stack()))
		}
	}()

	result, err := db.GetDB("REMOTEREQCACHE").Exec(deleteClickDetailSql)
	if err != nil {
		log.Errorf("[clearOutdatedClickDetails]Exec[%s] failed:%s", deleteClickDetailSql, err.Error())
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		log.Errorf("[clearOutdatedClickDetails] RowsAffected failed:%s", err.Error())
		return
	}
	log.Infof("[clearOutdatedClickDetails]%d clicks are deleted\n", affected)
}
