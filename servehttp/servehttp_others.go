//+build !windows

package servehttp

import (
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
)

// Serve 使用gracehttp启动HTTP服务
func Serve(servers ...*http.Server) error {
	return gracehttp.Serve(servers...)
}
