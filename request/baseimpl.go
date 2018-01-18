package request

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"

	"Service/common"
	"Service/log"
	"Service/tracking"
	"Service/util/countrycode"
	"Service/util/ip"
	"Service/util/ip2location"
	"Service/util/useragent"

	"strconv"

	"golang.org/x/text/language"
)

type reqbase struct {
	id string
	t  string
	ip string
	ua string

	externalId string   // 从traffic source里面的ExternalId字段传过来的externalId
	cost       float64  // 从traffic source里面的Cost字段传过来的cost
	tsCId      string   // 从traffic source里面的CampaignId字段传过来的campaignId
	websiteId  string   // 从traffic source里面的WebsiteId字段传过来的websiteId
	vars       []string // 从traffic source里面的Var1-10字段传过来的vars
	payout     float64
	txid       string

	trafficSourceId   int64
	trafficSourceName string

	userId            int64
	userIdText        string

	campaignHash      string
	campaignId        int64
	campaignName      string
	campaignCountry   string

	flowId            int64
	flowName          string

	ruleId            int64
	pathId            int64
	landerId          int64
	landerName        string
	offerId           int64
	optOfferId        int64
	offerName         string

	affiliateId       int64
	optAffiliateId    int64
	affiliateName     string

	impTimeStamp      int64
	visitTimeStamp    int64
	clickTimeStamp    int64
	postbackTimeStamp int64

	deviceType     string
	trackingDomain string
	trackingPath   string

	referrer       string
	referrerdomain string

	language       string
	model          string
	brand          string

	countryCode    string
	countryName    string
	region         string
	city           string

	carrier        string
	isp            string
	os             string
	osVersion      string
	browser        string
	browserVersion string
	connectionType string
	bot            bool

	cookie map[string]string

	urlParam map[string]string

	// Traffic Source的一些配置
	//	tsExternalId *common.TrafficSourceParams // 暂时不需要，去掉
	//	tsCost       *common.TrafficSourceParams // 暂时不需要，去掉
	tsVars   []common.TrafficSourceParams
	cpaValue float64

	redirectMode int64

	//tsCampaignId string
	//tsWebsiteId  string
}

//func (r *reqbase) SetTSExternalID(e *common.TrafficSourceParams) {
//	r.tsExternalId = e
//}

//func (r *reqbase) SetTSCost(c *common.TrafficSourceParams) {
//	r.tsCost = c
//}

func (r *reqbase) SetTSVars(vars []common.TrafficSourceParams) {
	r.tsVars = make([]common.TrafficSourceParams, len(vars))
	for i := range vars {
		r.tsVars[i] = vars[i]
	}
}

func (r *reqbase) SetCPAValue(v float64) {
	r.cpaValue = v
}

//func (r *reqbase) TSExternalID() *common.TrafficSourceParams {
//	return r.tsExternalId
//}

//func (r *reqbase) TSCost() *common.TrafficSourceParams {
//	return r.tsCost
//}

func (r *reqbase) TSVars() []common.TrafficSourceParams {
	t := make([]common.TrafficSourceParams, len(r.tsVars))
	for i := range r.tsVars {
		t[i] = r.tsVars[i]
	}
	return t
}

func (r *reqbase) CPAValue() float64 {
	return r.cpaValue
}

