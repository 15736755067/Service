package common

import (
	"testing"
	"time"
)

func BenchmarkParallelClickUniqueRandId(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenUniqueRandId()
		}
	})
}

func BenchmarkParallelClick(b *testing.B) {
	Init("asdf", "")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenClickId()
		}
	})
}

func BenchmarkSequentialClick(b *testing.B) {
	Init("asdf", "")
	for i := 0; i < b.N; i++ {
		clickId, _ := GenClickId()
		if !IndateClickId(clickId, 30*time.Hour) {
			b.Error(clickId)
		}
	}
}

func TestClick(t *testing.T) {
	Init("", "")
	for i := 0; i < 10000; i++ {
		clickId, _ := GenClickId()
		if !IndateClickId(clickId, 30*time.Hour) {
			t.Log(clickId)
			t.Fail()
		}
	}
}

//func TestClick2(t *testing.T) {
//	Init("", "")
//	for i := 0; i < 3; i++ {
//		clickId, _ := GenClickId()
//		time.Sleep(time.Second)
//		if IndateClickId(clickId, 30*time.Millisecond) {
//			t.Log(clickId)
//			t.Fail()
//		}
//	}
//}

func BenchmarkParallelClick4(b *testing.B) {
	Init("asdf", "")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			clickId, s := GenClickId()
			if !IndateClickId(clickId, 30*time.Hour) {
				b.Error(clickId)
			}
			pc := PlainClickId(clickId)
			if pc != s {
				b.Error(s, clickId, pc)
			}
		}
	})
}
func BenchmarkSequentialClick4(b *testing.B) {
	Init("asdf", "")
	for i := 0; i < b.N; i++ {
		clickId, s := GenClickId()
		if !IndateClickId(clickId, 30*time.Hour) {
			b.Error(clickId)
		}
		pc := PlainClickId(clickId)
		if pc != s {
			b.Error(s, clickId, pc)
		}
	}
}

func TestPlainClick(t *testing.T) {
	Init("", "")
	clickId := "e6712381ff83c0fe61b50d9eca979b3032"
	t.Log(TimeOfClickId(clickId))
}
