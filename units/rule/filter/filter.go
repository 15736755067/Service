package filter

import (
	"Service/log"
	"Service/request"
)

type Filter interface {
	Accept(request.Request) bool
	Marshal() string
}

func NewFilter(jsonConfig string) (Filter, error) {
	log.Debugf("Create new filter with %s\n", jsonConfig)
	fi := new(filterImpl)
	err := fi.Fill([]byte(jsonConfig))
	return fi, err
}
