package campaign

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"Service/common"
	"Service/log"
	"Service/request"
	"Service/units/ffrule"
	"Service/units/flow"
)

const (
	//0:URL;1:Flow;2:Rule;3:Path;4:Lander;5:Offer
	TargetTypeUrl    = 0
	TargetTypeFlow   = 1
	TargetTypeRule   = 2
	TargetTypePath   = 3
	TargetTypeLander = 4
	TargetTypeOffer  = 5
)

// TrafficSourceConfig 对应数据库里面的TrafficSource
type TrafficSourceConfig struct {
	Id               int64
	UserId           int64
	Name             string
	PostbackURL      string
	PixelRedirectURL string //TODO
	ImpTracking      int64

	ExternalId   common.TrafficSourceParams
	Cost         common.TrafficSourceParams
	Vars         []common.TrafficSourceParams
	TSCampaignId common.TrafficSourceParams
	TSWebsitId   common.TrafficSourceParams
}

func (c TrafficSourceConfig) ID() int64 {
	return c.Id
}

func (c TrafficSourceConfig) GetCopy() (cc TrafficSourceConfig) {
	cc = c
	cc.Vars = make([]common.TrafficSourceParams, len(c.Vars))
	for i := range c.Vars {
		cc.Vars[i] = c.Vars[i]
	}
	return
}

type CampaignFFRule struct {
	RuleId int64
	Active bool
}

type CampaignConfig struct {
	Id                int64
	Name              string
	UserId            int64
	Hash              string
	Url               string
	ImpPixelUrl       string
	TrafficSourceId   int64
	TrafficSourceName string
	CostModel         int
	CPCValue          float64
	CPAValue          float64
	CPMValue          float64
	PostbackUrl       string
	PixelRedirectUrl  string //TODO
	RedirectMode      int64
	TargetType        int64
	TargetFlowId      int64
	TargetUrl         string
	Status            int64
	Country           string // alpha-3 country code

	// 每个campaign的link中包含的参数(traffic source会进行替换，但是由用户自己指定)
	// 例如：[["bannerid","{bannerid}"],["campaignid","{campaignid}"],["zoneid","{zoneid}"]]
	TrafficSource TrafficSourceConfig
}

func (c CampaignConfig) GetCopy() (cc CampaignConfig) {
	cc = c
	cc.TrafficSource = c.TrafficSource.GetCopy()
	return
}

func (c CampaignConfig) ID() int64 {
	return c.Id
}

// ParseVars 根据Vars解析出10个参数，分别是，v1-v10
func (c *CampaignConfig) ParseVars(getter func(k string) string) []string {
	vars := []string{}
	for _, param := range c.TrafficSource.Vars {
		v := getter(param.Parameter)
		vars = append(vars, v)
	}
	return vars
}

func (c CampaignConfig) String() string {
	return fmt.Sprintf("Campaign %d:%d Status %d", c.Id, c.UserId, c.Status)
}

type Campaign struct {
	CampaignConfig

	// fraud filter rules
	ff []int64
}

