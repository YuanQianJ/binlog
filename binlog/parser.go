package binlog

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/schema"
	jsoniter "github.com/json-iterator/go"
)

type BinlogParser struct {
	columnTag string
	onceMap   map[string]*tableSchema
}
type tableSchema struct {
	once        sync.Once
	columnIdMap map[string]int
}

func (g *BinlogParser) GetBinLogData(element any, e *RowsEvent, n int) error {
	value := reflect.ValueOf(element).Elem()
	num := value.NumField()
	t := value.Type()
	for k := 0; k < num; k++ {
		columnName := t.Field(k).Tag.Get(g.columnTag)
		columnId := g.GetBinlogIdByName(e, columnName)
		switch value.Field(k).Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8:
			value.Field(k).SetInt(g.IntSetter(e.RowsEvent, n, columnId))
		case reflect.String:
			value.Field(k).SetString(g.StringSetter(e.RowsEvent, n, columnId))
		case reflect.Float32, reflect.Float64:
			value.Field(k).SetFloat(g.FloatSetter(e.RowsEvent, n, columnId))
		case reflect.Bool:
			value.Field(k).SetBool(g.BoolSetter(e.RowsEvent, n, columnId))
		default:
			newObject := reflect.New(value.Field(k).Type()).Interface()
			json := g.StringSetter(e.RowsEvent, n, columnId)
			err := jsoniter.Unmarshal([]byte(json), &newObject)
			if err != nil {
				return err
			}
			value.Field(k).Set(reflect.ValueOf(newObject).Elem().Convert(value.Field(k).Type()))
		}
	}
	return nil
}

func (m *BinlogParser) IntSetter(e *canal.RowsEvent, n int, columnId int) int64 {

	if e.Table.Columns[columnId].Type != schema.TYPE_NUMBER {
		return 0
	}

	switch v := e.Rows[n][columnId].(type) {
	case int8:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return e.Rows[n][columnId].(int64)
	case int:
		return int64(e.Rows[n][columnId].(int))
	case uint8:
		return int64(e.Rows[n][columnId].(uint8))
	case uint16:
		return int64(e.Rows[n][columnId].(uint16))
	case uint32:
		return int64(e.Rows[n][columnId].(uint32))
	case uint64:
		return int64(e.Rows[n][columnId].(uint64))
	case uint:
		return int64(e.Rows[n][columnId].(uint))
	}
	return 0
}
func (m *BinlogParser) FloatSetter(e *canal.RowsEvent, n int, columnId int) float64 {

	if e.Table.Columns[columnId].Type != schema.TYPE_FLOAT {
		panic("Not float type")
	}

	switch e.Rows[n][columnId].(type) {
	case float32:
		return float64(e.Rows[n][columnId].(float32))
	case float64:
		return float64(e.Rows[n][columnId].(float64))
	}
	return float64(0)
}

func (m *BinlogParser) StringSetter(e *canal.RowsEvent, n int, columnId int) string {

	if e.Table.Columns[columnId].Type == schema.TYPE_ENUM {

		values := e.Table.Columns[columnId].EnumValues
		if len(values) == 0 {
			return ""
		}
		if e.Rows[n][columnId] == nil {
			return ""
		}

		return values[e.Rows[n][columnId].(int64)-1]
	}

	value := e.Rows[n][columnId]

	switch value := value.(type) {
	case []byte:
		return string(value)
	case string:
		return value
	}
	return ""
}

func (m *BinlogParser) BoolSetter(e *canal.RowsEvent, n int, columnId int) bool {

	val := m.IntSetter(e, n, columnId)
	return val == 1
}

func (m *BinlogParser) GetBinlogIdByName(e *RowsEvent, columnName string) int {
	val, ok := m.onceMap[e.tableKey]
	if !ok {
		panic("No table schema:" + e.tableKey)
	}
	if val.columnIdMap == nil {
		val.once.Do(func() {
			val.columnIdMap = make(map[string]int, len(e.Table.Columns))
			for id, value := range e.Table.Columns {
				val.columnIdMap[value.Name] = id
			}
		})
	}
	id, ok := val.columnIdMap[columnName]
	if !ok {
		panic(fmt.Sprintf("There is no column %s in table %s.%s", columnName, e.Table.Schema, e.Table.Name))
	}
	return id
}
func (m *BinlogParser) registerOnce(key string) {
	m.onceMap[key] = &tableSchema{
		once: sync.Once{},
	}
}
