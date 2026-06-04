package xdb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakeInsertIgnore* below implement a minimal database/sql driver so that the
// return value of InsertIgnore can be exercised deterministically, without a
// live database. It simulates a table whose primary key is application
// generated (e.g. an xid string) rather than auto-increment: LastInsertId is
// always 0, while RowsAffected reflects whether a row was actually inserted.
// SQLite cannot reproduce this because every row gets an implicit rowid, so a
// purpose-built fake is the only server-free way to pin down the behaviour.

const fakeInsertIgnoreDriverName = "fake_insert_ignore_driver"

// fakeNextLastInsertID / fakeNextRowsAffected control what the next Exec
// reports. Tests set them before each InsertIgnore call.
var (
	fakeNextLastInsertID int64
	fakeNextRowsAffected int64
)

type fakeInsertIgnoreResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r fakeInsertIgnoreResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r fakeInsertIgnoreResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type fakeInsertIgnoreDriver struct{}

func (fakeInsertIgnoreDriver) Open(string) (driver.Conn, error) { return fakeInsertIgnoreConn{}, nil }

type fakeInsertIgnoreConn struct{}

func (fakeInsertIgnoreConn) Prepare(string) (driver.Stmt, error) { return fakeInsertIgnoreStmt{}, nil }
func (fakeInsertIgnoreConn) Close() error                        { return nil }
func (fakeInsertIgnoreConn) Begin() (driver.Tx, error)           { return fakeInsertIgnoreTx{}, nil }

type fakeInsertIgnoreTx struct{}

func (fakeInsertIgnoreTx) Commit() error   { return nil }
func (fakeInsertIgnoreTx) Rollback() error { return nil }

type fakeInsertIgnoreStmt struct{}

func (fakeInsertIgnoreStmt) Close() error  { return nil }
func (fakeInsertIgnoreStmt) NumInput() int { return -1 }
func (fakeInsertIgnoreStmt) Exec([]driver.Value) (driver.Result, error) {
	return fakeInsertIgnoreResult{
		lastInsertID: fakeNextLastInsertID,
		rowsAffected: fakeNextRowsAffected,
	}, nil
}
func (fakeInsertIgnoreStmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, errors.New("query not supported by fake driver")
}

func init() {
	sql.Register(fakeInsertIgnoreDriverName, fakeInsertIgnoreDriver{})
}

func Test_InsertIgnore_ReturnsRowsAffectedForNonAutoIncrementPK(t *testing.T) {
	err := Init(map[string]*Config{
		"fake_insert_ignore": {Driver: fakeInsertIgnoreDriverName, DSN: "fake"},
	})
	assert.NoError(t, err)

	m := New("event_log",
		WithConn("fake_insert_ignore"),
		WithPrimaryKey("id"),
	)

	// Case 1: a brand new row is inserted. On a non-auto-increment PK table
	// LastInsertId is 0 even though the row was inserted, so the dedup gate
	// (affected > 0) can only work if InsertIgnore reports RowsAffected.
	fakeNextLastInsertID = 0
	fakeNextRowsAffected = 1
	affected, err := m.InsertIgnore(Record{"id": "abc", "name": "Alice"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), affected, "fresh insert on a non-auto-increment PK must report affected=1")

	// Case 2: the row collides with the unique key and is silently ignored.
	fakeNextLastInsertID = 0
	fakeNextRowsAffected = 0
	affected, err = m.InsertIgnore(Record{"id": "abc", "name": "Alice"})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), affected, "ignored insert must report affected=0")
}
