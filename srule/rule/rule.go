package rule

import (
	"Service/config"
	"Service/log"
	"Service/srule/mail"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"html/template"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AdstatisData struct {
	CampaignId     int64  `json:"_"`
	DimensionKey   string `json:"_"`
	DimensionValue string `json:"_"`
	CampaignName   string `json:"_"`

	Impressions int64   `json:"impressions"`
	Visits      int64   `json:"visits"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	Cost        float32 `json:"cost"`
	Revenue     float32 `json:"revenue"`

	Profit float32 `json:"profit"`
	Cpv    float32 `json:"cpv"`
	Ictr   float32 `json:"ictr"`
	Ctr    float32 `json:"ctr"`
	Cr     float32 `json:"cr"`
	Cv     float32 `json:"cv"`
	Roi    float32 `json:"roi"`
	Epv    float32 `json:"epv"`
	Epc    float32 `json:"epc"`
	Ap     float32 `json:"ap"`

	Cpc float32 `json:"cpc"`
	Cpm float32 `json:"cpm"`
	Cpa float32 `json:"cpa"`
}

type DimensionValueAdstatis map[string]AdstatisData         // key 为dimension value
type DimensionKeyAdstatis map[string]DimensionValueAdstatis // key 为dimension key
type CampaignAdstatis map[string]DimensionKeyAdstatis       // key 为campaign id

type User struct {
	Id int64
}

type Rule struct {
	ID         int64
	UserId     string
	Name       string
	CampaignID string
	Dimension  string
	TimeSpan   string
	Condition  string
	Email      string
	Timezone   string

	Location   *time.Location
	Conditions []*RuleCondition
}

type RuleLog struct {
	Id          int64  `gorm:"primary_key"`
	ruleId      int64  `gorm:"column:ruleId"`
	ruleName    string `gorm:"column:ruleName"`
	campaignIDs string `gorm:"column:campaignID"`
	dimension   string `gorm:"column:dimension"`
	hit         int
	condition   string
	email       string
	sendStatus  int
	timeStamp   int64
}

type LetterData struct {
	Rule  *Rule
	Datas CampaignAdstatis
}

type RuleCondition struct {
	Expression string
	Operation  string
	Value      string
}

var mux sync.Mutex

func ProcessRuleBySchedule(schedule string) {
	rules := GetRulesBySchedule(schedule)
	for _, r := range rules {
		log.Infof("[ProcessRuleBySchedule]rule %+v", *r)
		r.ExecRule()
	}
}

func ProcessRuleByOneTime(oneTime string) {
	rules := GetRulesByOneTime(oneTime)
	for _, r := range rules {
		r.ExecRule()
	}
}

func parseLastTimeSpan(timespan string) time.Duration {
	var x int64
	reg, err := regexp.CompilePOSIX(`last([0-9]+)(seconds|minutes|hours|days|months|years)$`)
	if err != nil {
		fmt.Printf("parseTimeSpan err %v", err)
	}
	first := reg.FindStringSubmatch(timespan)

	if len(first) < 3 {
		return time.Duration(0)
	}
	if first[1] != "" {
		x1, _ := strconv.Atoi(first[1])
		x = int64(x1)
	}

	switch first[2] {
	case "seconds":
		return time.Duration(x) * time.Second
	case "minutes":
		return time.Duration(x) * time.Minute
	case "hours":
		return time.Duration(x) * time.Hour
	case "days":
		return time.Duration(x) * 24 * time.Hour
	case "months":
		return time.Duration(x) * 30 * 24 * time.Hour
	case "years":
		return time.Duration(x) * 365 * 24 * time.Hour
	default:
		return time.Duration(0)
	}
}

func lastTimeSpan(timespan string) int64 {
	now := time.Now().UTC().UnixNano()
	lastDur := parseLastTimeSpan(timespan)

	return (now - int64(lastDur)) / 1000000
}

func (rule *Rule) ParseCondition() {
	conditions := strings.Split(rule.Condition, ",")
	ruleConditions := make([]*RuleCondition, 0)
	for _, condition := range conditions {
		reg, err := regexp.CompilePOSIX(`([a-zA-Z]+)(>=|>|<=|<|=)(.+)`)
		if err != nil {
			log.Errorf("parse Condition err %v", err)
		}
		first := reg.FindStringSubmatch(condition)
		log.Infof("ParseCondition %+v", first)

		if len(first) < 4 {
			return
		}

		rc := &RuleCondition{
			Expression: first[1],
			Operation:  first[2],
			Value:      first[3],
		}
		ruleConditions = append(ruleConditions, rc)
	}
	rule.Conditions = ruleConditions
}

func (rule *Rule) ExecRule() bool {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("[ExecRule]rule %v panic recover! e: %v", rule.ID, e)
			debug.PrintStack()
		}
	}()
	if rule.CampaignID == "" {
		log.Errorf("[ExecRule]rule %v campaign id is empty", rule.ID)
		return false
	}
	db := getColumnStoreDB()
	//rule.ParseCondition()
	logDatas := make(CampaignAdstatis, 0)

	var selectStmt, whereStmt, groupStmt, havingStmt string
	var selectStmts, havingStmts, whereStmts []string
	//selectStmt
	selectStmts = []string{
		"sum(Impressions) AS sumImpressions",
		"sum(Visits) AS sumVisits",
		"sum(Clicks) AS sumClicks",
		"sum(Conversions) AS sumConversions",
		//"sum(Cost) AS sumCost",
		//"sum(Revenue) AS sumRevenue",
		"IFNULL(round(sum(Cost/1000000),2),0) as spent",
		"IFNULL(round(sum(Revenue/1000000),2),0) as revenue",
		"IFNULL(round(sum(Revenue / 1000000 - Cost / 1000000),2),0) as profit",
		"IFNULL(round(sum(Cost / 1000000) / sum(Visits),4),0) as cpv",
		"IFNULL(round(sum(Visits)/sum(Impressions)*100,2),0)  as  ictr",
		"IFNULL(round(sum(Clicks)/sum(Visits)*100,2),0) as ctr",
		"IFNULL(round(sum(Conversions)/sum(Clicks)*100,4),0) as  cr",
		"IFNULL(round(sum(Conversions)/sum(Visits)*100,2),0) as cv",
		"IFNULL(round((sum(Revenue) - sum(Cost))/sum(Cost)*100,2),0) as roi",
		"IFNULL(round(sum(Revenue)/ 1000000 / sum(Visits),4),0) as epv",
		"IFNULL(round(sum(Revenue)/ 1000000 / sum(Clicks),4),0) as epc",
		"IFNULL(round(sum(Revenue)/ 1000000 / sum(Conversions),2),0) as ap",
		"IFNULL(round(sum(Cost / 1000000) / sum(Clicks),4),0) as cpc",
		"IFNULL(round(sum(Cost / 1000000) / sum(Impressions)*1000,4),0) as cpm",
		"IFNULL(round(sum(Cost / 1000000) / sum(Conversions),4),0) as cpa",
	}
	selectStmt = strings.Join(selectStmts, ", ")

	//havingStmt
	havingStmts = strings.Split(rule.Condition, ",")
	havingStmt = strings.Join(havingStmts, " or ")

	//whereStmt
	whereStmts = append(whereStmts, fmt.Sprintf("CampaignID in (%v)", rule.CampaignID))

	if strings.HasPrefix(rule.TimeSpan, "last") {
		//micro second
		whereStmts = append(whereStmts, fmt.Sprintf("Timestamp > %d", lastTimeSpan(rule.TimeSpan)))
	} else if rule.TimeSpan == "previousDay" {
		todayStart := time.Now().In(rule.Location).Truncate(24 * time.Hour)
		previousDayStart := todayStart.Add(-24 * time.Hour).Unix()
		previousDayEnd := todayStart.Unix()
		whereStmts = append(whereStmts, fmt.Sprintf("Timestamp/1000 between %d and %d", previousDayStart, previousDayEnd))
	} else if rule.TimeSpan == "sameDay" {
		sameDayStart := time.Now().In(rule.Location).Truncate(24 * time.Hour).Unix()
		whereStmts = append(whereStmts, fmt.Sprintf("Timestamp/1000 > %d", sameDayStart))
	}

	whereStmt = strings.Join(whereStmts, " and ")

	groupStmt = "campaignID"
	if rule.Dimension != "" {
		groupStmt = groupStmt + "," + rule.Dimension
	}

	orderStmt := "campaignID desc"
	selectStmt = groupStmt + "," + selectStmt
	rows, err := db.Table("adstatis").
		Select(selectStmt).
		Where(whereStmt).
		Group(groupStmt).
		Having(havingStmt).
		Order(orderStmt).
		Limit(50).
		Rows()

	if err != nil {
		log.Infof("[ExecRule]DB query err: %v,rule id %v", err, rule.ID)
		return false
	}

	defer rows.Close()

	//ruleLog
	ruleLog := RuleLog{
		ruleId:      rule.ID,
		ruleName:    rule.Name,
		campaignIDs: rule.CampaignID,
		dimension:   rule.Dimension,
		condition:   rule.Condition,
		email:       rule.Email,
		sendStatus:  0,
		timeStamp:   time.Now().Unix(),
	}

	for rows.Next() {
		var data AdstatisData
		var group sql.NullString
		err := rows.Scan(&data.CampaignId,
			&group,
			&data.Impressions,
			&data.Visits,
			&data.Clicks,
			&data.Conversions,
			&data.Cost,
			&data.Revenue,
			&data.Profit,
			&data.Cpv,
			&data.Ictr,
			&data.Ctr,
			&data.Cr,
			&data.Cv,
			&data.Roi,
			&data.Epv,
			&data.Epc,
			&data.Ap,
			&data.Cpc,
			&data.Cpm,
			&data.Cpa,
		)
		if err != nil {
			log.Infof("[ExecRule] scan err: %v", err)
		}

		ruleLog.hit = 1
		DimensionValue := "UnKnown"
		if group.Valid {
			DimensionValue = group.String
		}
		data.DimensionKey = rule.Dimension
		data.DimensionValue = DimensionValue

		campaignId := fmt.Sprintf("%d", data.CampaignId)

		campaign := getCampaign(data.CampaignId)
		if campaign != nil {
			data.CampaignName = campaign.name
		}

		if _, ok := logDatas[campaignId]; !ok {
			logDatas[campaignId] = make(DimensionKeyAdstatis, 0)
		}

		dimensionKeyAdstatis := logDatas[campaignId]
		if _, ok := dimensionKeyAdstatis[data.DimensionKey]; !ok {
			dimensionKeyAdstatis[data.DimensionKey] = make(DimensionValueAdstatis, 0)
		}

		dimensionValueAdstatis := dimensionKeyAdstatis[data.DimensionKey]

		dimensionValueAdstatis[data.DimensionValue] = data

		//dimensionValueAdstatis = append(dimensionValueAdstatis, data)

		DimensionKey := rule.Dimension
		log.Debugf("[ExecRule]row: %v:%v,%+v\n", DimensionKey, DimensionValue, data)
	}

	// send mail
	if config.String("EMAIL", "enable") == "on" && ruleLog.hit == 1 {
		letterText := rule.makeLetter(logDatas)
		err = rule.sendMail(rule.Email, letterText)
		if err == nil {
			ruleLog.sendStatus = 1
		}
	}

	ruleLog.saveRuleLog(logDatas)

	return false
}

func (rule *Rule) makeLetter(logDatas CampaignAdstatis) string {

	templ := `
	<html>
        <head>
        <style type="text/css">
            .tab{border-top:1px solid #000;border-left:1px solid #000;text-align:center}
            .tab th{border-bottom:1px solid #000;border-right:1px solid #000;}
            .tab td{border-bottom:1px solid #000;border-right:1px solid #000;}
        </style>
        </head>
        <body>
        <div>
        	<a href="http://panel.newbidder.com/">
        		<img src="http://panel.newbidder.com/assets/img/logo-three.3c2764cf9c492de5.png">
        	</a>
        </div>
	<div>
	<br />
	<b><font color="#0B610B">Rule Infomation</font></b>
        <hr size="2" width="100%" align="center" />
	{{with .Rule}}
		<ul>
                    <li>Rule: {{.Name}}</li>
                    <li>Dimension: {{.Dimension}} </li>
                    <li>Condition: {{.Condition}} </li>
                    <li>TimeSpan: {{.TimeSpan}}</li>
                </ul>
	{{end}}

	</div>
	<b><font color="#0B610B">Report Data</font></b>
	<hr size="2" width="100%" align="center" />
		<table class="tab" >
		  <tr>
		    <th>CampaignName</th>
		    <th>Dimension</th>
		    <th>DimensionValue</th>
		    <th>Impressions</th>
		    <th>Visits</th>
		    <th>Clicks</th>
		    <th>Conversions</th>
		    <th>Spent</th>
		    <th>Ctr</th>
		    <th>Cr</th>
		  </tr>
		  {{range $dimensionkeydata := .Datas}}
		{{range $dimensionValuedata := .}}
		{{range $data := .}}
		  <tr>
		    <td>{{.CampaignName}}</td>
		    <td>{{.DimensionKey}}</td>
		    <td>{{.DimensionValue}}</td>
		    <td>{{.Impressions}}</td>
		    <td>{{.Visits}}</td>
		    <td>{{.Clicks}}</td>
		    <td>{{.Conversions}}</td>
		    <td>{{.Cost}}</td>
		    <td>{{.Ctr}}</td>
		    <td>{{.Cr}}</td>
		  </tr>
		  {{end}}
		  {{end}}
		{{end}}
		</table>

        </body>
</html>
`
	t := template.New("alter letter")
	t, err := t.Parse(templ)
	if err != nil {
		log.Errorf("err: %v", err)
	}
	var w bytes.Buffer

	letterData := LetterData{Rule: rule, Datas: logDatas}
	err = t.Execute(&w, letterData)
	return w.String()
}

func (r *Rule) sendMail(mailTo, mailText string) error {
	if mailTo == "" {
		log.Infof("mailto is empty: rule id %v", r.ID)
		return nil
	}
	user := config.String("EMAIL", "user")
	password := config.String("EMAIL", "password")
	smtpHost := config.String("EMAIL", "smtp_host")
	smtpPort := config.Int("EMAIL", "smtp_port")

	host := strings.TrimSpace(smtpHost) + ":" + strconv.Itoa(smtpPort)
	to := mailTo

	subject := "Newbidder Tracking AutoRule"

	header := make(map[string]string)
	header["From"] = "Dockboard " + "<" + user + ">"
	header["To"] = to
	header["Subject"] = subject
	header["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + mailText

	err := mail.SendMail(user, password, host, to, subject, mailText, "html")
	if err != nil {
		log.Errorf("[sendMail]send mail error %v!\n", err)
		return err
	} else {
		log.Infof("[sendMail]send mail success: to %v,rule id %v", to, r.ID)
	}
	return nil
}

func (l RuleLog) saveRuleLog(logDatas CampaignAdstatis) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("[saveRuleLog]panic recover! e: %v", e)
			debug.PrintStack()
		}
	}()
	db := getMysqlDB()
	//log.Debugf("SuddenChangeLog %v", db.Table("SuddenChangeLog").NewRecord(l))
	//db.Exec("INSERT INTO SuddenChangeLog(`ruleId`,`dimension`,`condition`,`notifiedEmails`,`sendStatus`,`timeStamp`) VALUES(?,?,?,?,?,?)", l.ruleId, l.dimension, l.condition, l.email, l.sendStatus, l.timeStamp)
	mux.Lock()
	defer mux.Unlock()
	res, err := db.DB().Exec("INSERT INTO SuddenChangeLog(`ruleId`,`dimension`,`hit`,`condition`,`notifiedEmails`,`sendStatus`,`timeStamp`) VALUES(?,?,?,?,?,?,?)", l.ruleId, l.dimension, l.hit, l.condition, l.email, l.sendStatus, l.timeStamp)
	if err != nil {
		log.Errorf("Cannot run insert statement", err)
	}

	logId, _ := res.LastInsertId()
	log.Infof("[SuddenChangeLog]Save rule log success: logId %v", logId)
	l.saveRuleLogDetail(logId, logDatas)
}

func (l RuleLog) saveRuleLogDetail(logId int64, logDatas CampaignAdstatis) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("[saveRuleLogDetail]panic recover! e: %v", e)
			debug.PrintStack()
		}
	}()
	db := getMysqlDB()

	//log.Debugf("SuddenChangeLogDetail %v", db.Table("SuddenChangeLogDetail").NewRecord(l))
	for campaigId, campaignData := range logDatas {
		for dimensionKey, DimensionKeyData := range campaignData {
			for dimensionValue, DimensionValueData := range DimensionKeyData {
				bs, err := json.Marshal(DimensionValueData)
				if err != nil {
					continue
				}
				db.Exec("INSERT INTO SuddenChangeLogDetail(`logId`,`campaignID`,`dimensionKey`,`dimensionValue`,`data`) VALUES(?,?,?,?,?)", logId, campaigId, dimensionKey, dimensionValue, string(bs))
			}
		}

		log.Infof("[SuddenChangeLogDetail]Save rule log detail success: logId %v", logId)
	}
}
