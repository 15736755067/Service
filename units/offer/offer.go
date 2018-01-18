package offer

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"Service/db"
	"Service/log"
	"Service/request"
	"Service/units/affiliate"
	"time"
)

const TrackCacheSvrTitle = "MSGQUEUE"

type OfferConfig struct {
	Id                   int64
	Name                 string
	UserId               int64
	Url                  string
	AffiliateNetworkId   int64
	AffiliateNetworkName string
	PostbackUrl          string
	PayoutMode           int64
	PayoutValue          float64
	CapEnabled           int
	DailyCap             int64
	CapTimezoneId        int64
	RedirectOfferId      int64
}

type Timezone struct {
	Id       int64
	Name     string
	Region   string
	UtcShift string
}

func (t Timezone) String() string {
	return fmt.Sprintf("Offer %d:%s", t.Id, t.Region)
}

func (t Timezone) ID() int64 {
	return t.Id
}

func (c OfferConfig) String() string {
	return fmt.Sprintf("Offer %d:%d", c.Id, c.UserId)
}

func (c OfferConfig) ID() int64 {
	return c.Id
}

type Offer struct {
	OfferConfig
}

var cmu sync.RWMutex                        // protects the following
var offers = make(map[int64]*Offer)         // offerId:offerInstance
var anId2OfferIds = make(map[int64][]int64) // affiliateNetworkId:[]offerId

func setOffer(o *Offer) error {
	if o == nil {
		return errors.New("setOffer error:o is nil")
	}
	if o.Id <= 0 {
		return fmt.Errorf("setOffer error:o.Id(%d) is not positive", o.Id)
	}
	cmu.Lock()
	defer cmu.Unlock()
	offers[o.Id] = o
	if o.AffiliateNetworkId > 0 {
		anId2OfferIds[o.AffiliateNetworkId] = append(anId2OfferIds[o.AffiliateNetworkId], o.Id)
	}
	return nil
}

func getOffer(oId int64) *Offer {
	if oId <= 0 {
		return nil
	}
	cmu.RLock()
	defer cmu.RUnlock()
	return offers[oId]
}

func getOfferIdsByANId(anId int64) (oIds []int64) {
	if anId <= 0 {
		return
	}
	cmu.RLock()
	defer cmu.RUnlock()
	if ids := anId2OfferIds[anId]; len(ids) > 0 {
		oIds = make([]int64, len(ids))
		for i, id := range ids {
			oIds[i] = id
		}
	}
	return
}

func delOffer(oId int64) {
	if oId <= 0 {
		return
	}
	cmu.Lock()
	defer cmu.Unlock()
	if o := offers[oId]; o != nil {
		delete(offers, oId)
		oIds := anId2OfferIds[o.AffiliateNetworkId]
		for i, id := range oIds {
			if id == oId {
				// remove from affiliate network maps
				anId2OfferIds[o.AffiliateNetworkId] = append(oIds[:i], oIds[i+1:]...)
				break
			}
		}
	}
}

func InitOffer(offerId int64) error {
	if offerId <= 0 {
		return fmt.Errorf("offerId(%d) is invalid", offerId)
	}
	o := newOffer(DBGetOffer(offerId))
	if o == nil {
		return fmt.Errorf("newOffer failed with offer(%d)", offerId)
	}
	return setOffer(o)
}

func InitAffiliateNetwork(anId int64) (err error) {
	if anId <= 0 {
		return fmt.Errorf("affiliate network id(%d) is invalid", anId)
	}
	oIds := getOfferIdsByANId(anId)
	if len(oIds) == 0 {
		return fmt.Errorf("empty offer ids for anId(%d)", anId)
	}
	for _, oId := range oIds {
		err = InitOffer(oId)
		if err != nil {
			return
		}
	}

	return nil
}

func GetOffer(offerId int64) (o *Offer) {
	if offerId == 0 {
		return nil
	}

	o = getOffer(offerId)
	if o == nil {
		o = newOffer(DBGetOffer(offerId))
		if o != nil {
			if err := setOffer(o); err != nil {
				return nil
			}
		}
	}
	return
}

func newOffer(c OfferConfig) (o *Offer) {
	log.Debugf("[newOffer]%+v\n", c)
	if c.Id <= 0 {
		return nil
	}
	_, err := url.ParseRequestURI(c.Url)
	if err != nil {
		log.Errorf("[newOffer]Invalid url for offer(%+v), err(%s)\n", c, err.Error())
		// not returning nil, because exists the following case
		// http://apps.coach.lightshines.top/276466/31307=%subid1%: invalid URL escape "%su"
		// return nil
	}
	if c.AffiliateNetworkId > 0 {
		err = affiliate.InitAffiliateNetwork(c.AffiliateNetworkId)
		if err != nil {
			log.Errorf("[newOffer]InitAffiliateNetwork failed with %+v, err(%s)\n", c, err.Error())
			return nil
		}
	}
	o = &Offer{
		OfferConfig: c,
	}
	return
}

var gr = &http.Request{
	Method: "GET",
	URL: &url.URL{
		Path: "",
	},
}

