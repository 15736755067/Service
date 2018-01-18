package common

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	UrlTokenClickId       = "cid"
	UrlTokenPayout        = "payout"
	UrlTokenTransactionId = "txid"
)

//http header
const (
	KHttpContentType            = "Content-Type"
	KHttpContentTypeProto       = "application/x-protobuf"
	KHttpContentTypeOctetStream = "application/octet-stream"
	KHttpContentTypeJson        = "application/json"
	KHttpContentTypeJsonp       = "application/javascript"
	KHttpContentLength          = "Content-Length"
	KHttpContentEncode          = "Content-Encoding"
	KHttpPost                   = "POST"
	KHttpGet                    = "GET"
)

// TrafficSourceParams {"Parameter":"X","Placeholder":"X","Name":"X","Track":N(0,1)}
type TrafficSourceParams struct {
	Parameter   string
	Placeholder string
	Name        string
	Track       int64
}

func (p TrafficSourceParams) Encode() string {
	return fmt.Sprintf("%s:%s:%s:%d",
		p.Name,
		p.Parameter,
		p.Placeholder,
		p.Track)
}

func (p *TrafficSourceParams) Decode(str string) {
	if str == "" {
		return
	}
	s := strings.Split(str, ":")
	if len(s) >= 4 {
		p.Parameter = s[0]
		p.Placeholder = s[1]
		p.Name = s[2]
		p.Track, _ = strconv.ParseInt(s[3], 10, 64)
	}
}

func EncodeParams(params []TrafficSourceParams) (str string) {
	for _, p := range params {
		str = str + ";" + p.Encode()
	}
	return
}

func DecodeParams(str string) (params []TrafficSourceParams) {
	for _, s := range strings.Split(str, ";") {
		var p TrafficSourceParams
		p.Decode(s)
		params = append(params, p)
	}
	return
}

func UnionOf(s1, s2 []int64) (sd []int64) {
	sd = make([]int64, 0, len(s1)+len(s2))
	sm := make(map[int64]bool)
	smax, smin := s1, s2
	if len(s2) > len(s1) {
		smax, smin = s2, s1
	}
	for _, i := range smax {
		sm[i] = true
		sd = append(sd, i)
	}
	for _, j := range smin {
		if !sm[j] {
			sd = append(sd, j)
		}
	}
	return
}

func RemoveDuplicates(s1 []string) (s2 []string) {
	if len(s1) == 0 {
		return
	}
	sm := make(map[string]bool, len(s1))
	for _, s := range s1 {
		sm[s] = true
	}
	s2 = make([]string, 0, len(sm))
	for s := range sm {
		s2 = append(s2, s)
	}
	return
}