var cmu sync.RWMutex                           // protects the following
var campaigns = make(map[int64]*Campaign)      // campaignId:instance
var campaignHash2Id = make(map[string]int64)   // campaignHash:campaignId
var tsId2CampaignIds = make(map[int64][]int64) // trafficSourceId:[]campaignId
func setCampaign(ca *Campaign) error {
	if ca == nil {
		return errors.New("setCampaign error:ca is nil")
	}
	if ca.Id <= 0 {
		return fmt.Errorf("setCampaign error:ca.Id(%d) is not positive", ca.Id)
	}
	log.Debugf("[setCampaign]ca.Id(%d),ca.Hash(%s)\n", ca.Id, ca.Hash)
	cmu.Lock()
	defer cmu.Unlock()
	campaigns[ca.Id] = ca
	campaignHash2Id[ca.Hash] = ca.Id
	if ca.TrafficSourceId > 0 {
		tsId2CampaignIds[ca.TrafficSourceId] = append(tsId2CampaignIds[ca.TrafficSourceId], ca.Id)
	}
	return nil
}
func getCampaign(campaignId int64) *Campaign {
	if campaignId <= 0 {
		return nil
	}
	cmu.RLock()
	defer cmu.RUnlock()
	return campaigns[campaignId]
}
func getCampaignByHash(campaignHash string) *Campaign {
	if campaignHash == "" {
		return nil
	}
	cmu.RLock()
	defer cmu.RUnlock()
	if campaignId, ok := campaignHash2Id[campaignHash]; ok {
		return campaigns[campaignId]
	}
	return nil
}
func getCampaignIdsByTSId(tsId int64) (caIds []int64) {
	if tsId <= 0 {
		return
	}
	cmu.RLock()
	defer cmu.RUnlock()
	if cIds := tsId2CampaignIds[tsId]; len(cIds) > 0 {
		caIds = make([]int64, len(cIds))
		for i, id := range cIds {
			caIds[i] = id
		}
	}
	return
}
func delCampaign(campaignId int64) {
	cmu.Lock()
	defer cmu.Unlock()
	if c, ok := campaigns[campaignId]; ok && c != nil {
		delete(campaigns, campaignId)
		delete(campaignHash2Id, c.Hash)
		if c.TrafficSourceId > 0 {
			cIds := tsId2CampaignIds[c.TrafficSourceId]
			for i, cId := range cIds {
				if cId == campaignId {
					// remove found campaignId
					tsId2CampaignIds[c.TrafficSourceId] = append(cIds[:i], cIds[i+1:]...)
				}
			}
		}
	}
}

func newCampaign(c CampaignConfig) (ca *Campaign) {
	if c.TargetUrl == "" && c.TargetFlowId <= 0 {
		log.Errorf("[newCampaign]Both TargetUrl&TargetFlowId are invalid for campaign%d\n", c.Id)
		return nil
	}
	switch c.TargetType {
	case TargetTypeUrl:
		_, err := url.ParseRequestURI(c.TargetUrl)
		if err != nil {
			log.Errorf("[newCampaign]TargetUrl is not a valid url(%s) for campaign%d\n", c.TargetUrl, c.Id)
			// not returning nil, because exists the following case
			// http://apps.coach.lightshines.top/276466/31307=%subid1%: invalid URL escape "%su"
			// return nil
		}
	case TargetTypePath:
		fallthrough
	case TargetTypeFlow:
		if c.TargetFlowId <= 0 {
			log.Errorf("[newCampaign]TargetFlowId(%d) is invalid for campaign%d\n", c.TargetFlowId, c.Id)
			return nil
		}
		if flow.GetFlow(c.TargetFlowId) == nil {
			log.Errorf("[newCampaign]GetFlow failed with flow(%d) for campaign(%d)\n", c.TargetFlowId, c.Id)
			return nil
		}
	default: // currently other target type is not supported
		log.Errorf("[newCampaign]TargetType(%d) is invalid for campaign(%d)\n", c.TargetType, c.Id)
	}
	ca = &Campaign{
		CampaignConfig: c,
		ff:             ffrule.DBGetCampaignAvailableFFRuleIds(c.Id),
	}
	return
}

func InitUserCampaigns(userId int64) error {
	cs := DBGetUserCampaigns(userId)
	var ca *Campaign
	for _, c := range cs {
		ca = newCampaign(c)
		if ca == nil {
			return fmt.Errorf("[InitUserCampaigns]Failed for user(%d) with config(%+v)", userId, c)
		}
		if err := setCampaign(ca); err != nil {
			return err
		}
	}
	return nil
}

