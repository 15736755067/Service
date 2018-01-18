package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"Service/util/simpleaes"
)

var uniqueTitle = "AABB" // 4位
const uniqueLen = 4
const originalClickIdLen = 16
const clickIdLen = (originalClickIdLen + 1) * 2 // 34位

var cipherKey = "e4e0c909c39dbee1f55507b1"
var aesCipher *simpleaes.Aes

func Init(unique string, cipherStr string) {
	var err error
	if unique != "" {
		uniqueTitle = unique
	} else {
		// init with mac address
		ifs, err := net.Interfaces()
		if err != nil {
			panic(err.Error())
		}

		if len(ifs) > 0 {
			macaddr := strings.Replace(ifs[0].HardwareAddr.String(), ":", "", -1)
			if len(macaddr) >= uniqueLen {
				uniqueTitle = macaddr
			}
		}
	}

	if len(uniqueTitle) > uniqueLen {
		uniqueTitle = uniqueTitle[len(uniqueTitle)-uniqueLen:] // 太长则取后几位
	}

	if len(uniqueTitle) < uniqueLen {
		uniqueTitle += randString(uniqueLen - len(uniqueTitle)) // 太短则随机补齐
	}

	if cipherStr != "" {
		aesCipher, err = simpleaes.New(16, cipherStr)
	} else {
		aesCipher, err = simpleaes.New(16, cipherKey)
	}

	if err != nil {
		panic(err.Error())
	}
}

func GetUerIdText(r *http.Request) string {
	if r == nil {
		return ""
	}
	return strings.Split(r.Host, ".")[0]
}

func GetCampaignHash(r *http.Request) string {
	if r == nil {
		return ""
	}
	s := strings.Split(r.URL.Path, "/")
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}

func HostPath(r *http.Request) string {
	return r.Host + r.RequestURI
}

func SchemeHost(r *http.Request) string {
	scheme := r.URL.Scheme
	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}
	return scheme + "://" + r.Host
}

func SchemeHostPath(r *http.Request) string {
	scheme := r.URL.Scheme
	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}
	return scheme + "://" + r.Host + r.URL.Path
}

func SchemeHostURI(r *http.Request) string {
	scheme := r.URL.Scheme
	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}
	return scheme + "://" + r.Host + r.RequestURI
}

func HostWithoutPort(r *http.Request) string {
	return strings.Split(r.Host, ":")[0]
}

func GenUniqueRandId() string {
	s := fmt.Sprintf("%s%d%s", uniqueTitle, time.Now().UnixNano(), randString(6))

	md5h := md5.New()
	md5h.Write([]byte(s))
	cipherStr := md5h.Sum(nil)

	return hex.EncodeToString(cipherStr)
}

// 自定义的系统开始时间，用于减少timestamp位数
var startTime time.Time = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)

func SystemUnix() int64 {
	return int64(time.Now().Sub(startTime).Seconds())
}

func System2UTCUnix(seconds int64) int64 {
	return startTime.Add(time.Second * time.Duration(seconds)).Unix()
}

func TimeOfClickId(clickId string) (t time.Time) {
	plainClickId := PlainClickId(clickId)
	if plainClickId == "" {
		return
	}
	seconds, err := strconv.ParseInt(strings.TrimPrefix(string(plainClickId[4:12]), "0"), 16, 64)
	if err != nil {
		return
	}
	return startTime.Add(time.Second * time.Duration(seconds))
}

func GenClickId() (id, s string) {
	// 16位:ttttFFFFCCCCxxxx
	s = fmt.Sprintf("%s%0.8X%s", uniqueTitle, SystemUnix(), randString(4))
	dst := aesCipher.Encrypt([]byte(s))
	if len(dst) == 0 {
		return "", s
	}

	cb := []byte(fmt.Sprintf("%X", crc32.ChecksumIEEE(dst)))
	id = hex.EncodeToString(append(dst, cb[len(cb)-1]))
	return
}

func ValidClickId(clickId string) bool {
	if len(clickId) != clickIdLen {
		// 长度是否正确
		return false
	}
	rb, err := hex.DecodeString(clickId)
	if err != nil {
		// 是否能正确初步解码
		return false
	}
	checkSum := []byte(fmt.Sprintf("%X", crc32.ChecksumIEEE(rb[:len(rb)-1])))
	if checkSum[len(checkSum)-1] != rb[len(rb)-1] {
		// 校验码是否正确
		return false
	}
	return true
}

func PlainClickId(clickId string) (s string) {
	if len(clickId) != clickIdLen {
		// 长度是否正确
		return ""
	}
	rb, err := hex.DecodeString(clickId)
	if err != nil {
		// 是否能正确初步解码
		return ""
	}
	checkSum := []byte(fmt.Sprintf("%X", crc32.ChecksumIEEE(rb[:len(rb)-1])))
	if checkSum[len(checkSum)-1] != rb[len(rb)-1] {
		// 校验码是否正确
		return ""
	}
	src := aesCipher.Decrypt(rb[:len(rb)-1])
	if len(src) != originalClickIdLen {
		// 解码出来长度是否正确
		return ""
	}
	return string(src)
}

func IndateClickId(clickId string, duration time.Duration) bool {
	if len(clickId) != clickIdLen {
		// 长度是否正确
		return false
	}
	rb, err := hex.DecodeString(clickId)
	if err != nil {
		// 是否能正确初步解码
		return false
	}
	checkSum := []byte(fmt.Sprintf("%X", crc32.ChecksumIEEE(rb[:len(rb)-1])))
	if checkSum[len(checkSum)-1] != rb[len(rb)-1] {
		// 校验码是否正确
		return false
	}
	src := aesCipher.Decrypt(rb[:len(rb)-1])
	if len(src) != originalClickIdLen {
		// 解码出来长度是否正确
		return false
	}
	ts, err := strconv.ParseInt(strings.TrimPrefix(string(src[4:12]), "0"), 16, 64)
	if err != nil {
		return false
	}
	return ts+int64(duration.Seconds()) >= SystemUnix()
}

func GenUUID(n uint, params ...string) (rs string) {
	s := randString(4)
	for _, param := range params {
		s += param
	}
	md5h := md5.New()
	md5h.Write([]byte(s))
	cipherByte := md5h.Sum(nil)

	rs = hex.EncodeToString(cipherByte)
	if len(rs) < 32 {
		rs += randString(32 - len(rs))
	}
	return rs[:8] + "-" + rs[8:16] + "-" + rs[16:24] + "-" + rs[24:32]
}

func randString(n int) string {
	var letterBytes = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func WritePidFile() {
	pidFile := "server.pid"
	f, err := os.OpenFile(pidFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic("open pid file error")
	}
	f.WriteString(fmt.Sprintf("%d", os.Getpid()))
	f.Close()
}
