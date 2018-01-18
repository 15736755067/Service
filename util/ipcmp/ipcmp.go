package ipcmp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type IP_INT int64

// IPCompare 用于ip的解析和对比
type IPCompare struct {
	ip1 IP_INT
	ip2 IP_INT
}

// In 判断ip是不是在这个范围内
func (cmp IPCompare) In(ip string) (bool, error) {
	n, err := IPToInt64(ip)
	if err != nil {
		return false, err
	}

	return n >= cmp.ip1 && n <= cmp.ip2, nil
}

// IpIntIn 先转换成IP_INT再进行比较可以提高速度
func (cmp IPCompare) IpIntIn(n IP_INT) bool {
	return n >= cmp.ip1 && n <= cmp.ip2
}

// NewIPCompare 从123.45.6.7构建出一个IPCompare
func NewIPCompare(ip string) (IPCompare, error) {
	n, err := IPToInt64(ip)
	if err != nil {
		return IPCompare{}, err
	}

	return IPCompare{
		ip1: n,
		ip2: n,
	}, nil
}

// NewIPCompareRange 123.45.6.7 - 123.45.6.10 构建出一个IPCompare
func NewIPCompareRange(ip1, ip2 string) (IPCompare, error) {
	n1, err := IPToInt64(ip1)
	if err != nil {
		return IPCompare{}, err
	}

	n2, err := IPToInt64(ip2)
	if err != nil {
		return IPCompare{}, err
	}

	if n1 > n2 {
		n1, n2 = n2, n1
	}

	return IPCompare{
		ip1: n1,
		ip2: n2,
	}, nil
}

// ErrNotIPFormat IP格式不正确
var ErrNotIPFormat = errors.New("Not IP Format")

// IPToInt64 从192.168.0.1解析出一个整数
// 支持从 192.168.0.1:63323解析出IP部分
func IPToInt64(s string) (IP_INT, error) {
	s = strings.TrimSpace(s)
	nums := strings.Split(s, ".")
	if len(nums) < 4 {
		return 0, ErrNotIPFormat
	}
	var n IP_INT
	for i := 0; i < 4; i++ {
		if i == 3 {
			// 为了支持1.2.3.4:29383
			// 忽略数字后面的东西
			pos := strings.IndexFunc(nums[i], func(c rune) bool {
				return !(c >= '0' && c <= '9')
			})
			if pos != -1 {
				nums[i] = nums[i][:pos]
			}
		}

		c, err := strconv.Atoi(nums[i])
		if err != nil {
			return 0, ErrNotIPFormat
		}
		if c < 0 || c > 255 {
			return 0, ErrNotIPFormat
		}
		n = n<<8 + IP_INT(c)
	}
	return n, nil
}

func Int64ToIP(ip int64) (string, error) {
	if ip < 16777216 || ip > 4294967295 {
		return "", fmt.Errorf("ip(%d) out of range", ip)
	}
	s_ip := strconv.FormatInt(ip>>24, 10) + "."
	s_ip = s_ip + strconv.FormatInt((ip&0x00ffffff)>>16, 10) + "."
	s_ip = s_ip + strconv.FormatInt((ip&0x0000ffff)>>8, 10) + "."
	s_ip = s_ip + strconv.FormatInt(ip&0x000000ff, 10)
	return s_ip, nil
}
