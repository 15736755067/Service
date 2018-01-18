package affiliate

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"

	"Service/log"
)

// AffiliateNetworkConfig 对应AffiliateNetwork表
type AffiliateNetworkConfig struct {
	Id                int64
	UserId            int64
	Name              string
	Hash              string
	PostbackUrl       string
	AppendClickId     int
	DuplicatePostback int
	IpWhiteList       []string
}

func (c *AffiliateNetworkConfig) OnlyAcceptFromWhiteList() bool {
	return len(c.IpWhiteList) != 0
}

// Allow 返回是否接受这个ip的postback请求
func (c *AffiliateNetworkConfig) Allow(ip string) bool {
	if !c.OnlyAcceptFromWhiteList() {
		return true
	}

	for _, whiteIP := range c.IpWhiteList {
		if whiteIP == ip {
			return true
		}
	}

	return false
}

func (c AffiliateNetworkConfig) String() string {
	return fmt.Sprintf("affiliateNetwork Id=%d:UserId=%d", c.Id, c.UserId)
}

type AffiliateNetwork struct {
	AffiliateNetworkConfig
}

var cmu sync.RWMutex // protects the following
var affs = make(map[int64]*AffiliateNetwork)

func setAffiliateNetwork(o *AffiliateNetwork) error {
	if o == nil {
		return errors.New("setAffiliateNetwork error:o is nil")
	}
	if o.Id <= 0 {
		return fmt.Errorf("setAffiliateNetwork error:o.Id(%d) is not positive", o.Id)
	}
	cmu.Lock()
	defer cmu.Unlock()
	affs[o.Id] = o
	return nil
}
func getAffiliateNetwork(oId int64) *AffiliateNetwork {
	cmu.RLock()
	defer cmu.RUnlock()
	return affs[oId]
}
func delAffiliateNetwork(oId int64) {
	cmu.Lock()
	defer cmu.Unlock()
	delete(affs, oId)
}

func InitAffiliateNetwork(affId int64) error {
	o := getAffiliateNetwork(affId)
	if o == nil {
		o = newAffiliateNetwork(DBGetAffiliateNetwork(affId))
	}
	if o == nil {
		return fmt.Errorf("[InitAffiliateNetwork]Failed because newAffiliateNetwork failed with aff(%d)", affId)
	}
	return setAffiliateNetwork(o)
}

func GetAffiliateNetwork(affId int64) (o *AffiliateNetwork) {
	if affId == 0 {
		return nil
	}

	o = getAffiliateNetwork(affId)
	if o == nil {
		o = newAffiliateNetwork(DBGetAffiliateNetwork(affId))
	}
	if o != nil {
		if err := setAffiliateNetwork(o); err != nil {
			return nil
		}
	}
	return
}

func newAffiliateNetwork(c AffiliateNetworkConfig) (o *AffiliateNetwork) {
	if c.Id <= 0 {
		return nil
	}
	if c.PostbackUrl != "" {
		_, err := url.ParseRequestURI(c.PostbackUrl)
		if err != nil {
			log.Errorf("[newAffiliateNetwork]Invalid url for affiliateNetwork(%+v), err(%s)\n", c, err.Error())
			return nil
		}
	}
	for _, addr := range c.IpWhiteList {
		if net.ParseIP(addr) == nil {
			log.Errorf("[newAffiliateNetwork]Invalid ip for affiliateNetwork(%+v)\n", c)
			return nil
		}
	}
	o = &AffiliateNetwork{
		AffiliateNetworkConfig: c,
	}
	return
}