func InitCampaign(campaignId int64) error {
	if campaignId <= 0 {
		return fmt.Errorf("campaignId(%d) is invalid", campaignId)
	}
	ca := newCampaign(DBGetCampaign(campaignId))
	if ca == nil {
		return fmt.Errorf("failed for campaignId(%d)", campaignId)
	}
	return setCampaign(ca)
}
func InitTrafficSource(tsId int64) (err error) {
	if tsId <= 0 {
		return fmt.Errorf("traffic source id(%d) is invalid", tsId)
	}
	cIds := getCampaignIdsByTSId(tsId)
	if len(cIds) == 0 {
		return fmt.Errorf("empty campaign ids for tsId(%d)", tsId)
	}
	for _, cId := range cIds {
		err = InitCampaign(cId)
		if err != nil {
			return
		}
	}

	return nil
}
func InitFFRuleCampaigns(ruleId int64) error {
	if ruleId <= 0 {
		return fmt.Errorf("ruleId(%d) is invalid", ruleId)
	}
	// FFRule中移除的Campaign，也要重新Init
	origin := ffrule.GetRule(ruleId)
	if err := ffrule.InitRule(ruleId); err != nil {
		return err
	}
	new := ffrule.GetRule(ruleId)
	toInitCampaignIds := make([]int64, 0)
	if origin.Id == 0 || new.Active != origin.Active {
		toInitCampaignIds = new.CampaignIds
	} else {
		//case1:ruleId是新增的;origin和new应该相等
		//case2:ruleId非新增，只是做了变化，origin和new有差异
		toInitCampaignIds = common.UnionOf(origin.CampaignIds, new.CampaignIds)
	}

	for _, cId := range toInitCampaignIds {
		if err := InitCampaign(cId); err != nil {
			log.Errorf("[InitFFRuleCampaigns]ruleId(%d) got err(%s) for campaign(%d)\n", ruleId, err.Error(), cId)
			continue
		}
	}
	return nil
}
func DeleteFFRuleCampaigns(ruleId int64) error {
	if ruleId <= 0 {
		return fmt.Errorf("ruleId(%d) is invalid", ruleId)
	}
	// FFRule中移除的Campaign，也要重新Init
	origin := ffrule.GetRule(ruleId)
	for _, cId := range origin.CampaignIds {
		if err := InitCampaign(cId); err != nil {
			log.Errorf("[InitFFRuleCampaigns]ruleId(%d) got err(%s) for campaign(%d)\n", ruleId, err.Error(), cId)
			continue
		}
	}
	ffrule.DeleteRule(ruleId)
	return nil
}
func GetCampaign(cId int64) (ca *Campaign) {
	if cId == 0 {
		return nil
	}

	ca = getCampaign(cId)
	if ca == nil {
		ca = newCampaign(DBGetCampaign(cId))
		if ca != nil {
			if err := setCampaign(ca); err != nil {
				return nil
			}
		}
	}

	return
}
func GetCampaignByHash(cHash string) (ca *Campaign) {
	if len(cHash) == 0 {
		return nil
	}

	defer func() {
		log.Infof("[GetCampaignByHash]cHash(%s), ca(%+v)\n", cHash, ca)
	}()
	ca = getCampaignByHash(cHash)
	if ca == nil {
		log.Warnf("GetCampaignByHash(%v) failed: not cached, reloading from db...", cHash)
		ca = newCampaign(DBGetCampaignByHash(cHash))
		if ca != nil {
			if err := setCampaign(ca); err != nil {
				return nil
			}
		}
	}

	return
}
func DelCampaign(campaignId int64) error {
	delCampaign(campaignId)
	return nil
}

var gr = &http.Request{
	Method: "GET",
	URL: &url.URL{
		Path: "",
	},
}

func (ca *Campaign) OnLPOfferRequest(w http.ResponseWriter, req request.Request) (err error) {
	if ca == nil {
		return errors.New("Nil ca")
	}
	//log.Infof("[Campaign][OnLPOfferRequest]Campaign(%s) handles request(%s)\n", ca.String(), req.String())

	for _, ruleId := range ca.ff {
		ffrule.GetRule(ruleId).OnVisits(req)
	}

	//req.SetTSExternalID(&ca.TrafficSource.ExternalId)
	//req.SetTSCost(&ca.TrafficSource.Cost)
	req.SetTSVars(ca.TrafficSource.Vars)
	req.SetCPAValue(ca.CPAValue)
	req.SetRedirectMode(ca.RedirectMode)
	req.SetTrafficSourceId(ca.TrafficSourceId)
	req.SetTrafficSourceName(ca.TrafficSourceName)
	req.SetCampaignCountry(ca.Country)

	if ca.TargetType == TargetTypeUrl {
		if ca.TargetUrl != "" {
			req.Redirect(w, gr, req.ParseUrlTokens(ca.TargetUrl))
			return nil
		}
	} else {
		f := flow.GetFlow(ca.TargetFlowId)
		if f == nil {
			return fmt.Errorf("Nil f(%d) for request(%s) in campaign(%d)", ca.TargetFlowId, req.Id(), ca.Id)
		}
		req.SetFlowId(f.Id)
		req.SetFlowName(f.Name)
		return f.OnLPOfferRequest(w, req)
	}

	return fmt.Errorf("Invalid dstination for request(%s) in campaign(%d)", req.Id(), ca.Id)
}

