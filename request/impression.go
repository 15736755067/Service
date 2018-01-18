package request

import (
	"net/http"
)

// ImpressionRequest for impression tracking
type ImpressionRequest struct {
	reqbase
}

// CreateImpressionRequest creates am impression request
func CreateImpressionRequest(id string, r *http.Request) *ImpressionRequest {
	breq := newReqBase(id, ReqImpression, r)
	if breq == nil {
		return nil
	}

	return &ImpressionRequest{*breq}
}