func newReqBase(id, t string, r *http.Request) (req *reqbase) {
	req = &reqbase{
		id:             id,
		t:              t,
		ip:             ip.GetIP(r),
		ua:             r.UserAgent(),
		trackingPath:   r.URL.Path,
		trackingDomain: strings.Split(r.Host, ":")[0],
		vars:           make([]string, VarsMaxNum),
		cookie:         make(map[string]string),
		urlParam:       make(map[string]string),
	}

	// parse traffic source externalId/cost/vars
	switch t {
	case ReqLPOffer:
		fallthrough
	case ReqImpression:
		// 从url参数中获取，由外部调用ParseTSParams处理
	case ReqLPClick:
		// 从user cookie中获取，由外部调用ParseTSParams处理
	case ReqS2SPostback:
		// 从Cache中获取，实际上不会运行到这里，由S2SPostbackRequest生成时处理掉了
		panic("newReqBase actually does not run for ReqS2SPostback")
	}

	//TODO connectiontype未采集到
	// chuck说connectiontype实际上用得少，可以先不做

	// parse location from ip
	location := ip2location.Get_all(req.ip)
	req.countryCode = countrycode.CountryCode2To3(location.Country_short) // alpha-2 => alpha-3
	req.countryName = location.Country_long
	req.region = location.Region
	req.city = location.City
	req.carrier = location.Mobilebrand
	req.isp = location.Isp
	req.connectionType = location.Netspeed

	// parse mobile info from ua
	ua := useragent.New(r.UserAgent())
	log.Debug("User-Agent:", ua.UA())
	req.os = ua.OSInfo().Name
	req.osVersion = ua.OSInfo().Version
	req.browser, req.browserVersion = ua.Browser()
	req.bot = ua.Bot()
	if ua.Mobile() {
		req.brand = ua.MobileBrand()
		req.model = ua.MobileModel()
		req.deviceType = ua.DeviceType()
	} else {
		req.brand = "Desktop"
		req.model = "Desktop"
		req.deviceType = "Desktop"
	}

	req.referrer = r.Referer()
	if rurl, err := url.Parse(req.referrer); err == nil && rurl != nil {
		req.referrerdomain = rurl.Host
	}
	tag, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil || len(tag) == 0 {
		req.language = "UNKNOWN"
	} else {
		req.language = tag[0].String() // 只用取quality最高的language就可以了
	}

	return req
}

func (r *reqbase) SetId(id string) {
	r.id = id
}

func (r *reqbase) Id() string {
	return r.id
}
func (r *reqbase) Type() string {
	return r.t
}
func (r *reqbase) String() string {
	return fmt.Sprintf("[Request][%s][%s][%d][%d]", r.Type(), r.Id(), r.UserId(), r.CampaignId())
}
func (r *reqbase) TrackingPath() string {
	return r.trackingPath
}

func (r *reqbase) ExternalId() string {
	return r.externalId
}

func (r *reqbase) Cost() float64 {
	return r.cost
}

func (r *reqbase) Vars(n uint) string { // n:0~VarsMaxNum-1
	if n < 0 || n >= VarsMaxNum {
		return ""
	}
	return r.vars[n]
}

func (r *reqbase) ParseTSParams(
	externalId common.TrafficSourceParams,
	cost common.TrafficSourceParams,
	params []common.TrafficSourceParams,
	tsCampaign common.TrafficSourceParams,
	tsWebsite common.TrafficSourceParams,
	values QueryHolder) {
	if r == nil {
		return
	}
	r.externalId = values.Get(externalId.Parameter)

	if len(cost.Parameter) != 0 {
		costStr := values.Get(cost.Parameter)
		var err error
		r.cost, err = strconv.ParseFloat(costStr, 64)
		if err != nil {
			log.Errorf("Parse Cost from:%v failed:%v", costStr, err)
		}
	}

	if len(tsCampaign.Parameter) != 0 {
		r.tsCId = values.Get(tsCampaign.Parameter)
	}

	if len(tsWebsite.Parameter) != 0 {
		r.websiteId = values.Get(tsWebsite.Parameter)
	}

	for i, p := range params {
		if i >= VarsMaxNum {
			break
		}
		r.vars[i] = values.Get(p.Parameter)
	}
}

func (r *reqbase) TrafficSourceId() int64 {
	return r.trafficSourceId
}
func (r *reqbase) SetTrafficSourceId(id int64) {
	r.trafficSourceId = id
}
func (r *reqbase) TrafficSourceName() string {
	return r.trafficSourceName
}
func (r *reqbase) SetTrafficSourceName(name string) {
	r.trafficSourceName = name
}
func (r *reqbase) UserId() int64 {
	return r.userId
}
func (r *reqbase) SetUserId(id int64) {
	r.userId = id
}
func (r *reqbase) UserIdText() string {
	return r.userIdText
}
func (r *reqbase) SetUserIdText(idText string) {
	r.userIdText = idText
}
func (r *reqbase) CampaignHash() string {
	return r.campaignHash
}
func (r *reqbase) SetCampaignHash(hash string) {
	r.campaignHash = hash
}
func (r *reqbase) CampaignId() int64 {
	return r.campaignId
}
func (r *reqbase) SetCampaignId(id int64) {
	r.campaignId = id
}
func (r *reqbase) CampaignName() string {
	return r.campaignName
}
func (r *reqbase) SetCampaignName(name string) {
	r.campaignName = name
}
func (r *reqbase) CampaignCountry() string {
	return r.campaignCountry
}
func (r *reqbase) SetCampaignCountry(country string) {
	r.campaignCountry = country
}
func (r *reqbase) FlowId() int64 {
	return r.flowId
}
func (r *reqbase) SetFlowId(id int64) {
	r.flowId = id
}

