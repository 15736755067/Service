// Package gather 提供统一的数据汇总模型
// 注册key只能使用结构体或者值，不能使用指针
package gather

import (
	"time"
)

// Saver 用于把gather到的数据给存储起来
type Saver interface {
	Save(data map[interface{}]interface{}) error
}

// Gather 就是外部要使用的
type Gather struct {
	data map[interface{}]interface{}

	// 通过这个收集事件
	GatherChan chan Event

	// 当值不存在的时候创建一个新的值
	NewValue func() interface{}

	saver Saver

	saveInterval time.Duration // 多长时间存储一次
}

// Event 一条修改请求
type Event struct {
	Key    interface{}
	Action func(v interface{})
}

func (g *Gather) flush() {
	g.saver.Save(g.data)
	g.data = make(map[interface{}]interface{})
}

// Gathering 汇总协和
func (g *Gather) Gathering(stop chan struct{}) {
	ticker := time.NewTicker(g.saveInterval)
	for {
		select {
		case a := <-g.GatherChan:
			d := g.getValue(a.Key)
			a.Action(d)

		case <-ticker.C:
			g.flush()

		case <-stop:
			// 把已经有的收完
			for {
				select {
				case a := <-g.GatherChan:
					d := g.getValue(a.Key)
					a.Action(d)
				default:
					// 没有多余的数据了
					goto allreceived
				}
			}
		allreceived:
			g.flush()
			return
		}
	}
}

// getValue 只能在action里面调用。
func (g *Gather) getValue(key interface{}) interface{} {
	v, ok := g.data[key]
	if ok {
		return v
	}

	v = g.NewValue()
	g.data[key] = v
	return v
}

// NewGather 创建一个新的汇总器，但是并不启动其协程。
// 外面负责启动其协程
func NewGather(bufferSize int, valueNewer func() interface{}, saver Saver, interval time.Duration) *Gather {
	return &Gather{
		data:         make(map[interface{}]interface{}),
		GatherChan:   make(chan Event, bufferSize),
		NewValue:     valueNewer,
		saver:        saver,
		saveInterval: interval,
	}
}
