package tracking

import (
	"Service/log"
	"database/sql"
)

// NewCampMap 数据库中插入一条新的campaign map
func InsertCampMap(ourCamp int64, theirCamp string) {
	timestamp := Timestamp()
	select {
	case campMapBuffer <- campMap{
		ourCamp:   ourCamp,
		timestamp: timestamp,
		theirCamp: theirCamp,
	}:
	default:
		return
	}
}

type campMap struct {
	ourCamp   int64
	timestamp int64
	theirCamp string
}

var campMapBuffer = make(chan campMap, 1024)

// CampMapSaving saving campMapBuffer
func CampMapSaving(db *sql.DB, stop chan struct{}) {
	for {
		select {
		case m := <-campMapBuffer:
			err := saveCampMap(db, m)
			if err != nil {
				log.Errorf("saveCampMap m:%+v failed:%v", m, err)
			}
		case <-stop:
			return
		}
	}
}

func saveCampMap(db *sql.DB, m campMap) error {
	sql := `INSERT IGNORE INTO CampaignMap values (?, ?, ?)`
	_, err := db.Exec(sql, m.ourCamp, m.timestamp, m.theirCamp)
	return err
}
