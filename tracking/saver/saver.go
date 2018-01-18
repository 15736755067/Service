// Package saver 用于统一存储key, value结构的mysql表。
// 采用Reflect实现
package saver

import (
	"Service/log"
	"database/sql"
	"sync"
)

// Saver 执行具体保存任务
// 异步执行
type Saver struct {
	tasks chan map[interface{}]interface{} // 要保存的数据
	sql   string
}

// NewSaver 创建一个新的保存器。
// Running协程不会自动启动，要靠外面启动
func NewSaver(bufferSize int, sql string) *Saver {
	return &Saver{
		tasks: make(chan map[interface{}]interface{}, bufferSize),
		sql:   sql,
	}
}

// Save 异步保存数据库
func (s *Saver) Save(data map[interface{}]interface{}) error {
	s.tasks <- data
	return nil
}

func (s *Saver) doSave(db *sql.DB, data map[interface{}]interface{}) error {
	// 也就是想办法拼出来这样的语句
	// INSERT INTO table (key1, key2, key3, field1, field2, field3) VALUES (key1, key2, key3, field1, field2, field3)
	// ON DUPLICATE KEY UPDATE fields1+=fields1, fields2+=fields2, fields3+=fields3
	// 也可以先在外面把SQL外进来
	// 然后提供把struct的字段给拼接成[]interface{}
	// 然后使用...即可

	if len(data) == 0 {
		return nil
	}

	// 提交Prepare可以避免重复解析SQL语句
	stmt, err := db.Prepare(s.sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	w := sync.WaitGroup{}
	for k, v := range data {
		w.Add(1)
		go func(k, v interface{}) {
			defer func() {
				w.Done()
				if x := recover(); x != nil {
					log.Error("[tracking][Saver][doSave] insert panic:", x)
				}
			}()
			var args Args
			args = args.AddFlatValues(k)
			args = args.AddFlatValues(v)
			args = args.AddFlatValues(v)
			_, err := stmt.Exec(args...)
			if err != nil {
				log.Errorf("Exec sql:%s with args:%+v failed:%v", s.sql, args, err)
			}
		}(k, v)
	}
	w.Wait()

	return nil
}

// Running 存储协程
func (s *Saver) Running(db *sql.DB, stop chan struct{}) {
	for {
		select {
		case m := <-s.tasks:
			s.doSave(db, m)
		case <-stop:
			// 收所有的数据，防止的未写入数据库的
			for {
				select {
				case m := <-s.tasks:
					s.doSave(db, m)
				default:
					goto allreceived
				}
			}
		allreceived:
			return
		}
	}
}
