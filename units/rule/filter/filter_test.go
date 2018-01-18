package filter

import (
	"Service/common"
	"Service/config"
	"Service/log"
	"Service/request"
	"Service/tracking"
	"flag"
	"fmt"
	"net/http"
	"testing"
	"time"
)

var testfilter string = `[[["os","in","#ALL IOS"],["country","in","#SameAsCampaign","JPN"]]]`

func TestAdFilter(t *testing.T) {
	flag.Set("config", `C:\workspace\innotech_ssp\src\AdClickServer\Service\config\development.ini`)
	flag.Parse()
	if err := config.LoadConfig(true); err != nil {
		panic(err.Error())
	}
	log.Init("console", `{"level":7}`, false)
	nf, err := NewFilter(testfilter)
	if err != nil {
		fmt.Println("NewFilter error:", err.Error())
		t.Fail()
	}

	if !nf.Accept(getFakeRequest()) {
		t.Fail()
	}
}

func getFakeRequest() request.Request {
	return &fakeRequest{}
}

type fakeRequest struct{}

func (f fakeRequest) Id() string {
	return "XXXYYYZZZ"
}
func (f fakeRequest) Type() string {
	return request.ReqImpression
}
func (f fakeRequest) String() string {
	return "a fake request.Request"
}
func (f fakeRequest) ParseUrlTokens(url string) string {
	return url
}
func (f fakeRequest) ExternalId() string {
	return "externalYYY"
}
func (f fakeRequest) Cost() float64 {
	return 0.0025
}
func (f fakeRequest) Vars(n uint) string {
	return fmt.Sprintf("%d", n*10)
}
func (f fakeRequest) ParseTSParams(
	externalId common.TrafficSourceParams,
	cost common.TrafficSourceParams,
	params []common.TrafficSourceParams,
	tsCampaign common.TrafficSourceParams,
	tsWebsite common.TrafficSourceParams,
	values request.QueryHolder) {

}
func (f fakeRequest) TrafficSourceId() int64 {
	return 100
}
func (f fakeRequest) SetTrafficSourceId(id int64) {
}
func (f fakeRequest) TrafficSourceName() string {
	return "fakeTrafficeSourceName"
}
func (f fakeRequest) SetTrafficSourceName(name string) {
}
func (f fakeRequest) UserId() int64 {
	return 200
}
func (f fakeRequest) SetUserId(id int64) {
}
func (f fakeRequest) UserIdText() string {
	return "zx1jg"
}
func (f fakeRequest) SetUserIdText(idText string) {
}
func (f fakeRequest) CampaignHash() string {
	return "8a145d88-39d3-48fd-ab89-7e0e8a856301"
}
func (f fakeRequest) SetCampaignHash(hash string) {
}
func (f fakeRequest) CampaignId() int64 {
	return 300
}
func (f fakeRequest) SetCampaignId(id int64) {
}
func (f fakeRequest) CampaignName() string {
	return "Popads - Viet Nam - mobisummer4-Cleaner-benson-0113"
}
func (f fakeRequest) SetCampaignName(name string) {
}
func (f fakeRequest) CampaignCountry() string {
	return "ZZZ"
}
func (f fakeRequest) SetCampaignCountry(country string) {
}
func (f fakeRequest) FlowId() int64 {
	return 400
}
func (f fakeRequest) SetFlowId(id int64) {
}
func (f fakeRequest) RuleId() int64 {
	return 500
}
func (f fakeRequest) SetRuleId(id int64) {
}
func (f fakeRequest) PathId() int64 {
	return 600
}
func (f fakeRequest) SetPathId(id int64) {
}
func (f fakeRequest) LanderId() int64 {
	return 700
}
func (f fakeRequest) SetLanderId(id int64) {
}
func (f fakeRequest) LanderName() string {
	return "United States - zhuanpan"
}
func (f fakeRequest) SetLanderName(name string) {
}
func (f fakeRequest) OfferId() int64 {
	return 800
}
func (f fakeRequest) SetOfferId(id int64) {
}
func (f fakeRequest) OfferName() string {
	return "EffectMobi - United States - yoshop1.8"
}
func (f fakeRequest) SetOfferName(name string) {
}
func (f fakeRequest) AffiliateId() int64 {
	return 900
}
func (f fakeRequest) SetAffiliateId(id int64) {
}
func (f fakeRequest) AffiliateName() string {
	return "EffectMobi"
}
func (f fakeRequest) SetAffiliateName(name string) {
}
func (f fakeRequest) ImpTimeStamp() int64 {
	return time.Now().Unix()
}
func (f fakeRequest) SetImpTimeStamp(timestamp int64) {
}
func (f fakeRequest) VisitTimeStamp() int64 {
	return time.Now().Unix()
}
func (f fakeRequest) SetVisitTimeStamp(timestamp int64) {
}
func (f fakeRequest) ClickTimeStamp() int64 {
	return time.Now().Unix()
}
func (f fakeRequest) SetClickTimeStamp(timestamp int64) {
}
func (f fakeRequest) PostBackTimeStamp() int64 {
	return time.Now().Unix()
}
func (f fakeRequest) SetPostbackTimeStamp(timestamp int64) {
}
func (f fakeRequest) DeviceType() string {
	return "Mobile"
}
func (f fakeRequest) UserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36"
}
func (f fakeRequest) RemoteIp() string {
	return "180.168.12.58"
}
func (f fakeRequest) Language() string {
	return "en"
}
func (f fakeRequest) Model() string {
	return "iPhone 5"
}
func (f fakeRequest) CountryCode() string { // ISO-ALPHA-3
	return "JPN"
}
func (f fakeRequest) CountryName() string {
	return "United States"
}
func (f fakeRequest) City() string {
	return "New York"
}
func (f fakeRequest) Region() string {
	return "New York"
}
func (f fakeRequest) Carrier() string {
	return "AT&T"
}
func (f fakeRequest) ISP() string {
	return "AT&T"
}
func (f fakeRequest) TrackingDomain() string {
	return "zx1jg.voluumtrk.com"
}
func (f fakeRequest) TrackingPath() string {
	return "/8a145d88-39d3-48fd-ab89-7e0e8a856301?WEBSITEID=[WEBSITEID]&CATEGORYNAME=[SCREENRESOLUTION]&ISPNAME=[ISPNAME]&QUALITY=[QUALITY]&BID=[BID]&SCREENRESOLUTION=[SCREENRESOLUTION]&BROWSERNAME=[BROWSERNAME]&ADBLOCK=[ADBLOCK]"
}
func (f fakeRequest) Referrer() string {
	return "www.newyorker.com/topnews"
}
func (f fakeRequest) ReferrerDomain() string {
	return "www.newyorker.com"
}
func (f fakeRequest) Brand() string {
	return "Apple"
}
func (f fakeRequest) OS() string {
	return "IOS"
}
func (f fakeRequest) OSVersion() string {
	return "9.1"
}
func (f fakeRequest) Browser() string {
	return "Chrome"
}
func (f fakeRequest) BrowserVersion() string {
	return "55.0.2883.87"
}
func (f fakeRequest) ConnectionType() string {
	return "Mobile"
}
func (f fakeRequest) IsBot() bool {
	return false
}
func (f fakeRequest) AddCookie(key, value string) {
}
func (f fakeRequest) DelCookie(key string) {
}
func (f fakeRequest) CookieString() string {
	return ""
}
func (f fakeRequest) AddUrlParam(key, value string) {
}
func (f fakeRequest) DelUrlParam(key string) {
}
func (f fakeRequest) UrlParamString() string {
	return ""
}
func (f fakeRequest) CacheSave(localDuration, remoteDuration time.Duration) bool {
	return true
}
func (f fakeRequest) CacheClear(local bool) bool {
	return true
}
func (f fakeRequest) AdStatisKey(timestamp int64) (key tracking.AdStatisKey) {
	return
}
func (f fakeRequest) IPKey(timestamp int64) (key tracking.IPStatisKey) {
	return
}
func (f fakeRequest) ReferrerKey(timestamp int64) (key tracking.ReferrerStatisKey) {
	return
}
func (f fakeRequest) DomainKey(timestamp int64) (key tracking.ReferrerDomainStatisKey) {
	return
}
func (f fakeRequest) ConversionKey() (key tracking.Conversion) {
	return
}
func (f fakeRequest) SetPayout(p float64) {
}
func (f fakeRequest) Payout() float64 {
	return 0.012
}
func (f fakeRequest) SetTransactionId(txid string) {
}
func (f fakeRequest) TransactionId() string {
	return "fakeTransactionIdZZZ"
}
func (f fakeRequest) SetTSExternalID(e *common.TrafficSourceParams) {
}
func (f fakeRequest) SetTSCost(c *common.TrafficSourceParams) {
}
func (f fakeRequest) SetTSVars(vars []common.TrafficSourceParams) {
}
func (f fakeRequest) SetCPAValue(v float64) {
}
func (f fakeRequest) TSExternalID() *common.TrafficSourceParams {
	return nil
}
func (f fakeRequest) TSCost() *common.TrafficSourceParams {
	return nil
}
func (f fakeRequest) TSVars() []common.TrafficSourceParams {
	return nil
}
func (f fakeRequest) CPAValue() float64 {
	return 0.02
}
func (f fakeRequest) SetRedirectMode(m int64) {
}
func (f fakeRequest) GetRedirectMode() int64 {
	return 1
}
func (f fakeRequest) Redirect(w http.ResponseWriter, req *http.Request, url string) {
}

func (f fakeRequest) AppendClickOfferId(id int64) {
}
func (f fakeRequest) ClickOfferIdExisted(id int64) bool {
	return true
}
func (f fakeRequest) OptOfferId() int64 {
	return 1
}
func (f fakeRequest) SetOptOfferId(id int64) {
}
func (f fakeRequest) OptAffiliateId() int64 {
	return 1
}
func (f fakeRequest) SetOptAffiliateId(id int64) {

}
func (f fakeRequest) SetFlowName(name string) {

}
func (f fakeRequest) SetId(id string) {

}
func (f fakeRequest) TSCampaignId() string {
	return ""
}
func (f fakeRequest) WebsiteId() string {
	return ""
}
