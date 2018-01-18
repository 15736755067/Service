package gracequit

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// working 一直运行，直到收到停止信号
// 收到停止信号的时候，把自己的索引给出去
// 方便外部判断顺序
func working(t *testing.T, c StopSigChan, name string, index int, to chan int) {
	ticker := time.NewTicker(time.Duration(1+rand.Float64()*2) * time.Second)

	for {
		select {
		case <-ticker.C:
			t.Logf("%s: %s", name, "working...")
		case <-c:
			t.Logf("%s: %s", name, "stop signale received")
			time.Sleep(time.Duration(rand.Float64()*2) * time.Second)
			to <- index
			return
		}
	}
}

func TestGraceQuit(t *testing.T) {
	to := make(chan int, 10)
	for i := 0; i < 10; i++ {
		func(index int) {
			StartGoroutine(func(c StopSigChan) {
				working(t, c, fmt.Sprintf("worker.%d", index), index, to)
			})
		}(i)
	}

	t.Log("all started")

	go func() {
		StopAll()
	}()

	for i := 0; i < 10; i++ {
		index := <-to
		if index != 10-i-1 {
			t.Errorf("index:%v expected:%v", index, 10-i-1)
		}
		t.Logf("index:%v stopped", index)
	}
	t.Log("all stopped")
}
