package rule

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"Service/log"
	"Service/request"
	"Service/units/path"
	"Service/units/rule/filter"
)

const (
	StatusPaused  = 0
	StatusRunning = 1
)

type RuleConfig struct {
	Id     int64
	UserId int64
	Json   string
	Status int64

	Type int
}

func (c RuleConfig) String() string {
	return fmt.Sprintf("Rule %d:%d Status %d", c.Id, c.UserId, c.Status)
}
func (c RuleConfig) Detail() string {
	return fmt.Sprintf("Rule %d:%d Status %d Json %s", c.Id, c.UserId, c.Status, c.Json)
}
func (c RuleConfig) ID() int64 {
	return c.Id
}

const (
	RulePathStatusPaused  = 0
	RulePathStatusRunning = 1
)

type RulePath struct {
	PathId int64
	Weight uint64
	Status int64
}

func (rp RulePath) ID() int64 {
	return rp.PathId
}

type Rule struct {
	RuleConfig
	f     filter.Filter
	paths []RulePath
	pwSum uint64
}

// // Rand returns, as an int, a non-negative pseudo-random number in [0,n)
func (r *Rule) Rand() int {
	if r.pwSum <= 0 {
		return 0
	}
	return rand.Intn(int(r.pwSum))
}

var cmu sync.RWMutex // protects the following
var rules = make(map[int64]*Rule)

func setRule(r *Rule) error {
	if r == nil {
		return errors.New("setRule error:r is nil")
	}
	if r.Id <= 0 {
		return fmt.Errorf("setRule error:r.Id(%d) is not positive", r.Id)
	}
	cmu.Lock()
	defer cmu.Unlock()
	rules[r.Id] = r
	return nil
}
func getRule(rId int64) *Rule {
	cmu.RLock()
	defer cmu.RUnlock()
	return rules[rId]
}
func delRule(rId int64) {
	cmu.Lock()
	defer cmu.Unlock()
	delete(rules, rId)
}

func InitRule(ruleId int64) error {
	r := newRule(DBGetRule(ruleId))
	if r == nil {
		return fmt.Errorf("newRule failed with rule(%d)", ruleId)
	}
	return setRule(r)
}

func newRule(c RuleConfig) (r *Rule) {
	log.Debugf("[newRule]%+v\n", c)
	f, err := filter.NewFilter(c.Json)
	if err != nil || f == nil {
		log.Errorf("[newRule]NewFilter failed for rule(%+v) with err(%v)\n", c, err)
		return nil
	}
	var pwSum uint64
	paths := DBGetRulePaths(c.Id)
	if len(paths) == 0 {
		log.Errorf("rule:%d have no paths", c.Id)
	}
	for _, p := range paths {
		if p.Status != RulePathStatusRunning {
			continue
		}
		pwSum += p.Weight
		if path.GetPath(p.PathId) == nil {
			log.Errorf("[newRule]GetPath failed for p%d:r%d\n", p.PathId, c.Id)
			return nil
		}
	}
	r = &Rule{
		RuleConfig: c,
		f:          f,
		paths:      paths,
		pwSum:      pwSum,
	}
	return
}

func GetRule(ruleId int64) (r *Rule) {
	if ruleId == 0 {
		return nil
	}

	r = getRule(ruleId)
	if r == nil {
		r = newRule(DBGetRule(ruleId))
		if r != nil {
			if err := setRule(r); err != nil {
				return nil
			}
		}
	}
	return
}

func (r *Rule) Accept(req request.Request) bool {
	if r == nil {
		return false
	}
	if r.f == nil {
		return false
	}
	return r.f.Accept(req)
}

func (r *Rule) OnLPOfferRequest(w http.ResponseWriter, req request.Request) error {
	if r == nil {
		return fmt.Errorf("Nil r for request(%s)", req.Id())
	}
	log.Infof("[Rule][OnLPOfferRequest]Rule(%s) handles request(%s)\n", r.String(), req.String())
	if !r.Accept(req) {
		return fmt.Errorf("Request(%s) not accepted by rule(%d)", req.Id(), r.Id)
	}
	x := r.Rand()
	cx := 0
	for _, p := range r.paths {
		if p.PathId <= 0 {
			continue
		}
		if p.Status != path.StatusRunning {
			continue
		}
		cx += int(p.Weight)
		if x < cx {
			req.SetPathId(p.PathId)
			return path.GetPath(p.PathId).OnLPOfferRequest(w, req)
		}
	}
	return fmt.Errorf("Request(%s) does not match any path(%d:%d) in rule(%d)", req.Id(), cx, x, r.Id)
}

func (r *Rule) OnLandingPageClick(w http.ResponseWriter, req request.Request) error {
	if r == nil {
		return fmt.Errorf("Nil r for request(%s)", req.Id())
	}

	// 不需要find，因为可能中途已被移除
	/*
		for _, p := range r.paths {
			if p.PathId == req.PathId() {
				return path.GetPath(req.PathId()).OnLandingPageClick(w, req)
			}
		}
	*/

	p := path.GetPath(req.PathId())
	if p != nil {
		return p.OnLandingPageClick(w, req)
	}

	return fmt.Errorf("Target Path(%d) not found for request(%s) in rule(%d)", req.PathId(), req.Id(), r.Id)
}

func (r *Rule) OnImpression(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (r *Rule) OnS2SPostback(w http.ResponseWriter, req request.Request) error {
	if r == nil {
		return fmt.Errorf("Nil r for request(%s)", req.Id())
	}

	// 不需要find，因为可能中途已被移除
	p := path.GetPath(req.PathId())
	if p != nil {
		return p.OnS2SPostback(w, req)
	}

	return fmt.Errorf("Target Path(%d) not found for request(%s) in rule(%d)", req.PathId(), req.Id(), r.Id)
}

func (r *Rule) OnConversionPixel(w http.ResponseWriter, req request.Request) error {
	return nil
}

func (r *Rule) OnConversionScript(w http.ResponseWriter, req request.Request) error {
	return nil
}
