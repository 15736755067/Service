package lander

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"Service/log"
	"Service/request"
)

type LanderConfig struct {
	Id             int64
	Name           string
	UserId         int64
	Url            string
	NumberOfOffers int64
}

func (c LanderConfig) String() string {
	return fmt.Sprintf("Lander %d:%d", c.Id, c.UserId)
}

func (c LanderConfig) ID() int64 {
	return c.Id
}

type Lander struct {
	LanderConfig
}

var cmu sync.RWMutex // protects the following
var landers = make(map[int64]*Lander)

func setLander(l *Lander) error {
	if l == nil {
		return errors.New("setLander error:l is nil")
	}
	if l.Id <= 0 {
		return fmt.Errorf("setLander error:l.Id(%d) is not positive", l.Id)
	}
	cmu.Lock()
	defer cmu.Unlock()
	landers[l.Id] = l
	return nil
}

func getLander(lId int64) *Lander {
	cmu.RLock()
	defer cmu.RUnlock()
	return landers[lId]
}

func delLander(lId int64) {
	cmu.Lock()
	defer cmu.Unlock()
	delete(landers, lId)
}

func InitLander(landerId int64) error {
	l := newLander(DBGetLander(landerId))
	if l == nil {
		return fmt.Errorf("newLander failed with lander(%d)", landerId)
	}
	return setLander(l)
}

func GetLander(landerId int64) (l *Lander) {
	if landerId == 0 {
		return nil
	}

	l = getLander(landerId)
	if l == nil {
		l = newLander(DBGetLander(landerId))
		if l != nil {
			if err := setLander(l); err != nil {
				return nil
			}
		}
	}
	return
}

func newLander(c LanderConfig) (l *Lander) {
	log.Debugf("[newLander]%+v\n", c)
	if c.Id <= 0 {
		return nil
	}
	_, err := url.ParseRequestURI(c.Url)
	if err != nil {
		log.Errorf("[NewLander]Invalid url for lander(%+v), err(%s)\n", c, err.Error())
		// not returning nil, because exists the following case
		// http://apps.coach.lightshines.top/276466/31307=%subid1%: invalid URL escape "%su"
		// return nil
	}
	l = &Lander{
		LanderConfig: c,
	}
	return
}

var gr = &http.Request{
	Method: "GET",
	URL: &url.URL{
		Path: "",
	},
}

func (l *Lander) OnLPOfferRequest(w http.ResponseWriter, req request.Request) error {
	if l == nil {
		return fmt.Errorf("Nil l for request(%s)", req.Id())
	}
	log.Infof("[Lander][OnLPOfferRequest]Lander(%s) handles request(%s)\n", l.String(), req.String())

	req.SetLanderId(l.Id)
	req.SetLanderName(l.Name)
	req.Redirect(w, gr, req.ParseUrlTokens(l.Url))
	return nil
}

// Lander不需要处理该请求
//func (l *Lander) OnLandingPageClick(w http.ResponseWriter, req request.Request) error {
//	return nil
//}

func (l *Lander) OnImpression(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (l *Lander) OnS2SPostback(w http.ResponseWriter, req request.Request) error {
	// lander不做特别处理应该也可以，但是接口先留下来
	return nil
}

func (l *Lander) OnConversionPixel(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (l *Lander) OnConversionScript(w http.ResponseWriter, req request.Request) error {
	return nil
}
