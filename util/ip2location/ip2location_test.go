package ip2location

import "testing"

func TestGet_all(t *testing.T) {
	Open("/Users/robin/Program/gopath/src/AdClickService/src/Service/DB19SAMPLE.BIN")
	ip := "23.23.11.23"
	location := Get_all(ip)
	t.Logf("%#v", location)
}
