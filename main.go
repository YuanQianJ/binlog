package main

import (
	"binlog/binlog"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type MyTableSchema struct {
	Title       string `db:"title"`
	Brand       string `db:"brand"`
	Price       string `db:"price"`
	ImageURL    string `db:"imageURL"`
	Category    string `db:"category"`
	Description string `db:"description"`
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
func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	config := &binlog.Config{
		Addr:     "127.0.0.1:3306",
		User:     "root",
		Password: "000000",
	}

	listener, err := binlog.NewBinlogListener(
		config,
		rdb,
		"binglog:pos",
	)
	if err != nil {
		panic(err)
	}

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
}