func (ca *Campaign) OnLandingPageClick(w http.ResponseWriter, req request.Request) error {
	if ca == nil {
		return errors.New("Nil ca")
	}

	for _, ruleId := range ca.ff {
		ffrule.GetRule(ruleId).OnClick(req)
	}

	// 不要用Campaign现在的设置，因为有可能中途被改变
	f := flow.GetFlow(req.FlowId())
	if f == nil {
		return fmt.Errorf("Nil f(%d) for request(%s) in campaign(%d)", req.FlowId(), req.Id(), ca.Id)
	}
	req.SetFlowId(f.Id)
	req.SetFlowName(f.Name)
	return f.OnLandingPageClick(w, req)
}

func (ca *Campaign) OnImpression(w http.ResponseWriter, req request.Request) error {
	if ca == nil {
		return errors.New("Nil ca")
	}
	for _, ruleId := range ca.ff {
		ffrule.GetRule(ruleId).OnImpression(req)
	}
	req.SetTrafficSourceId(ca.TrafficSourceId)
	req.SetTrafficSourceName(ca.TrafficSourceName)
	req.SetCampaignCountry(ca.Country)
	return nil
}

func (ca *Campaign) OnS2SPostback(w http.ResponseWriter, req request.Request) error {
	if ca == nil {
		return errors.New("Nil ca")
	}

	// 对于直接跳转到指定URL的campaign，也要能够postback回traffic source
	go func(req request.Request) {
		var err error
		defer func() {
			if x := recover(); x != nil {
				log.Errorf("[Campaign][OnS2SPostback]PostbackToTrafficSource to trafficsource(%d) failed for request(%s) in campaign(%d) with err(%v)\n", ca.TrafficSourceId, req.Id(), ca.Id, x)
			} else if err != nil {
				log.Errorf("[Campaign][OnS2SPostback]PostbackToTrafficSource to trafficsource(%d) failed for request(%s) in campaign(%d) with err(%s)\n", ca.TrafficSourceId, req.Id(), ca.Id, err.Error())
			}
		}()
		// 这里的err可能是因为对traffic source的postaback失败，暂时对这种错误不记录
		//TODO 找个地方，记录往traffic source的postback失败
		err = ca.PostbackToTrafficSource(req)
	}(req)

	if req.FlowId() == 0 {
		return nil
	}

	f := flow.GetFlow(req.FlowId())
	if f == nil {
		return fmt.Errorf("Nil f(%d) for request(%s) in campaign(%d)", ca.TargetFlowId, req.Id(), ca.Id)
	}
	err := f.OnS2SPostback(w, req)
	if err != nil {
		return err
	}

	return nil
}

func (ca *Campaign) OnConversionPixel(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (ca *Campaign) OnConversionScript(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (ca *Campaign) getPostbackUrl() string {
	if len(ca.PostbackUrl) > 0 {
		return ca.PostbackUrl
	}
	return ca.TrafficSource.PostbackURL
}

func (ca *Campaign) PostbackToTrafficSource(req request.Request) error {
	//TODO 需要检查用户是否已经针对该campaign单独设置了PostbackUrl
	url := req.ParseUrlTokens(ca.getPostbackUrl())
	if len(url) == 0 {
		// 有可能不需要postback
		return nil
	}

	err := func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		log.Infof("[campaign][PostbackToTrafficSource] req:%v success:[%s] with url(%s)\n",
			req.Id(), body, url)
		return nil
	}()

	if err != nil {
		log.Errorf("[campaign][PostbackToTrafficSource] req:%v postback failed:[%v] with url(%s)\n", req.Id(), err, url)
	}

	return nil
}