func (r *reqbase) FlowName() string {
	return r.flowName
}
func (r *reqbase) SetFlowName(name string) {
	r.flowName = name
}

func (r *reqbase) RuleId() int64 {
	return r.ruleId
}
func (r *reqbase) SetRuleId(id int64) {
	r.ruleId = id
}
func (r *reqbase) PathId() int64 {
	return r.pathId
}
func (r *reqbase) SetPathId(id int64) {
	r.pathId = id
}
func (r *reqbase) LanderId() int64 {
	return r.landerId
}
func (r *reqbase) SetLanderId(id int64) {
	r.landerId = id
}
func (r *reqbase) LanderName() string {
	return r.landerName
}
func (r *reqbase) SetLanderName(name string) {
	r.landerName = name
}
func (r *reqbase) OfferId() int64 {
	return r.offerId
}
func (r *reqbase) SetOfferId(id int64) {
	r.offerId = id
}
func (r *reqbase) OptOfferId() int64 {
	return r.optOfferId
}
func (r *reqbase) SetOptOfferId(id int64) {
	r.optOfferId = id
}
func (r *reqbase) OfferName() string {
	return r.offerName
}
func (r *reqbase) SetOfferName(name string) {
	r.offerName = name
}
func (r *reqbase) AffiliateId() int64 {
	return r.affiliateId
}
func (r *reqbase) SetAffiliateId(id int64) {
	r.affiliateId = id
}
func (r *reqbase) OptAffiliateId() int64 {
	return r.optAffiliateId
}
func (r *reqbase) SetOptAffiliateId(id int64) {
	r.optAffiliateId = id
}
func (r *reqbase) AffiliateName() string {
	return r.affiliateName
}
func (r *reqbase) SetAffiliateName(name string) {
	r.affiliateName = name
}

func (r *reqbase) ImpTimeStamp() int64 {
	return r.impTimeStamp
}
func (r *reqbase) SetImpTimeStamp(timestamp int64) {
	r.impTimeStamp = timestamp
}
func (r *reqbase) VisitTimeStamp() int64 {
	return r.visitTimeStamp
}
func (r *reqbase) SetVisitTimeStamp(timestamp int64) {
	r.visitTimeStamp = timestamp
}
func (r *reqbase) ClickTimeStamp() int64 {
	return r.clickTimeStamp
}
func (r *reqbase) SetClickTimeStamp(timestamp int64) {
	r.clickTimeStamp = timestamp
}
func (r *reqbase) PostBackTimeStamp() int64 {
	return r.postbackTimeStamp
}
func (r *reqbase) SetPostbackTimeStamp(timestamp int64) {
	r.postbackTimeStamp = timestamp
}

func (r *reqbase) DeviceType() string {
	return r.deviceType
}
func (r *reqbase) UserAgent() string {
	return r.ua
}
func (r *reqbase) RemoteIp() string {
	return r.ip
}
func (r *reqbase) Language() string {
	return r.language
}
func (r *reqbase) Model() string {
	return r.model
}
func (r *reqbase) CountryCode() string { // ISO-ALPHA-3
	return r.countryCode
}
func (r *reqbase) CountryName() string {
	return r.countryName
}
func (r *reqbase) City() string {
	return r.city
}
func (r *reqbase) Region() string {
	return r.region
}
func (r *reqbase) Carrier() string {
	return r.carrier
}
func (r *reqbase) ISP() string {
	return r.isp
}
func (r *reqbase) TrackingDomain() string {
	return r.trackingDomain
}
func (r *reqbase) Referrer() string {
	return r.referrer
}
func (r *reqbase) ReferrerDomain() string {
	return r.referrerdomain
}
func (r *reqbase) Brand() string {
	return r.brand
}
func (r *reqbase) OS() string {
	return r.os
}
func (r *reqbase) OSVersion() string {
	return r.osVersion
}
func (r *reqbase) Browser() string {
	return r.browser
}
func (r *reqbase) BrowserVersion() string {
	return r.browserVersion
}
func (r *reqbase) ConnectionType() string {
	return r.connectionType
}
func (r *reqbase) IsBot() bool {
	return r.bot
}