func (o *Offer) OnLPOfferRequest(w http.ResponseWriter, req request.Request) error {
	if o == nil {
		return fmt.Errorf("Nil o for request(%s)", req.Id())
	}
	log.Infof("[Offer][OnLPOfferRequest]Offer(%s) handles request(%s)\n", o.String(), req.String())

	req.SetOfferId(o.Id)
	req.SetOptOfferId(o.Id)
	req.SetOfferName(o.Name)
	req.SetAffiliateId(o.AffiliateNetworkId)
	req.SetAffiliateName(o.AffiliateNetworkName)

	//Cap控制
	if o.CapEnabled == 1 {
		if o.CapControlled() {
			redirectOfferId := o.RedirectOfferId

			if redirectOfferId > 0 {
				req.SetOfferId(redirectOfferId)
				req.SetOptOfferId(redirectOfferId)
				req.SetOptAffiliateId(GetOffer(redirectOfferId).AffiliateNetworkId)
				ro := GetOffer(redirectOfferId)
				if ro != nil {
					return ro.OnLPOfferRequest(w, req)
				}
			}
		}
	}

	req.Redirect(w, gr, req.ParseUrlTokens(o.Url))

	return nil
}

func (o *Offer) CapControlled() bool {
	if o.CapEnabled == 1 && o.RedirectOfferId > 0 {
		impCnt := o.GetImpCount()
		log.Infof("[Offer] get imp count:%d", impCnt)
		if impCnt >= o.DailyCap {
			return true
		}
	}
	return false
}

func (o *Offer) GetImpCount() int64 {
	var impCnt int64
	svr := db.GetRedisClient(TrackCacheSvrTitle)
	if svr == nil {
		log.Errorf("[Offer]%s local Redis DB does not exist", TrackCacheSvrTitle)
		return impCnt
	}
	//key := "offer" + fmt.Sprintf("_%d", o.Id)

	key := o.GetImpKey()
	impCnt, err := svr.Get(key).Int64()

	log.Info("---GetImpCount", key, impCnt)

	if err != nil {
		log.Errorf("[Offer]%s local Redis DB does not exist", TrackCacheSvrTitle)
		return impCnt
	}
	return impCnt
}

func (o *Offer) SetImpCount() error {
	svr := db.GetRedisClient(TrackCacheSvrTitle)
	if svr == nil {
		return fmt.Errorf("[Offer]%s local Redis DB does not exist", TrackCacheSvrTitle)
	}
	//err := svr.Set(o.Id, v, localExpire).Err()
	//key := "offer" + fmt.Sprintf("_%d", o.Id)

	key := o.GetImpKey()
	err := svr.Incr(key).Err()
	if err != nil {
		log.Errorf("[Offer]set local key:%d err:%v\n", o.Id, err)
	}

	/*
		loc := GetLocation(o.CapTimezoneId)
		if loc != nil {
			now := time.Now().In(loc)
			y, m, d := now.Year(), now.Month(), now.Day()
			end := time.Date(y, m, d+1, 0, 0, 0, 0, loc)
			dur := end.Sub(now)
			svr.Expire(key, dur)
		}
	*/

	svr.Expire(key, 24*time.Hour)

	log.Infof("[Offer]set local key:%d \n", o.Id)
	return nil
}

func (o *Offer) GetImpKey() string {
	now := time.Now()

	loc := GetLocation(o.CapTimezoneId)
	if loc != nil {
		now = now.In(loc)
	}

	y, m, d := now.Year(), now.Month(), now.Day()

	key := fmt.Sprintf("offer_conversion_%d_%v%v%v", o.Id, y, m, d)
	return key
}

func (o *Offer) OnLandingPageClick(w http.ResponseWriter, req request.Request) error {
	if o == nil {
		return fmt.Errorf("Nil o for request(%s)", req.Id())
	}

	req.SetOfferId(o.Id)
	req.SetOfferName(o.Name)
	req.SetAffiliateId(o.AffiliateNetworkId)
	req.SetAffiliateName(o.AffiliateNetworkName)

	// 加载AffiliateNetwork配置，如果Append Click ID to offer URLs勾选，添加click Id(request id)
	oldId := req.Id()
	req.SetId(fmt.Sprintf("%s_%d", oldId, o.Id))

	appended := ""
	if aff := affiliate.GetAffiliateNetwork(o.AffiliateNetworkId); aff != nil && aff.AppendClickId == 1 {
		appended = req.Id()
	}

	//Cap控制
	if o.CapEnabled == 1 {
		if o.CapControlled() {
			redirectOfferId := o.RedirectOfferId

			if redirectOfferId > 0 {
				req.SetOfferId(redirectOfferId)
				req.SetOptOfferId(redirectOfferId)
				req.SetOptAffiliateId(GetOffer(redirectOfferId).AffiliateNetworkId)
				ro := GetOffer(redirectOfferId)
				if ro != nil {
					return ro.OnLandingPageClick(w, req)
				}
			}
		}
	}

	req.Redirect(w, gr, req.ParseUrlTokens(o.Url)+appended)
	req.SetId(oldId)

	return nil
}

func (o *Offer) OnImpression(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (o *Offer) OnS2SPostback(w http.ResponseWriter, req request.Request) error {
	if o == nil {
		return fmt.Errorf("Nil o for request(%s)", req.Id())
	}

	// CAP control
	if o.CapEnabled == 1 && o.RedirectOfferId > 0 {
		o.SetImpCount()
	}

	return nil
}

func (o *Offer) OnConversionPixel(w http.ResponseWriter, req request.Request) error {
	// CAP control
	if o.CapEnabled == 1 && o.RedirectOfferId > 0 {
		o.SetImpCount()
	}
	return nil
}

func (o *Offer) OnConversionScript(w http.ResponseWriter, req request.Request) error {
	// CAP control
	if o.CapEnabled == 1 && o.RedirectOfferId > 0 {
		o.SetImpCount()
	}
	return nil
}
