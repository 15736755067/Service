package units

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"Service/common"
	"Service/log"
	"Service/request"
)

//const (
//	TrackingStepLandingPage = "lp"
//	TrackingStepImpression  = "imp"
//	TrackingStepOffer       = "offer"
//	TrackingStepPostback    = "pb"
//)

//TODO 可能的bug：如果第一步请求了offer，但是第二步请求landing page的click，这种case没有处理

func cookie(step string, req request.Request) (c *http.Cookie) {
	c = &http.Cookie{}
	defer func() {
		log.Infof("new cookie:%+v\n", *c)
	}()
	req.AddCookie("reqId", req.Id())
	switch step {
	case request.ReqImpression:
		req.AddCookie("step", request.ReqImpression)
	case request.ReqLPOffer:
		req.AddCookie("step", request.ReqLPOffer)
	case request.ReqLPClick:
		req.AddCookie("step", request.ReqLPClick)
	case request.ReqS2SPostback:
		// 应该不需要写入cookie
	default:
		return
	}
	c.Domain = req.TrackingDomain()
	// 同一用户的所有cookie共享，所以不应该限制Path
	c.Path = "/"
	c.Name = "tstep"
	c.HttpOnly = true // 客户端无法访问该Cookie
	// 关闭浏览器，就无法继续跳转到后续页面，所以Cookie失效即可
	//c.Expires = time.Now().Add(time.Hour * 1)
	c.Value = base64.URLEncoding.EncodeToString([]byte(req.CookieString()))
	return
}

func SetCookie(w http.ResponseWriter, step string, req request.Request) {
	http.SetCookie(w, cookie(step, req))
}

func ParseCookie(step string, r *http.Request) (req request.Request, err error) {
	switch step {
	case request.ReqImpression:
		return nil, fmt.Errorf("do not parse cookie in step(%s)", step)
	case request.ReqLPOffer:
	//OK
	case request.ReqLPClick:
	//OK
	case request.ReqS2SPostback: // 实际上Postback是单独解析的，现在不走这条线
	//OK
	default:
		return nil, fmt.Errorf("unsupported step(%s)", step)
	}
	c, err := r.Cookie("tstep")
	if err != nil || c == nil {
		return nil, fmt.Errorf("desired cookie(tstep) is empty with error(%v) in step(%s)", err, step)
	}
	vb, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil || len(vb) == 0 {
		return nil, fmt.Errorf("cookie(%s) decode error(%v) in step(%s)", c.Value, err, step)
	}
	vs := string(vb)
	cInfo, err := url.ParseQuery(vs)
	if err != nil || cInfo == nil {
		return nil, fmt.Errorf("cookie(%s) parseQuery error(%v) in step(%s)", c.Value, err, step)
	}

	reqId := cInfo.Get("reqId")
	es := cInfo.Get("step")
	if reqId == "" || es == "" {
		return nil, fmt.Errorf("cookie(%s) does not have valid parameters in step(%s) reqId:%s es:%s",
			c.Value, step, reqId, es)
	}
	//TODO 在线上所有的clickId，都转化为aes clickId之前，先屏蔽这个检查 2017/3/21
	/*if !common.ValidClickId(reqId) {
		return nil, fmt.Errorf("cookie(%s) does not have valid reqId in step(%s) reqId:%s es:%s",
			c.Value, step, reqId, es)
	}*/
	switch step {
	case request.ReqLPClick:
		switch es {
		case request.ReqLPOffer:
		case request.ReqLPClick:
		case request.ReqS2SPostback:
		default:
			return nil, fmt.Errorf("request step(%s) does not match last step(%s) for request(%s)",
				step, es, reqId)
		}
	case request.ReqLPOffer:
		switch es {
		case request.ReqImpression:
		default:
			return nil, fmt.Errorf("request step(%s) does not match last step(%s) for request(%s)",
				step, es, reqId)
		}
	case request.ReqS2SPostback:
		switch es {
		case request.ReqLPOffer:
		case request.ReqLPClick:
		case request.ReqS2SPostback:
		default:
			return nil, fmt.Errorf("request step(%s) does not match last step(%s) for request(%s)",
				step, es, reqId)
		}
	}
	req, err = request.CreateRequest(reqId, false, step, r)
	if req == nil || err != nil {
		return nil, fmt.Errorf("createRequest error(%v) in step(%s)", err, step)
	}

	log.Infof("[ParseCookie]Cookie(%s) for request(%s) in step(%s) with url(%s)\n",
		vs, req.Id(), step, common.SchemeHostURI(r))
	return
}