func (r *reqbase) AddCookie(key string, value string) {
	r.cookie[key] = value
}
func (r *reqbase) DelCookie(key string) {
	delete(r.cookie, key)
}
func (r *reqbase) CookieString() (encode string) {
	encode = ""
	for k, v := range r.cookie {
		encode += fmt.Sprintf("%s=%s&", k, v)
	}
	return encode[:len(encode)-1]
}

func (r *reqbase) AddUrlParam(key string, value string) {
	r.urlParam[key] = value
}
func (r *reqbase) DelUrlParam(key string) {
	delete(r.urlParam, key)
}
func (r *reqbase) UrlParamString() (encode string) {
	encode = ""
	for k, v := range r.urlParam {
		encode += fmt.Sprintf("%s=%s&", k, v)
	}
	return encode[:len(encode)-1]
}

func (r *reqbase) ParseUrlTokens(url string) string {
	url = strings.Replace(url, "{externalid}", r.ExternalId(), -1)
	url = strings.Replace(url, "{payout}", fmt.Sprintf("%v", r.Payout()), -1)
	url = strings.Replace(url, "{campaign.id}", fmt.Sprintf("%v", r.CampaignId()), -1)
	url = strings.Replace(url, "{trafficsource.id}", fmt.Sprintf("%v", r.TrafficSourceId()), -1)
	url = strings.Replace(url, "{trafficsource.name}", r.TrafficSourceName(), -1)
	url = strings.Replace(url, "{lander.id}", fmt.Sprintf("%v", r.LanderId()), -1)
	url = strings.Replace(url, "{offer.id}", fmt.Sprintf("%v", r.OfferId()), -1)
	url = strings.Replace(url, "{offer.id}", fmt.Sprintf("%v", r.OfferId()), -1)
	url = strings.Replace(url, "{device}", r.DeviceType(), -1)
	url = strings.Replace(url, "{brand}", r.Brand(), -1)
	url = strings.Replace(url, "{model}", r.Model(), -1)
	url = strings.Replace(url, "{browser}", r.Browser(), -1)
	url = strings.Replace(url, "{browserversion}", r.BrowserVersion(), -1)
	url = strings.Replace(url, "{os}", r.OS(), -1)
	url = strings.Replace(url, "{osversion}", r.OSVersion(), -1)
	url = strings.Replace(url, "{country}", r.CountryCode(), -1)
	url = strings.Replace(url, "{countrycode}", r.CountryCode(), -1)
	url = strings.Replace(url, "{countryname}", r.CountryName(), -1)
	url = strings.Replace(url, "{region}", r.Region(), -1)
	url = strings.Replace(url, "{city}", r.City(), -1)
	url = strings.Replace(url, "{isp}", r.ISP(), -1)
	url = strings.Replace(url, "{useragent}", r.UserAgent(), -1)
	url = strings.Replace(url, "{connection.type}", r.ConnectionType(), -1)
	url = strings.Replace(url, "{carrier}", r.Carrier(), -1)
	url = strings.Replace(url, "{ip}", r.RemoteIp(), -1)
	url = strings.Replace(url, "{referrerdomain}", r.ReferrerDomain(), -1)
	url = strings.Replace(url, "{language}", r.Language(), -1)
	url = strings.Replace(url, "{transaction.id}", r.TransactionId(), -1)
	url = strings.Replace(url, "{click.id}", r.Id(), -1) // ClickId是我们自己的Visits的id

	// {varN},N:1~VarsMaxNum
	for i := 0; i < len(r.TSVars()); i++ {
		vn := r.Vars(uint(i))
		if len(vn) != 0 {
			url = strings.Replace(url, fmt.Sprintf("{var%d}", i+1), vn, -1)
		}
	}

	// {var:NAME}
	for i := 0; i < len(r.TSVars()); i++ {
		vn := r.Vars(uint(i))
		if len(vn) != 0 {
			from := fmt.Sprintf("{var:%s}", r.TSVars()[i].Name)
			url = strings.Replace(url, from, vn, -1)
		}
	}

	url = strings.Replace(url, "{campaign.cpa}", fmt.Sprintf("%v", r.CPAValue()), -1)
	return url
}

