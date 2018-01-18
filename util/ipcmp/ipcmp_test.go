package ipcmp

import "testing"

func TestToInt64Normal(t *testing.T) {
	s := "1.2.3.4"
	n, err := IPToInt64(s)
	t.Logf("n:%v, err:%v", n, err)
	if err != nil {
		t.Errorf("parse s:%v err:%v expected nil", s, err)
	} else {
		if n != 0x01020304 {
			t.Errorf("parse s:%v expected:%v actual:%v", s, 0x01020304, n)
		}
	}
}

func TestToInt64Exception(t *testing.T) {
	s := ` 1.2.3.4      
    `
	n, err := IPToInt64(s)
	t.Logf("n:%v, err:%v", n, err)
	if err != nil {
		t.Errorf("parse s:%v err:%v expected nil", s, err)
	} else {
		if n != 0x01020304 {
			t.Errorf("parse s:%v expected:%v actual:%v", s, 0x01020304, n)
		}
	}
}

func TestToInt64Exception2(t *testing.T) {
	s := ` 1.2.3.4:2322      
    `
	n, err := IPToInt64(s)
	t.Logf("n:%v, err:%v", n, err)
	if err != nil {
		t.Errorf("parse s:%v err:%v expected nil", s, err)
	} else {
		if n != 0x01020304 {
			t.Errorf("parse s:%v expected:%v actual:%v", s, 0x01020304, n)
		}
	}
}

func TestToInt64Error1(t *testing.T) {
	s := `1.2.3.256:2322`
	n, err := IPToInt64(s)
	t.Logf("n:%v, err:%v", n, err)
	if err == nil {
		t.Errorf("parse s:%v err:%v expected not nil", s, err)
	}
}

func TestCompare(t *testing.T) {
	r, err := NewIPCompareRange("123.4.56.100", "123.4.56.10")
	if err != nil {
		t.Errorf("NewIPCompareRange failed:%v", err)
		return
	}

	inRanges := []string{
		" 123.4.56.100 ",
		" 123.4.56.99:58392",
		"123.4.56.98",
	}

	for _, in := range inRanges {
		yes, err := r.In(in)
		if err != nil {
			t.Errorf("In(%s) returns:%v", in, err)
			return
		}

		if !yes {
			t.Errorf("%s shoud in range", in)
		}
	}

	notInRanges := []string{
		"1.2.3.4",
		"123.4.56.101",
		"123.4.56.9",
	}

	for _, in := range notInRanges {
		no, err := r.In(in)
		if err != nil {
			t.Errorf("In(%s) returns:%v", in, err)
			return
		}

		if no {
			t.Errorf("%s shoud not in range", in)
		}
	}

}

func TestMinMaxIP(t *testing.T) {
	minIP := "1.0.0.0"
	maxIP := "255.255.255.255"
	min, _ := IPToInt64(minIP)
	max, _ := IPToInt64(maxIP)
	t.Log("minIP:", min)
	t.Log("maxIP:", max)
}
