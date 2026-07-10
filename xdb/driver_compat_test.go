package xdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const compatDriverName = "xdb_compat_driver"

var compatState = struct {
	sync.Mutex
	execDSN   string
	execQuery string
	execArgs  []driver.NamedValue
	queryDSN  string
	querySQL  string
	queryArgs []driver.NamedValue
	pings     int
	txOptions driver.TxOptions
	beginTxs  int
}{}

type compatDriver struct{}

func (compatDriver) Open(dsn string) (driver.Conn, error) {
	return &compatConn{dsn: dsn}, nil
}

type compatConn struct {
	dsn string
}

func (c *compatConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("prepare not used") }
func (c *compatConn) Close() error                        { return nil }
func (c *compatConn) Begin() (driver.Tx, error)           { return compatTx{}, nil }

func (c *compatConn) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.dsn == "ping-fails" {
		return errors.New("ping failed")
	}
	compatState.Lock()
	compatState.pings++
	compatState.Unlock()
	return nil
}

func (c *compatConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	compatState.Lock()
	compatState.txOptions = opts
	compatState.beginTxs++
	compatState.Unlock()
	return compatTx{}, nil
}

func (c *compatConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	compatState.Lock()
	compatState.execDSN = c.dsn
	compatState.execQuery = query
	compatState.execArgs = append([]driver.NamedValue(nil), args...)
	compatState.Unlock()
	return driver.RowsAffected(1), nil
}

func (c *compatConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	compatState.Lock()
	compatState.queryDSN = c.dsn
	compatState.querySQL = query
	compatState.queryArgs = append([]driver.NamedValue(nil), args...)
	compatState.Unlock()
	return &compatRows{values: [][]driver.Value{{int64(1)}}}, nil
}

type compatTx struct{}

func (compatTx) Commit() error   { return nil }
func (compatTx) Rollback() error { return nil }

type compatRows struct {
	values [][]driver.Value
	index  int
}

func (r *compatRows) Columns() []string {
	return []string{"id"}
}

func (r *compatRows) Close() error {
	return nil
}

