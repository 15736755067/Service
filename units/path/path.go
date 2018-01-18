package path

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"Service/log"
	"Service/request"
	"Service/units/lander"
	"Service/units/offer"
)

const (
	StatusPaused  = 0
	StatusRunning = 1
)

type PathConfig struct {
	Id           int64
	UserId       int64
	RedirectMode int64
	DirectLink   int64
	Status       int64
}

func (c PathConfig) String() string {
	return fmt.Sprintf("Path %d:%d Status %d", c.Id, c.UserId, c.Status)
}
func (c PathConfig) ID() int64 {
	return c.Id
}

func (c PathConfig) Detail() string {
	return fmt.Sprintf("%#+v", c)
}

type PathLander struct {
	LanderId int64
	Weight   uint64
}

func (pl PathLander) ID() int64 {
	return pl.LanderId
}

type PathOffer struct {
	OfferId int64
	Weight  uint64
}

func (po PathOffer) ID() int64 {
	return po.OfferId
}

type Path struct {
	PathConfig
	landers []PathLander
	lwSum   uint64 // lander总权重
	offers  []PathOffer
	owSum   uint64 // offer总权重
}

func (p *Path) RandOwSum() int {
	if p.owSum == 0 {
		return 0
	}
	return rand.Intn(int(p.owSum))
}

func (p *Path) RandLwSum() int {
	if p.lwSum == 0 {
		return 0
	}
	return rand.Intn(int(p.lwSum))
}

var cmu sync.RWMutex // protects the following
var paths = make(map[int64]*Path)

func setPath(p *Path) error {
	if p == nil {
		return errors.New("setPath error:p is nil")
	}
	if p.Id <= 0 {
		return fmt.Errorf("setPath error:p.Id(%d) is not positive", p.Id)
	}
	cmu.Lock()
	defer cmu.Unlock()
	paths[p.Id] = p
	return nil
}
func getPath(pId int64) *Path {
	cmu.RLock()
	defer cmu.RUnlock()
	return paths[pId]
}
func delPath(pId int64) {
	cmu.Lock()
	defer cmu.Unlock()
	delete(paths, pId)
}

func InitPath(pathId int64) error {
	p := newPath(DBGetPath(pathId))
	if p == nil {
		return fmt.Errorf("newPath failed with path(%d)", pathId)
	}
	return setPath(p)
}

func GetPath(pathId int64) (p *Path) {
	if pathId == 0 {
		return nil
	}

	p = getPath(pathId)
	if p == nil {
		p = newPath(DBGetPath(pathId))
		if p != nil {
			if err := setPath(p); err != nil {
				return nil
			}
		}
	}
	return
}

func newPath(c PathConfig) (p *Path) {
	log.Debugf("[newPath]%s\n", c.Detail())
	var lwSum, owSum uint64
	landers := DBGetPathLanders(c.Id)
	for _, c := range landers {
		if lander.GetLander(c.LanderId) == nil {
			log.Errorf("[NewPath]GetLander failed with %+v\n", c)
			return nil
		}
		lwSum += c.Weight
	}
	offers := DBGetPathOffers(c.Id)
	for _, c := range offers {
		if offer.GetOffer(c.OfferId) == nil {
			log.Errorf("[NewPath]GetOffer failed with %+v\n", c)
			return nil
		}
		owSum += c.Weight
	}

	if owSum == 0 {
		log.Errorf("path:%v have %v offers and owSum=%v", c.Id, len(offers), owSum)
	}

	p = &Path{
		PathConfig: c,
		landers:    landers,
		offers:     offers,
		lwSum:      lwSum,
		owSum:      owSum,
	}

	return
}

// return value: >0 if valid;0 if invalid
func (p *Path) RandLanderId(reqId string) (id int64) {
	x := p.RandLwSum() // rand.Intn(int(p.lwSum))
	lx := 0
	for _, l := range p.landers {
		if l.LanderId <= 0 {
			continue
		}
		lx += int(l.Weight)
		if x < lx {
			return l.LanderId
		}
	}
	log.Errorf("[Path][RandLanderId]Request(%s) does not match any lander(%d:%d) in path(%d)",
		reqId, lx, x, p.Id)
	return 0
}

// return value: >0 if valid;0 if invalid
func (p *Path) RandOfferId(reqId string) (id int64) {
	y := p.RandOwSum() // rand.Intn(int(p.owSum))
	oy := 0
	for _, o := range p.offers {
		if o.OfferId <= 0 {
			continue
		}
		oy += int(o.Weight)
		if y < oy {
			return o.OfferId
		}
	}
	log.Errorf("[Path][RandOfferId]Request(%s) does not match any offer(%d:%d) in path(%d)",
		reqId, oy, y, p.Id)
	return 0
}

