package request

import (
	"net/http"
)

type LPOfferRequest struct {
	reqbase
}

func CreateLPOfferRequest(reqId string, r *http.Request, brandNew bool) (req *LPOfferRequest) {
	var breq *reqbase
	if !brandNew { // 不是全新的Request，则先尝试去cache中获取
		breq, err := getReqCache(reqId, true)
		if err == nil && breq != nil {
			breq.t = ReqLPOffer
			breq.trackingPath = r.URL.Path
			return &LPOfferRequest{*breq}
		}
	}

	breq = newReqBase(reqId, ReqLPOffer, r)
	if breq == nil {
		return nil
	}

	return &LPOfferRequest{*breq}
}
