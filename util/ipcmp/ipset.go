package ipcmp

// IPSet 保存所有的IP
type IPSet struct {
	m map[IP_INT]string
}

// AddIP 添加一个IP和它对应的说明
func (s *IPSet) AddIP(ip IP_INT, desc string) {
	s.m[ip] = desc
}

// IPIn 判断一个IP是否在集合中
// 如果在，返回其desc和true
// 如果不在，返回""和false
func (s *IPSet) IPIn(ip IP_INT) (string, bool) {
	desc, ok := s.m[ip]
	return desc, ok
}

// AddrIn 返回192.168.0.155:23234是否在IP列表中
func (s *IPSet) AddrIn(addr string) (string, bool) {
	ip, err := IPToInt64(addr)
	if err != nil {
		return "", false
	}

	return s.IPIn(ip)
}

// NewIPSet make一个新的IPSet
func NewIPSet() IPSet {
	return IPSet{make(map[IP_INT]string)}
}

// Count 返回有多少个IP
func (s *IPSet) Count() int {
	return len(s.m)
}
