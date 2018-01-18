package request

import (
	"net/http"

	"Service/common"
	"Service/log"
)

type LPClickRequest struct {
	reqbase
}

func CreateLPClickRequest(reqId string, r *http.Request) Request {
	breq, err := getReqCache(reqId, true)
	if err != nil || breq == nil {
		log.Errorf("[CreateLPClickRequest]Failed with reqId(%s) from %s with err(%v)\n",
			reqId, common.SchemeHostURI(r), err)
		return nil
	}

	breq.t = ReqLPClick
	breq.trackingPath = r.URL.Path

	return &LPClickRequest{*breq}
}
