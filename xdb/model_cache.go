package xdb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"

	"github.com/daodao97/xgo/xlog"

	"github.com/pkg/errors"

	cache2 "github.com/daodao97/xgo/cache"
)

func (m *model) cacheKeyPrefix(id string) string {
	return fmt.Sprintf("%s-%s-%s", m.connection, m.table, id)
}

func (m *model) FindById(id string) (Record, error) {
	row := m.FindBy(id)
	if row.Err != nil {
		return nil, row.Err
	}
	return row.Data, nil
}

func (m *model) FindBy(id string) *Row {
	if cache == nil {
		return &Row{Err: errors.New("cache instance is nil")}
	}

	pk := m.PrimaryKey()
	if pk == "" {
		return &Row{Err: errors.New("primary is not defined")}
	}

	key := m.cacheKeyPrefix(id)

	c, err := cache.Get(context.Background(), key)
	if err != nil && !errors.Is(err, cache2.ErrNotFound) {
		return &Row{Err: err}
	}
	if c != "" {
		var result map[string]interface{}
		err = json.Unmarshal([]byte(c), &result)
		if err != nil {
			return &Row{Err: err}
		}
		xlog.Debug("FindBy id:", xlog.Any("id", id), xlog.Any("cache", c))
		return &Row{Data: result}
	}

	row := m.SelectOne(WhereEq(pk, id))
	if row.Err == nil && row.Data != nil {
		c, err := json.Marshal(row.Data)
		if err != nil {
			xlog.Error("json marshal after FindBy id", xlog.Any("id", id), xlog.Err(err))
			return row
		}
		err = cache.Set(context.Background(), key, string(c))
		if err != nil {
			xlog.Error("set key after FindBy id", xlog.Any("id", id), xlog.Err(err))
		} else {
			xlog.Debug("set key after FindBy id", xlog.Any("id", id))
		}
	}

	return row
}

func (m *model) UpdateBy(id string, record Record) (bool, error) {
	if cache == nil {
		return false, errors.New("cache instance is nil")
	}
	_, err := m.Update(record, WhereEq("id", id))
	if err != nil {
		return false, err
	}
	key := m.cacheKeyPrefix(id)
	err = cache.Del(context.Background(), key)
	if err != nil {
		xlog.Error("del key after UpdateBy id", xlog.Any("id", id), xlog.Err(err))
	} else {
		xlog.Debug("del key after UpdateBy id", xlog.Any("id", id))
	}

	return true, nil
}

func (m *model) FindByField(field string, val string) (Record, error) {
	row := m.FindByKey(field, val)
	if row.Err != nil {
		return nil, row.Err
	}
	return row.Data, nil
}

func (m *model) FindByKey(key string, val string) *Row {
	if cache == nil {
		return &Row{Err: errors.New("cache instance is nil")}
	}

	pk := m.PrimaryKey()
	if pk == "" {
		return &Row{Err: errors.New("primary is not defined")}
	}

	cacheKey := m.cacheKeyPrefix(val)

	c, err := cache.Get(context.Background(), cacheKey)
	if err != nil && !errors.Is(err, cache2.ErrNotFound) {
		return &Row{Err: err}
	}

	if c != "" {
		xlog.Debug("FindBy key:", xlog.Any("key", key), xlog.Any("val", val), xlog.Any("cache", c))
		return m.FindBy(c)
	}

	row := m.SelectOne(WhereEq(key, val), Field(m.primaryKey))

	if row.Err == nil && row.Data != nil {
		row = m.FindBy(cast.ToString(row.Data[m.primaryKey]))
		if row.Err != nil {
			return row
		}
		cacheKey := row.GetString(key)
		err = cache.Set(context.Background(), m.cacheKeyPrefix(cacheKey), row.GetString(m.primaryKey))
		if err != nil {
			xlog.Error("set key after FindByKey", xlog.Any("key", cacheKey), xlog.Err(err))
		} else {
			xlog.Debug("set key after FindByKey", xlog.Any("key", cacheKey))
		}
	}

	return row
}
