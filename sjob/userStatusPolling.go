// 这个job实现的是，定期去检查、更新用户的Status，然后发出变化通知
package main

import (
	"Service/db"
	"Service/log"
	"fmt"
	"runtime/debug"
	"strings"
)

var publishChannel = "channel_campaign_changed_users"

// overdue check
var newOverdueUserSql = `SELECT User.id FROM AdClickTool.UserBilling,AdClickTool.User 
		WHERE UserBilling.expired = 0
		AND UserBilling.userId = User.id
		AND User.status = 1
		AND UserBilling.planEnd>0 
		AND UserBilling.planEnd<=unix_timestamp(current_timestamp())`
var updateUserStatuOverdueSql = `UPDATE User, UserBilling SET User.status=2 
		WHERE UserBilling.expired = 0
		AND UserBilling.userId = User.id
		AND User.status = 1
		AND UserBilling.planEnd>0 
		AND UserBilling.planEnd<=unix_timestamp(current_timestamp())
		AND UserBilling.userId IN (`

// overage check(已过期的user，依旧可以透支，对于这部分User要做overage check)
var newOveragedUserSql = `SELECT User.id FROM AdClickTool.UserBilling,AdClickTool.User 
		WHERE UserBilling.expired = 0
		AND UserBilling.userId = User.id
		AND ( User.status = 1 OR User.status = 2 )
		AND UserBilling.totalEvents >= 2*(UserBilling.includedEvents+UserBilling.boughtEvents+UserBilling.freeEvents+UserBilling.overageLimit)`
var updateUserStatusOverageSql = `UPDATE User, UserBilling SET User.status=3 
		WHERE UserBilling.expired = 0
		AND UserBilling.userId = User.id
		AND ( User.status = 1 OR User.status = 2 )
		AND UserBilling.totalEvents >= 2*(UserBilling.includedEvents+UserBilling.boughtEvents+UserBilling.freeEvents+UserBilling.overageLimit)
		AND UserBilling.userId IN (`

// update user status to overdue(2)
func userStatusOverdueUpdate() {
	defer func() {
		if x := recover(); x != nil {
			log.Alert(x, string(debug.Stack()))
		}
	}()
	rows, err := db.GetDB("DB").Query(newOverdueUserSql)
	if err != nil {
		log.Errorf("[userStatusOverdueUpdate]newOverdueUserSql query failed with error(%s)\n", err.Error())
		return
	}
	defer rows.Close()
	candidates := make([]int, 0)
	userId := 0
	for rows.Next() {
		err = rows.Scan(&userId)
		if err != nil {
			log.Errorf("[userStatusOverdueUpdate]newOveragedUserSql rows.Scan failed with error(%s)\n", err.Error())
			return
		}
		candidates = append(candidates, userId)
	}
	log.Infof("[userStatusOverdueUpdate]%d users' status should be changed to 2\n", len(candidates))

	idStr := userIdStr(candidates)
	if idStr == "" {
		return
	}
	log.Info("[userStatusOverdueUpdate] <", idStr, ">")
	result, err := db.GetDB("DB").Exec(updateUserStatuOverdueSql + idStr + `)`)
	if err != nil {
		log.Errorf("[userStatusOverdueUpdate]Exec[%s] failed:%s", updateUserStatuOverdueSql, err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Errorf("[userStatusOverdueUpdate]RowsAffected failed:%s", err.Error())
		return
	}
	log.Infof("[userStatusOverdueUpdate]%d users' status changed to 2\n", affected)

	for _, id := range candidates {
		// notify tracking service
		db.GetRedisClient("MSGQUEUE").Publish(publishChannel, fmt.Sprintf("0.update.user.%d", id))
	}
}

// update user status to overage(3)
func userStatusOverageUpdate() {
	defer func() {
		if x := recover(); x != nil {
			log.Alert(x, string(debug.Stack()))
		}
	}()
	rows, err := db.GetDB("DB").Query(newOveragedUserSql)
	if err != nil {
		log.Errorf("[userStatusOverageUpdate]newOveragedUserSql query failed with error(%s)\n", err.Error())
		return
	}
	defer rows.Close()
	candidates := make([]int, 0)
	userId := 0
	for rows.Next() {
		err = rows.Scan(&userId)
		if err != nil {
			log.Errorf("[userStatusOverageUpdate]newOveragedUserSql rows.Scan failed with error(%s)\n", err.Error())
			return
		}
		candidates = append(candidates, userId)
	}
	log.Infof("[userStatusOverageUpdate]%d users' status should be changed to 3\n", len(candidates))

	idStr := userIdStr(candidates)
	if idStr == "" {
		return
	}
	log.Info("[userStatusOverageUpdate] <", idStr, ">")
	result, err := db.GetDB("DB").Exec(updateUserStatusOverageSql + idStr + `)`)
	if err != nil {
		log.Errorf("[userStatusOverageUpdate]Exec[%s] failed:%s", updateUserStatusOverageSql, err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Errorf("[userStatusOverageUpdate]RowsAffected failed:%s", err.Error())
		return
	}
	log.Infof("[userStatusOverageUpdate]%d users' status changed to 3\n", affected)

	for _, id := range candidates {
		// notify tracking service
		db.GetRedisClient("MSGQUEUE").Publish(publishChannel, fmt.Sprintf("0.update.user.%d", id))
	}
}

func userIdStr(ids []int) (s string) {
	if len(ids) == 0 {
		return ""
	}
	for _, id := range ids {
		s += fmt.Sprintf("%d,", id)
	}
	return strings.TrimSuffix(s, ",")
}