func (p *Path) OnLPOfferRequest(w http.ResponseWriter, req request.Request) error {
	if p == nil {
		return fmt.Errorf("Nil p for request(%s)", req.Id())
	}
	log.Infof("[Path][OnLPOfferRequest]Path(%s) handles request(%s)\n", p.String(), req.String())

	req.SetRedirectMode(p.RedirectMode)

	if p.DirectLink == 0 {
		landerId := p.RandLanderId(req.Id())
		if landerId > 0 {
			req.SetLanderId(landerId)
			// set optional offer id & affiliate network id
			offerId := p.RandOfferId(req.Id())
			req.SetOptOfferId(offerId)
			if offer.GetOffer(offerId) != nil {
				// do not panic here
				req.SetOptAffiliateId(offer.GetOffer(offerId).AffiliateNetworkId)
			}
			return lander.GetLander(landerId).OnLPOfferRequest(w, req)
		}
	}

	offerId := p.RandOfferId(req.Id())
	if offerId > 0 {
		req.SetOfferId(offerId)
		req.SetOptOfferId(offerId)
		req.SetOptAffiliateId(offer.GetOffer(offerId).AffiliateNetworkId)
		return offer.GetOffer(offerId).OnLPOfferRequest(w, req)
	}

	return fmt.Errorf(
		"[Path][OnLPOfferRequest]Request(%s) does not match any lander or offer in path(%d)",
		req.Id(), p.Id)
}

func (p *Path) OnLandingPageClick(w http.ResponseWriter, req request.Request) error {
	if p == nil {
		return fmt.Errorf("Nil p for request(%s)", req.Id())
	}

	req.SetRedirectMode(p.RedirectMode)

	// 不需要find，因为可能中途已被移除
	/*
		found := false
		for _, l := range p.landers {
			if l.LanderId == req.LanderId() {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Target Lander(%d) not found for request(%s) in path(%d)",
				req.LanderId(), req.Id(), p.Id)
		}
	*/

	pp := strings.Split(strings.TrimRight(req.TrackingPath(), "/"), "/")
	switch len(pp) {
	case 2: // path为/click或/click/，按照权重选择一个Offer
		offerId := p.RandOfferId(req.Id())
		if offerId > 0 {
			o := offer.GetOffer(offerId)
			if o != nil {
				req.SetOfferId(offerId)
				return o.OnLandingPageClick(w, req)
			}
		}
	case 3: // path为/click/N，按照指定顺序(1~)选择一个Offer
		i, err := strconv.ParseInt(pp[2], 10, 64)
		if err != nil || i == 0 || i > int64(len(p.offers)) {
			return fmt.Errorf("Target offer path(%s)(i:%d) parse failed err(%v) for request(%s) in path(%d)(offers:%d)",
				req.TrackingPath(), i, err, req.Id(), p.Id, len(p.offers))
		}
		//TODO 是否需要加上lander的NumberOfOffers的检查？
		if p.offers[i-1].OfferId > 0 {
			o := offer.GetOffer(p.offers[i-1].OfferId)
			if o != nil {
				req.SetOfferId(p.offers[i-1].OfferId)
				return o.OnLandingPageClick(w, req)
			}
		}
	}

	return fmt.Errorf("Target offer path(%s) not found for request(%s) in path(%d)",
		req.TrackingPath(), req.Id(), p.Id)
}

func (p *Path) OnImpression(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (p *Path) OnS2SPostback(w http.ResponseWriter, req request.Request) error {
	if p == nil {
		return fmt.Errorf("Nil p for request(%s)", req.Id())
	}

	// 不需要find，因为可能中途已被移除
	if req.LanderId() != 0 {
		l := lander.GetLander(req.LanderId())
		if l != nil {
			// 不一定肯定存在Lander
			l.OnS2SPostback(w, req)
		}
	}

	o := offer.GetOffer(req.OfferId())
	if o != nil {
		// 但是Offer是一定存在的
		return o.OnS2SPostback(w, req)
	}

	return fmt.Errorf("Target offer id(%d) not found for request(%s) in path(%d)",
		req.OfferId(), req.Id(), p.Id)
}

func (p *Path) OnConversionPixel(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (p *Path) OnConversionScript(w http.ResponseWriter, req request.Request) error {
	return nil
}
