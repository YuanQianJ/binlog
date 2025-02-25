# MySQL Binlog Listener

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/mysql-binlog-listener)](https://goreportcard.com/report/github.com/yourusername/mysql-binlog-listener)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight MySQL binlog listener with Redis position persistence, written in Go.

## 📌 Features

- ✅ Real-time monitoring of INSERT/UPDATE/DELETE events
- 🚀 Automatic schema mapping with struct tags
- 🔄 Redis-based position recovery
- 🛠 Extensible position storage interface
- 🧩 Built-in error handling with channel
- 🔌 Easy integration with existing systems

## 📦 Installation

```bash
go get github.com/YuanQianJ/binlog/tree/master/binlog
```
## 🚀 Quick Start
# 1. Implement Event Handler
```
type MyTableSchema struct {
	Id       string `db:"id"`
	User       string `db:"user"`
}

type MyEventHandler struct{}

func (h *MyEventHandler) DbName() string {
	return "cd_vinyl"
}

func (h *MyEventHandler) TableName() string {
	return "product_test"
}

func (h *MyEventHandler) OnUpdate(data ...binlog.UpdateHandler) {
	for _, d := range data {
		oldSchema := d.Old.(*MyTableSchema)
		newSchema := d.New.(*MyTableSchema)
		fmt.Printf("[UPDATE] Old: %+v → New: %+v\n", oldSchema, newSchema)
	}
}

func (h *MyEventHandler) OnDelete(data ...any) {
	for _, d := range data {
		schema := d.(*MyTableSchema)
		fmt.Printf("[DELETE] %+v\n", schema)
	}
}

func (h *MyEventHandler) OnInsert(data ...any) {
	for _, d := range data {
		schema := d.(*MyTableSchema)
		fmt.Printf("[INSERT] %+v\n", schema)
	}
}

func (h *MyEventHandler) Schema() any {
	return &MyTableSchema{}
}
```
# 2. Initialize Listener
```
	rdb := redis.NewClient(&redis.Options{
		Addr: "",
	})

	config := &binlog.Config{
		Addr:     "",
		User:     "",
		Password: "",
	}

	listener, err := binlog.NewBinlogListener(
		config,
		rdb,
		"binglog:pos",
	)
	if err != nil {
		panic(err)
	}
```
# 3. Register Handlers & Run
'''
	handler := &MyEventHandler{}
	listener.RegisterEventHandler(handler)

	go func() {
		for err := range listener.Errors() {
			fmt.Println("Error:", err)
		}
	}()
	if err := listener.Run(); err != nil {
		panic(err)
	}
```

