package request

import (
	"fmt"
	"net/http"
	"time"

	"Service/common"
	"Service/tracking"
)

const (
	ReqLPOffer           = "lpofferreq"
	ReqLPClick           = "landingpageclick"
	ReqImpression        = "impression"
	ReqS2SPostback       = "s2spostback"
	ReqUploadConversions = "conversions"
	ReqConversionPixel   = "conversionpixel"
	ReqConversionScript  = "conversionscript"
)

const (
	VarsMaxNum = 10
)

type QueryHolder interface {
	Get(key string) string
	//Set(key, value string)
}

type Request interface {
	Id() string
	SetId(id string)
	Type() string
	String() string

	ParseUrlTokens(url string) string

	ExternalId() string
	Cost() float64
	TSCampaignId() string
	WebsiteId() string
	Vars(n uint) string // n:0~VarsMaxNum-1
	ParseTSParams(
		externalId common.TrafficSourceParams,
		cost common.TrafficSourceParams,
		params []common.TrafficSourceParams,
		tsCampaign common.TrafficSourceParams,
		tsWebsite common.TrafficSourceParams,
		values QueryHolder)

	TrafficSourceId() int64
	SetTrafficSourceId(id int64)
	TrafficSourceName() string
	SetTrafficSourceName(name string)
	UserId() int64
	SetUserId(id int64)
	UserIdText() string
	SetUserIdText(idText string)
	CampaignHash() string
	SetCampaignHash(hash string)
	CampaignId() int64
	SetCampaignId(id int64)
	CampaignName() string
	SetCampaignName(name string)
	CampaignCountry() string
	SetCampaignCountry(country string)
	FlowId() int64
	SetFlowId(id int64)
	SetFlowName(name string)
	RuleId() int64
	SetRuleId(id int64)
	PathId() int64
	SetPathId(id int64)
	LanderId() int64
	SetLanderId(id int64)
	LanderName() string
	SetLanderName(name string)
	OfferId() int64
	SetOfferId(id int64)
	OptOfferId() int64
	SetOptOfferId(id int64)
	OfferName() string
	SetOfferName(name string)
	AffiliateId() int64
	SetAffiliateId(id int64)
	OptAffiliateId() int64
	SetOptAffiliateId(id int64)
	AffiliateName() string
	SetAffiliateName(name string)

	// timeStamp,都是精确到毫秒
	ImpTimeStamp() int64
	SetImpTimeStamp(timestamp int64)
	VisitTimeStamp() int64
	SetVisitTimeStamp(timestamp int64)
	ClickTimeStamp() int64
	SetClickTimeStamp(timestamp int64)
	PostBackTimeStamp() int64
	SetPostbackTimeStamp(timestamp int64)

	DeviceType() string
	UserAgent() string
	RemoteIp() string
	Language() string
	Model() string
	CountryCode() string // ISO-ALPHA-3
	CountryName() string
	City() string
	Region() string
	Carrier() string
	ISP() string
	TrackingDomain() string
	TrackingPath() string
	Referrer() string
	ReferrerDomain() string
	Brand() string
	OS() string
	OSVersion() string
	Browser() string
	BrowserVersion() string
	ConnectionType() string
	IsBot() bool

	AddCookie(key, value string)
	DelCookie(key string)
	CookieString() string

	AddUrlParam(key, value string)
	DelUrlParam(key string)
	UrlParamString() string // 不包括"?"和"/"部分

	// duration
	// < 0:do not save
	// = 0:save, without expiration
	// > 0:save, expired after duration
	CacheSave(localDuration, remoteDuration time.Duration) bool
	CacheClear(local bool) bool

	// for tracking
	AdStatisKey(timestamp int64) tracking.AdStatisKey
	IPKey(timestamp int64) tracking.IPStatisKey
	ReferrerKey(timestamp int64) tracking.ReferrerStatisKey
	DomainKey(timestamp int64) tracking.ReferrerDomainStatisKey
	ConversionKey() tracking.Conversion

	// Payout
	SetPayout(p float64)
	Payout() float64
	SetTransactionId(txid string)
	TransactionId() string

	//	SetTSExternalID(e *common.TrafficSourceParams) // 暂时不需要，去掉
	//	SetTSCost(c *common.TrafficSourceParams) // 暂时不需要，去掉
	SetTSVars(vars []common.TrafficSourceParams)
	SetCPAValue(v float64)
	//	TSExternalID() *common.TrafficSourceParams // 暂时不需要，去掉
	//	TSCost() *common.TrafficSourceParams // 暂时不需要，去掉
	TSVars() []common.TrafficSourceParams
	CPAValue() float64

	SetRedirectMode(m int64)
	GetRedirectMode() int64
	Redirect(w http.ResponseWriter, req *http.Request, url string)
}

func CreateRequest(reqId string, brandNew bool, reqType string, r *http.Request) (req Request, err error) {
	if reqId == "" || reqType == "" {
		return nil, fmt.Errorf("[CreateRequest]Either reqId or reqType is empty for request for %s", common.SchemeHostURI(r))
	}
	switch reqType {
	case ReqImpression:
		req = CreateImpressionRequest(reqId, r)
	case ReqLPOffer:
		req = CreateLPOfferRequest(reqId, r, brandNew)
		return req, nil
	case ReqLPClick:
		req = CreateLPClickRequest(reqId, r)
		return req, nil
	case ReqS2SPostback:
		req = CreateS2SPostbackRequest(reqId, r)
	case ReqUploadConversions:
		req = CreateUploadConversionsRequest(reqId, r)
	case ReqConversionPixel:
		//TODO
	case ReqConversionScript:
		//TODO
	}
	if req == nil {
		return nil, fmt.Errorf("CreateRequest failed for %s;%s;%s%s", reqType, reqId, r.Host, r.RequestURI)
	}
	return req, nil
}
