package request

import (
	"net/http"

	"Service/common"
	"Service/log"
)

type S2SPostbackRequest struct {
	reqbase
}

func CreateS2SPostbackRequest(reqId string, r *http.Request) Request {
	breq, err := getReqCache(reqId, false)
	if err != nil || breq == nil {
		log.Errorf("[CreateS2SPostbackRequest]Failed with reqId(%s) from %s with err(%v)\n",
			reqId, common.SchemeHostURI(r), err)
		return nil
	}

	breq.t = ReqS2SPostback
	breq.trackingPath = r.URL.Path

	return &S2SPostbackRequest{*breq}
}