func (r *reqbase) CacheSave(localDuration, remoteDuration time.Duration) bool {
	if err := setReqCache(r, localDuration, remoteDuration); err != nil {
		log.Errorf("[reqbase][CacheSave]setReqCache failed for expire(%s:%s) with err(%s) for request(%s)\n",
			localDuration.String(), remoteDuration.String(), err.Error(), r.String())
		return false
	}
	return true
}
func (r *reqbase) CacheClear(local bool) bool {
	delReqCache(r.id, local)
	return true
}

func (r *reqbase) AdStatisKey(timestamp int64) (key tracking.AdStatisKey) {
	if r == nil {
		return
	}
	key.UserID = r.UserId()
	key.CampaignID = r.CampaignId()
	key.CampaignName = r.CampaignName()
	key.FlowID = r.FlowId()
	key.FlowName = r.FlowName()
	key.LanderID = r.LanderId()
	key.LanderName = r.LanderName()
	if r.OfferId() > 0 { // 默认记录OfferId
		key.OfferID = r.OfferId()
	} else { // 当OfferId无效时，说明处于Impression或者LPOfferReq阶段
		key.OfferID = r.OptOfferId()
	}
	key.OfferName = r.OfferName()
	if r.AffiliateId() > 0 { // 默认记录AffiliateId
		key.AffiliateNetworkID = r.AffiliateId()
	} else { // 当AffiliateId无效时，说明处于Impression或者LPOfferReq阶段
		key.AffiliateNetworkID = r.OptAffiliateId()
	}
	key.AffiliateNetworkName = r.AffiliateName()
	key.TrafficSourceID = r.TrafficSourceId()
	key.TrafficSourceName = r.TrafficSourceName()
	key.Language = r.Language()
	key.Model = r.Model()
	key.Country = r.CountryCode()
	key.City = r.City()
	key.Region = r.Region()
	key.ISP = r.ISP()
	key.MobileCarrier = r.Carrier()
	key.Domain = r.ReferrerDomain()
	key.DeviceType = r.DeviceType()
	key.Brand = r.Brand()
	key.OS = r.OS()
	key.OSVersion = r.OSVersion()
	key.Browser = r.Browser()
	key.BrowserVersion = r.BrowserVersion()
	key.ConnectionType = r.ConnectionType()
	key.Timestamp = timestamp

	key.TSCampaignId = r.TSCampaignId()
	key.TSWebsiteId = r.WebsiteId()

	// 解析v1-v10
	v := []*string{&key.V1, &key.V2, &key.V3, &key.V4, &key.V5, &key.V6, &key.V7, &key.V8, &key.V9, &key.V10}
	for i := 0; i < len(v); i++ {
		*v[i] = r.Vars(uint(i))
	}
	return
}
func (r *reqbase) IPKey(timestamp int64) tracking.IPStatisKey {
	var ipKey tracking.IPStatisKey
	if r == nil {
		return ipKey
	}
	ipKey.UserID = r.UserId()
	ipKey.Timestamp = timestamp
	ipKey.CampaignID = r.CampaignId()
	ipKey.IP = r.RemoteIp()
	//ipKey.CountryCode = countrycode.CountryCode2To3(ip2location.Get_country_short(ipKey.IP).Country_short)
	return ipKey
}
func (r *reqbase) ReferrerKey(timestamp int64) tracking.ReferrerStatisKey {
	var referrerKey tracking.ReferrerStatisKey
	if r == nil {
		return referrerKey
	}
	referrerKey.UserID = r.UserId()
	referrerKey.Timestamp = timestamp
	referrerKey.CampaignID = r.CampaignId()
	referrerKey.Referrer = r.Referrer()
	return referrerKey
}
func (r *reqbase) DomainKey(timestamp int64) tracking.ReferrerDomainStatisKey {
	var domain tracking.ReferrerDomainStatisKey
	if r == nil {
		return domain
	}
	domain.UserID = r.UserId()
	domain.Timestamp = timestamp
	domain.CampaignID = r.CampaignId()
	domain.ReferrerDomain = r.ReferrerDomain()
	return domain
}
func (r *reqbase) ConversionKey() tracking.Conversion {
	var conv tracking.Conversion
	if r == nil {
		return conv
	}
	conv.UserID = r.UserId()
	conv.PostbackTimestamp = r.PostBackTimeStamp()
	conv.VisitTimestamp = r.VisitTimeStamp()
	conv.ExternalID = r.ExternalId()
	conv.ClickID = r.Id()
	conv.TransactionID = r.TransactionId()
	conv.Revenue = r.Payout()
	conv.Cost = r.Cost()
	conv.CampaignID = r.CampaignId()
	conv.CampaignName = r.CampaignName()
	conv.LanderID = r.LanderId()
	conv.LanderName = r.LanderName()

	conv.OfferID = r.OfferId()
	conv.OfferName = r.OfferName()
	conv.Country = r.CountryCode()
	conv.CountryCode = r.CountryCode()
	conv.TrafficSourceName = r.TrafficSourceName()
	conv.TrafficSourceID = r.TrafficSourceId()
	conv.AffiliateNetworkID = r.AffiliateId()
	conv.AffiliateNetworkName = r.AffiliateName()
	// conv.Device = r.Device	// 目前还没有
	conv.OS = r.OS()
	conv.OSVersion = r.OSVersion()
	conv.Brand = r.Brand()
	conv.Model = r.Model()
	conv.Browser = r.Browser()
	conv.BrowserVersion = r.BrowserVersion()
	conv.ISP = r.ISP()
	conv.MobileCarrier = r.Carrier()
	conv.VisitorIP = r.RemoteIp()
	conv.VisitorReferrer = r.Referrer()

	// 解析v1-v10
	v := []*string{&conv.V1, &conv.V2, &conv.V3, &conv.V4, &conv.V5, &conv.V6, &conv.V7, &conv.V8, &conv.V9, &conv.V10}
	for i := 0; i < len(v); i++ {
		*v[i] = r.Vars(uint(i))
	}
	return conv
}

