package xdb

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	Name            string
	DSN             string
	ReadDsn         string
	Driver          string
	MaxOpenConn     int
	MaxIdleConn     int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

var pool = sync.Map{}

type DbPool struct {
	db   *sql.DB
	conf *Config
}

func Inits(conns []Config) error {
	var connMap = make(map[string]*Config)
	for _, conf := range conns {
		_conf := conf
		if conf.Name == "" {
			return errors.New("connection name is empty")
		}
		connMap[conf.Name] = &_conf
	}
	return Init(connMap)
}

func Init(conns map[string]*Config) error {
	for conn, conf := range conns {
		db, err := NewDb(conf)
		if err != nil {
			return err
		}
		pool.Store(conn, db)
		if conf.ReadDsn != "" {
			rdb, err := NewDb(&Config{
				DSN:             conf.ReadDsn,
				Driver:          conf.Driver,
				MaxOpenConn:     conf.MaxOpenConn,
				MaxIdleConn:     conf.MaxIdleConn,
				ConnMaxIdleTime: conf.ConnMaxIdleTime,
				ConnMaxLifetime: conf.ConnMaxLifetime,
			})
			if err != nil {
				return err
			}
			pool.Store(readConn(conn), rdb)
		}
	}
	return nil
}

func Close() {
	pool.Range(func(key, value any) bool {
		db := value.(*sql.DB)
		_ = db.Close()
		return true
	})
}

func readConn(conn string) string {
	return conn + "_read"
}

func db(conn string) (*DbPool, error) {
	if _db, ok := pool.Load(conn); ok {
		return _db.(*DbPool), nil
	}
	return nil, errors.New("connection not found : " + conn)
}

func DB(conn string) (*sql.DB, error) {
	_db, err := db(conn)
	if err != nil {
		return nil, err
	}
	return _db.db, nil
}

func NewDb(conf *Config) (*DbPool, error) {
	driver := conf.Driver
	if driver == "" {
		driver = "mysql"
	}

	db, err := sql.Open(driver, conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed Connection database: %s", err)
	}

	// 设置数据库连接池最大连接数
	MaxOpen := 100
	if conf.MaxOpenConn != 0 {
		MaxOpen = conf.MaxOpenConn
	}
	db.SetMaxOpenConns(MaxOpen)

	// 连接池最大允许的空闲连接数
	// 如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭
	MaxIdle := 20
	if conf.MaxIdleConn != 0 {
		MaxIdle = conf.MaxIdleConn
	}
	db.SetMaxIdleConns(MaxIdle)

	if conf.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(conf.ConnMaxIdleTime)
	}

	if conf.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(conf.ConnMaxLifetime)
	}
	return &DbPool{
		db:   db,
		conf: conf,
	}, nil
}
