package blacklist

import (
	"Service/util/ipcmp"
	"testing"
)

func TestParseIpRanges(t *testing.T) {
	j := `["1.2.3.4", "1.2.3.5 - 1.2.3.100", " 9.10.11.12", "1.2.3.200 - 1.2.3.100"]`
	ranges, err := parseIpRanges(j)
	if err != nil {
		t.Errorf("parseIpRanges failed:%v", err)
		return
	}
	t.Logf("%+v", ranges)
}

type ipUserAgent struct {
	ip string
	ua string
}

// TestBlacklistConfig 测试单个配置是不是工作正常
func TestBlacklistConfig(t *testing.T) {
	ipRange := `["1.2.3.4", "1.2.3.90 - 1.2.3.100", " 9.10.11.12", "255.255.255.100 - 255.255.255.20"]`
	userAgent := `["Mozilla/5.0", "Chrome/55"]`
	c, err := BuildBlacklistConfig(1, 2, ipRange, userAgent, 1)
	if err != nil {
		t.Errorf("BuildBlacklistConfig failed:%v", err)
		return
	}

	t.Logf("BuildBlacklistConfig: %+v", c)

	notAllowed := []ipUserAgent{
		{"1.2.3.4", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36"},
		{"1.2.3.90", "Mozilla/5.0"},
		{"255.255.255.21", "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55"},
		{"255.255.255.99", "Chrome/55"},
	}

	for _, r := range notAllowed {
		ipInt, _ := ipcmp.IPToInt64(r.ip)
		if c.Allowed(ipInt, r.ua) {
			t.Errorf("r:%+v should not allowed", r)
		}
	}

	allowed := []ipUserAgent{
		{"1.2.3.4", "Mozilla/6.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2883.95 Safari/537.36"},
		{"1.2.3.89", "Mozilla/5.0"},
		{"255.255.255.21", "AppleWebKit/537.36 (KHTML, like Gecko)"},
	}

	for _, r := range allowed {
		ipInt, _ := ipcmp.IPToInt64(r.ip)
		if !c.Allowed(ipInt, r.ua) {
			t.Errorf("r:%+v should allowed", r)
		}
	}

}

// TestBlacklistConfig 测试单个配置是不是工作正常
func TestBlacklistConfig2(t *testing.T) {
	ipRange := `["1.2.3.4", "1.2.3.90 - 1.2.3.100", " 9.10.11.12", "255.255.255.100 - 255.255.255.20"]`
	userAgent := ``
	c, err := BuildBlacklistConfig(1, 2, ipRange, userAgent, 1)
	if err != nil {
		t.Errorf("BuildBlacklistConfig failed:%v", err)
		return
	}

	t.Logf("BuildBlacklistConfig: %+v", c)

	notAllowed := []ipUserAgent{
		{"1.2.3.4", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36"},
		{"1.2.3.90", "Mozilla/5.0"},
		{"1.2.3.90", ""},
		{"255.255.255.21", "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55"},
		{"255.255.255.99", "Chrome/55"},
	}

	for _, r := range notAllowed {
		ipInt, _ := ipcmp.IPToInt64(r.ip)
		if c.Allowed(ipInt, r.ua) {
			t.Errorf("r:%+v should not allowed", r)
		}
	}

	allowed := []ipUserAgent{
		{"1.2.3.5", "Mozilla/6.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2883.95 Safari/537.36"},
		{"1.2.3.89", "Mozilla/5.0"},
		{"255.255.255.19", "AppleWebKit/537.36 (KHTML, like Gecko)"},
	}

	for _, r := range allowed {
		ipInt, _ := ipcmp.IPToInt64(r.ip)
		if !c.Allowed(ipInt, r.ua) {
			t.Errorf("r:%+v should allowed", r)
		}
	}

}
