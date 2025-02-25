package binlog

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/go-redis/redis/v8"
)

type EventHandler interface {
	DbName() string
	TableName() string
	OnUpdate(data ...UpdateHandler)
	OnDelete(data ...any)
	OnInsert(data ...any)
	Schema() any
}

type BinlogHandler struct {
	canal.DummyEventHandler
	BinlogParser
	eventMap map[string]EventHandler
	config   *Config
	canalCli *canal.Canal
	errors   chan error
	running  bool
}

type Config struct {
	Addr     string
	User     string
	Password string

	ColumnTag string

	PosHandler PositionHandler
}

type RowsEvent struct {
	*canal.RowsEvent
	tableKey string
}

type UpdateHandler struct {
	Old any
	New any
}

func NewBinlogListener(config *Config, redisCli *redis.Client, redisKey string) (*BinlogHandler, error) {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = config.Addr
	cfg.User = config.User
	cfg.Password = config.Password
	cfg.Dump.ExecutionPath = ""
	c, err := canal.NewCanal(cfg)
	if err != nil {
		return nil, err
	}
	if config.ColumnTag == "" {
		config.ColumnTag = "db"
	}
	if config.PosHandler == nil {
		posHandler, err := NewRedisPosHandler(redisCli, redisKey)
		if err != nil {
			return nil, err
		}
		config.PosHandler = posHandler
	}
	listener := &BinlogHandler{
		BinlogParser: BinlogParser{
			columnTag: config.ColumnTag,
			onceMap:   make(map[string]*tableSchema, 16),
		},
		canalCli: c,
		errors:   make(chan error, 1),
		config:   config,
		eventMap: make(map[string]EventHandler, 16),
	}
	listener.canalCli.SetEventHandler(listener)
	return listener, nil
}

func (b *BinlogHandler) RegisterEventHandler(e EventHandler) {
	if b.running {
		panic("can not register event handler after Run")
	}
	value := reflect.ValueOf(e.Schema())
	kind := value.Kind()
	if kind != reflect.Ptr {
		panic(fmt.Errorf("expected struct pointer, got %s", kind.String()))
	}
	key := e.DbName() + "." + e.TableName()
	b.eventMap[key] = e
	b.registerOnce(key)
}

func (b *BinlogHandler) OnRow(e *canal.RowsEvent) error {
	var err error
	defer func() {
		//recover() 用于捕获 panic，如果panic，recover() 会返回 panic 的值
		if panic := recover(); panic != nil {
			b.ErrorHandler(errors.New("panic: " + fmt.Sprint(panic)))
		}
		if err != nil {
			b.ErrorHandler(err)
		}
	}()
	event := &RowsEvent{
		RowsEvent: e,
		tableKey:  e.Table.Schema + "." + e.Table.Name,
	}
	handler, ok := b.eventMap[event.tableKey]
	if !ok {
		return nil
	}
	var n = 0
	var step = 1
	var inserts, deletes []any
	var updateHandlers []UpdateHandler
	if e.Action == canal.UpdateAction {
		n = 1
		step = 2
	} else {
		inserts = make([]any, 0, len(e.Rows))
		deletes = make([]any, 0, len(e.Rows))
	}
	for i := n; i < len(e.Rows); i += step {
		data := handler.Schema()
		err = b.GetBinLogData(data, event, i)
		if err != nil {
			return err
		}
		switch e.Action {
		case canal.UpdateAction:
			oldData := handler.Schema()
			err = b.GetBinLogData(oldData, event, i-1)
			if err != nil {
				return err
			}
			updateHandlers = append(updateHandlers, UpdateHandler{
				Old: oldData,
				New: data,
			})
		case canal.InsertAction:
			inserts = append(inserts, data)
		case canal.DeleteAction:
			deletes = append(deletes, data)
		default:
			err = errors.New("unknown action of onRow")
			return err
		}
	}
	if len(updateHandlers) > 0 {
		handler.OnUpdate(updateHandlers...)
		return nil
	}
	if len(inserts) > 0 {
		handler.OnInsert(inserts...)
		return nil
	}
	if len(deletes) > 0 {
		handler.OnDelete(deletes...)
		return nil
	}
	return nil
}

func (h *BinlogHandler) String() string {
	return "BinlogHandler"
}

func (b *BinlogHandler) OnPosSynced(header *replication.EventHeader, pos mysql.Position, _ mysql.GTIDSet, _ bool) error {
	fmt.Printf("Binlog position changed: %s:%d\n", pos.Name, pos.Pos)
	err := b.config.PosHandler.UpdatePos(pos)
	if err != nil {
		b.ErrorHandler(err)
	}
	return err
}

func (b *BinlogHandler) OnRotate(header *replication.EventHeader, event *replication.RotateEvent) error {
	err := b.config.PosHandler.UpdatePos(mysql.Position{
		Pos:  uint32(event.Position),
		Name: string(event.NextLogName),
	})
	if err != nil {
		b.ErrorHandler(err)
	}
	return err
}
func (b *BinlogHandler) Run() error {
	masterPos, err := b.canalCli.GetMasterPos()
	if err != nil {
		b.ErrorHandler(err)
	}
	pos, err := b.config.PosHandler.GetLatestPos()
	if err != nil {
		b.ErrorHandler(err)
	}
	if pos.Name != "" {
		return b.canalCli.RunFrom(pos)
	}
	b.running = true
	return b.canalCli.RunFrom(masterPos)
}

func (b *BinlogHandler) ErrorHandler(err error) {
	select {
	case b.errors <- err:
	default:
	}
}

func (b *BinlogHandler) Errors() <-chan error {
	return b.errors
}
