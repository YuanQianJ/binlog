package binlog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-redis/redis/v8"
)

type PositionHandler interface {
	UpdatePos(pos mysql.Position) error
	GetLatestPos() (mysql.Position, error)
}
type RedisPosHandler struct {
	redisCli *redis.Client
	key      string
	ctx      context.Context
}

func NewRedisPosHandler(redisCli *redis.Client, key string) (*RedisPosHandler, error) {
	return &RedisPosHandler{
		redisCli: redisCli,
		key:      key,
		ctx:      context.Background(),
	}, nil
}

func (r *RedisPosHandler) UpdatePos(pos mysql.Position) error {
	posJson, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	fmt.Printf("Updating binlog position in Redis: %s:%d\n", pos.Name, pos.Pos)
	return r.redisCli.Set(r.ctx, r.key, posJson, 0).Err()
}

func (r *RedisPosHandler) GetLatestPos() (mysql.Position, error) {
	posJson, err := r.redisCli.Get(r.ctx, r.key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 如果 key 不存在，返回空位置
			fmt.Println("not exist")
			return mysql.Position{}, nil
		}
		return mysql.Position{}, err
	}
	var pos mysql.Position
	err = json.Unmarshal(posJson, &pos)
	if err != nil {
		return mysql.Position{}, err
	}
	fmt.Printf("Binlog position from Redis: %s:%d\n", pos.Name, pos.Pos)
	return pos, err
}
