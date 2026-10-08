package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openExprMySQL(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("XDB_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set XDB_MYSQL_DSN to run MySQL expression integration tests")
	}
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestMySQLUpdateExprJSONCAS(t *testing.T) {
	db := openExprMySQL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `CREATE TEMPORARY TABLE xdb_update_expr (
		tenant_id BIGINT NOT NULL,
		conversation_id VARCHAR(64) NOT NULL,
		state JSON,
		status VARCHAR(32) NOT NULL,
		PRIMARY KEY (tenant_id, conversation_id)
	) ENGINE=InnoDB`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `INSERT INTO xdb_update_expr VALUES
		(1, 'null-state', NULL, 'initial'),
		(1, 'existing-state', ?, 'initial'),
		(2, 'null-state', NULL, 'other-tenant')`,
		`{"support":{"revision":2,"pending_operation_id":"old"},"other":{"keep":true}}`)
	require.NoError(t, err)
	m := New("xdb_update_expr", WithDB(db), WithDriver("mysql"), ColumnHook(Json("state"))).Tx(tx).Ctx(ctx)
	update := func(conversation string, revision int, operation string) (bool, error) {
		return m.Update(Record{
			"state": Expr("JSON_SET(COALESCE(state, JSON_OBJECT()), ?, JSON_OBJECT('revision', ?, 'pending_operation_id', ?))",
				"$.support", revision+1, operation),
			"status": operation,
		}, WhereEq("tenant_id", 1), WhereEq("conversation_id", conversation),
			WhereRawArgs("COALESCE(JSON_EXTRACT(state, ?), 0) = ?", "$.support.revision", revision))
	}

	ok, err := update("null-state", 0, "op-init")
	require.NoError(t, err)
	require.True(t, ok)
	var state, status string
	err = tx.QueryRowContext(ctx, "SELECT state, status FROM xdb_update_expr WHERE tenant_id = 1 AND conversation_id = 'null-state'").Scan(&state, &status)
	require.NoError(t, err)
	assert.JSONEq(t, `{"support":{"revision":1,"pending_operation_id":"op-init"}}`, state)
	assert.Equal(t, "op-init", status)

	ok, err = update("existing-state", 2, "op-next")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = update("existing-state", 2, "stale")
	require.NoError(t, err)
	assert.False(t, ok)
	err = tx.QueryRowContext(ctx, "SELECT state, status FROM xdb_update_expr WHERE tenant_id = 1 AND conversation_id = 'existing-state'").Scan(&state, &status)
	require.NoError(t, err)
	assert.JSONEq(t, `{"support":{"revision":3,"pending_operation_id":"op-next"},"other":{"keep":true}}`, state)
	assert.Equal(t, "op-next", status)

	var otherState sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT state, status FROM xdb_update_expr WHERE tenant_id = 2 AND conversation_id = 'null-state'").Scan(&otherState, &status)
	require.NoError(t, err)
	assert.False(t, otherState.Valid)
	assert.Equal(t, "other-tenant", status)
}

func TestMySQLUpdateExprConcurrentCAS(t *testing.T) {
	db := openExprMySQL(t)
	tableName := fmt.Sprintf("xdb_expr_cas_%d", time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, "CREATE TABLE "+tableName+" (id BIGINT PRIMARY KEY, state JSON, winner VARCHAR(32)) ENGINE=InnoDB")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE " + tableName)
		assert.NoError(t, err)
	})
	_, err = db.ExecContext(ctx, "INSERT INTO "+tableName+" (id, state) VALUES (1, NULL)")
	require.NoError(t, err)

	type result struct {
		operation string
		ok        bool
		err       error
	}
	const contenders = 8
	results := make(chan result, contenders)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < contenders; i++ {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			m := New(tableName, WithDB(db), WithDriver("mysql"), ColumnHook(Json("state"))).Ctx(ctx)
			operation := fmt.Sprintf("op-%d", i)
			<-start
			ok, err := m.Update(Record{
				"state": Expr("JSON_SET(COALESCE(state, JSON_OBJECT()), ?, JSON_OBJECT('revision', ?, 'pending_operation_id', ?))",
					"$.support", 1, operation),
				"winner": operation,
			}, WhereEq("id", 1), WhereRawArgs("COALESCE(JSON_EXTRACT(state, ?), 0) = ?", "$.support.revision", 0))
			results <- result{operation: operation, ok: ok, err: err}
		}(i)
	}
	close(start)
	workers.Wait()
	close(results)
	successes, winningOperation := 0, ""
	for result := range results {
		require.NoError(t, result.err)
		if result.ok {
			successes++
			winningOperation = result.operation
		}
	}
	require.Equal(t, 1, successes)
	var revision int
	var operation, winner string
	err = db.QueryRowContext(ctx, "SELECT JSON_EXTRACT(state, '$.support.revision'), JSON_UNQUOTE(JSON_EXTRACT(state, '$.support.pending_operation_id')), winner FROM "+tableName+" WHERE id = 1").Scan(&revision, &operation, &winner)
	require.NoError(t, err)
	assert.Equal(t, 1, revision)
	assert.Equal(t, winningOperation, operation)
	assert.Equal(t, winningOperation, winner)
}
