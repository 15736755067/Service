package blacklist

import (
	"Service/util/ipcmp"
	"bufio"
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"
)

// GlobalBlacklist 全局的block列表
type GlobalBlacklist struct {
	set  ipcmp.IPSet
	lock sync.RWMutex
}

var ipLineParser = regexp.MustCompile(`([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)[\s]*#[\s]*(.*)`)

// ErrLineFormatError 格式错误
var ErrLineFormatError = errors.New("Format error")

func (g *GlobalBlacklist) addLine(line string) error {
	line = strings.TrimSpace(line)
	// 1.0.0.4				 # 2013-04-02, 1.0.0.4, AUS, 50
	l := ipLineParser.FindStringSubmatch(line)
	if len(l) == 0 {
		return ErrLineFormatError
	}

	if l[0] != line {
		return ErrLineFormatError
	}

	ips, desc := l[1], l[2]
	ipn, _ := ipcmp.IPToInt64(ips)
	g.set.AddIP(ipn, desc)
	return nil
}

// Reload reloads blacklist file
func (g *GlobalBlacklist) Reload(path string) error {
	lock.Lock()
	defer lock.Unlock()

	// 打开配置文件，如果失败，则内部老数据仍然保留
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// 清空列表
	g.set = ipcmp.NewIPSet()

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')

		g.addLine(line)

		if err != nil {
			break
		}
	}
	return nil
}

// IPIn 判断一个IP是否在集合中
// 如果在，返回其desc和true
// 如果不在，返回""和false
func (g *GlobalBlacklist) IPIn(ip ipcmp.IP_INT) (string, bool) {
	lock.RLock()
	defer lock.RUnlock()

	return g.set.IPIn(ip)
}

// AddrIn 返回192.168.0.155:23234是否在IP列表中
func (g *GlobalBlacklist) AddrIn(addr string) (string, bool) {
	lock.RLock()
	defer lock.RUnlock()

	return g.set.AddrIn(addr)
}

// New as is
func New() *GlobalBlacklist {
	return &GlobalBlacklist{
		set: ipcmp.NewIPSet(),
	}
}

// IGlobalBlacklist 定义GlobalBlacklist接口
type IGlobalBlacklist interface {
	IPIn(ip ipcmp.IP_INT) (string, bool)
	AddrIn(addr string) (string, bool)
}

// DisabledGlobalBlacklist 没有功能的Blacklist
type DisabledGlobalBlacklist struct {
}

// IPIn 直接返回false
func (d DisabledGlobalBlacklist) IPIn(ip ipcmp.IP_INT) (string, bool) {
	return "", false
}

// AddrIn 直接返回false
func (d DisabledGlobalBlacklist) AddrIn(addr string) (string, bool) {
	return "", false
}

// G 方便外部使用的全局变量
// 默认是禁用的
var G IGlobalBlacklist = DisabledGlobalBlacklist{}

// EnableBlacklist 启用blacklist
func EnableBlacklist(path string) error {
	var g = New()
	if err := g.Reload(path); err != nil {
		return err
	}

	G = g
	return nil
}

// DisableBlacklist 禁用blacklist
func DisableBlacklist() {
	G = DisabledGlobalBlacklist{}
}