func (r *reqbase) SetPayout(p float64) {
	r.payout = p
}

func (r *reqbase) Payout() float64 {
	return r.payout
}

func (r *reqbase) SetTransactionId(txid string) {
	r.txid = txid
}

func (r *reqbase) TransactionId() string {
	return r.txid
}

// SetRedirectMode 设置跳转模式（0:302;1:Meta refresh;2:Double meta refresh）
func (r *reqbase) SetRedirectMode(m int64) {
	r.redirectMode = m
}

// GetRedirectMode 获取跳转模式（0:302;1:Meta refresh;2:Double meta refresh）
func (r *reqbase) GetRedirectMode() int64 {
	return r.redirectMode
}

// Redirect 根据当前的跳转模式跳转
func (r *reqbase) Redirect(w http.ResponseWriter, req *http.Request, dest string) {
	switch r.redirectMode {
	case 1:
		// meta refresh
		w.Header().Set("Content-Type", "text/html")
		meta := `<meta http-equiv="refresh" content="0;url=` + html.EscapeString(dest) + `">`
		fmt.Fprintln(w, meta)
	case 2:
		// double meta refresh
		w.Header().Set("Content-Type", "text/html")
		to := url.QueryEscape(dest)
		first := fmt.Sprintf("/dmr?dest=%s", to)
		meta := `<meta http-equiv="refresh" content="0;url=` + html.EscapeString(first) + `">`
		fmt.Fprintln(w, meta)
	default: // 0或者其它非法的值，都是302跳转
		http.Redirect(w, req, dest, http.StatusFound)
	}
}

func (r *reqbase) TSCampaignId() string {
	return r.tsCId
}

func (r *reqbase) WebsiteId() string {
	return r.websiteId
}
