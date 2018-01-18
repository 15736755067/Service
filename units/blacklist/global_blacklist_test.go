package blacklist

import (
	"Service/util/ipcmp"
	"testing"
)

func TestGlobalBlackList(t *testing.T) {
	var g GlobalBlacklist
	g.set = ipcmp.NewIPSet()

	err := g.addLine("1.1.169.109			 # 2017-01-18, node-86l.pool-1-1.dynamic.totbb.net, THA, 1")
	if err != nil {
		t.Errorf("addLine failed:%v", err)
		return
	}

	// path := "/Users/zhanchenxing/Downloads/full_blacklist_database-2.txt"

	// // 打开配置文件，如果失败，则内部老数据仍然保留
	// f, err := os.Open(path)
	// if err != nil {
	// 	t.Errorf("Open %s failed", path)
	// 	return
	// }
	// defer f.Close()

	// // 清空列表
	// g.set = ipcmp.NewIPSet()

	// reader := bufio.NewReader(f)
	// for {
	// 	line, err := reader.ReadString('\n')
	// 	t.Logf("line: %s", line)

	// 	if err := g.addLine(line); err != nil {
	// 		fmt.Println(line)
	// 	}

	// 	if err != nil {
	// 		break
	// 	}
	// }

	err = g.Reload("/Users/zhanchenxing/Downloads/full_blacklist_database-2.txt")
	if err != nil {
		t.Errorf("Reload failed:%v", err)
		return
	}

	count := g.set.Count()
	t.Logf("count=%v", count)

	desc, in := g.AddrIn("1.0.0.4")
	if !in {
		t.Errorf("1.0.0.4 should in")
		return
	}
	t.Logf("1.0.0.4 in: %s", desc)

	ins := []string{
		"1.9.171.51",
		"54.221.33.228",
		"54.221.33.228:63323",
		"54.221.45.156",
	}

	for _, in := range ins {
		desc, in := g.AddrIn(in)
		if !in {
			t.Errorf("%s should in", in)
			return
		}
		t.Logf("%v in: %v", in, desc)
	}

	notIns := []string{
		"1.90.171.51",
		"4.221.33.228",
		"4.221.33.228:63323",
		"4.221.45.156",
	}

	for _, in := range notIns {
		_, in := g.AddrIn(in)
		if in {
			t.Errorf("%s should not in", in)
			return
		}
	}

}