func (r *compatRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func init() {
	sql.Register(compatDriverName, compatDriver{})
}

func resetCompatState() {
	compatState.Lock()
	defer compatState.Unlock()
	compatState.execDSN = ""
	compatState.execQuery = ""
	compatState.execArgs = nil
	compatState.queryDSN = ""
	compatState.querySQL = ""
	compatState.queryArgs = nil
	compatState.pings = 0
	compatState.txOptions = driver.TxOptions{}
	compatState.beginTxs = 0
}

func TestRawExecUsesContextAndDialect(t *testing.T) {
	resetCompatState()
	db, err := sql.Open(compatDriverName, "write")
	require.NoError(t, err)
	defer db.Close()

	m := New("users", WithDB(db), WithDriver("postgres"))
	_, err = m.Exec("update users set name = ? where id = ?", "alice", 7)
	require.NoError(t, err)

	compatState.Lock()
	assert.Equal(t, "write", compatState.execDSN)
	assert.Equal(t, "update users set name = $1 where id = $2", compatState.execQuery)
	assert.Len(t, compatState.execArgs, 2)
	compatState.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = m.Ctx(ctx).Exec("update users set name = ? where id = ?", "bob", 8)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRawQueryUsesReadDBAndDialect(t *testing.T) {
	resetCompatState()
	writeDB, err := sql.Open(compatDriverName, "write")
	require.NoError(t, err)
	defer writeDB.Close()
	readDB, err := sql.Open(compatDriverName, "read")
	require.NoError(t, err)
	defer readDB.Close()

	m := New("users", WithDB(writeDB), WithReadDB(readDB), WithDriver("postgres"))
	rows, err := m.Query("select * from users where id = ?", 9)
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	compatState.Lock()
	assert.Equal(t, "read", compatState.queryDSN)
	assert.Equal(t, "select * from users where id = $1", compatState.querySQL)
	assert.Len(t, compatState.queryArgs, 1)
	compatState.Unlock()
}

func TestRawQueryRowUsesReadDBAndDialect(t *testing.T) {
	resetCompatState()
	writeDB, err := sql.Open(compatDriverName, "write")
	require.NoError(t, err)
	defer writeDB.Close()
	readDB, err := sql.Open(compatDriverName, "read")
	require.NoError(t, err)
	defer readDB.Close()

	m := New("users", WithDB(writeDB), WithReadDB(readDB), WithDriver("postgres"))
	var id int64
	err = m.QueryRow("select id from users where id = ?", 9).Scan(&id)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	compatState.Lock()
	assert.Equal(t, "read", compatState.queryDSN)
	assert.Equal(t, "select id from users where id = $1", compatState.querySQL)
	assert.Len(t, compatState.queryArgs, 1)
	compatState.Unlock()
}

func TestTransactionWithOptions(t *testing.T) {
	resetCompatState()
	db, err := sql.Open(compatDriverName, "write")
	require.NoError(t, err)
	defer db.Close()

	m := New("users", WithDB(db), WithDriver("postgres"))
	err = m.TransactionWithOptions(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  true,
	}, func(tx *sql.Tx, txModel Model) error {
		_, err := txModel.Exec("update users set name = ? where id = ?", "alice", 7)
		return err
	})
	require.NoError(t, err)

	compatState.Lock()
	assert.Equal(t, 1, compatState.beginTxs)
	assert.True(t, compatState.txOptions.ReadOnly)
	assert.Equal(t, driver.IsolationLevel(sql.LevelSerializable), compatState.txOptions.Isolation)
	assert.Equal(t, "update users set name = $1 where id = $2", compatState.execQuery)
	compatState.Unlock()
}

func TestInitPingsAndCloseClosesPoolEntries(t *testing.T) {
	resetCompatState()
	err := Init(map[string]*Config{
		"compat": {Driver: compatDriverName, DSN: "write"},
	})
	require.NoError(t, err)

	compatState.Lock()
	assert.Equal(t, 1, compatState.pings)
	compatState.Unlock()

	assert.NotPanics(t, Close)
	_, err = DB("compat")
	assert.Error(t, err)
}

func TestInitReadDSNInheritsPingOptions(t *testing.T) {
	resetCompatState()
	err := Init(map[string]*Config{
		"compat": {
			Driver:      compatDriverName,
			DSN:         "write",
			ReadDsn:     "ping-fails",
			DisablePing: true,
		},
	})
	require.NoError(t, err)
	defer Close()

	compatState.Lock()
	assert.Equal(t, 0, compatState.pings)
	compatState.Unlock()
}

func TestModelWritesQuoteIdentifiersWithDialect(t *testing.T) {
	resetCompatState()
	db, err := sql.Open(compatDriverName, "write")
	require.NoError(t, err)
	defer db.Close()

	m := New("event_log", WithDB(db), WithDriver("postgres"), WithPrimaryKey("event_id"))
	ok, err := m.Update(Record{"event_id": 1, "event_type": "login"})
	require.NoError(t, err)
	assert.True(t, ok)

	compatState.Lock()
	assert.Equal(t, `update "event_log" set "event_type" = $1 where "event_id" = $2`, compatState.execQuery)
	compatState.Unlock()

	affected, err := m.InsertIgnore(Record{"event_id": 1, "event_type": "login"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	compatState.Lock()
	assert.Equal(t, `INSERT INTO "event_log" ("event_id", "event_type") VALUES ($1, $2) ON CONFLICT DO NOTHING`, compatState.execQuery)
	compatState.Unlock()

	affected, err = m.InsertOrUpdate(Record{"event_id": 1, "event_type": "logout"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	compatState.Lock()
	assert.Equal(t, `INSERT INTO "event_log" ("event_id", "event_type") VALUES ($1, $2) ON CONFLICT ("event_id") DO UPDATE SET "event_type" = EXCLUDED."event_type"`, compatState.execQuery)
	compatState.Unlock()
}

func TestStrictIdentifierRejectsStructuredIdentifiersAndAllowsRaw(t *testing.T) {
	resetCompatState()
	db, err := sql.Open(compatDriverName, "write")
	require.NoError(t, err)
	defer db.Close()

	m := New("users", WithDB(db), WithDriver("postgres"), WithStrictIdentifier())
	_, err = m.Update(Record{"bad field": "alice"}, WhereEq("id", 1))
	assert.ErrorIs(t, err, ErrInvalidIdentifier)

	ok, err := m.Delete(WhereRawArgs("data->>'kind' = ?", "login"))
	require.NoError(t, err)
	assert.True(t, ok)

	compatState.Lock()
	assert.Equal(t, `delete from "users" where data->>'kind' = $1`, compatState.execQuery)
	compatState.Unlock()
}
