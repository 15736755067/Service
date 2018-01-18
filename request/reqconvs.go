package request

import (
	"net/http"

	"Service/common"
	"Service/log"
)

type UploadConversionsRequest struct {
	reqbase
}

func CreateUploadConversionsRequest(reqId string, r *http.Request) Request {
	breq, err := getReqCache(reqId, false)
	if err != nil || breq == nil {
		log.Errorf("[CreateUploadConversionsRequest]Failed with reqId(%s) from %s with err(%v)\n",
			reqId, common.SchemeHostURI(r), err)
		return nil
	}

	breq.t = ReqUploadConversions
	breq.trackingPath = r.URL.Path

	return &UploadConversionsRequest{*breq}
}
