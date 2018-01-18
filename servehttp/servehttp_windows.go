//+build windows

package servehttp

import (
	"errors"
	"fmt"
	"net/http"
)

// Serve 使用ListenAndServe启动HTTP服务
func Serve(servers ...*http.Server) error {
	type indexError struct {
		index int
		err   error
	}

	errs := make(chan indexError, len(servers))

	for i, s := range servers {
		go func(idx int, s *http.Server) {
			err := s.ListenAndServe()
			errs <- indexError{
				index: idx,
				err:   err,
			}
		}(i, s)
	}

	errString := ""
	for range servers {
		idxErr := <-errs
		errInfo := fmt.Sprintf("[%dth server error:%v]", idxErr.index, idxErr.err)
		errString += errInfo
	}

	return errors.New(errString)
}
