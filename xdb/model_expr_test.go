package xdb

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newExprTestModel(t *testing.T, driver string, opts ...With) Model {
	t.Helper()
	resetCompatState()
	db, err := sql.Open(compatDriverName, "write")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return New("users", append([]With{WithDB(db), WithDriver(driver)}, opts...)...)
}

func TestModelUpdateExprHooksAndDialect(t *testing.T) {
	for _, tc := range []struct {
		driver  string
		wantSQL string
	}{
		{"mysql", "update `users` set `metadata` = ?,`profile` = COALESCE(profile, ?) where `id` = ?"},
		{"postgres", `update "users" set "metadata" = $1,"profile" = COALESCE(profile, $2) where "id" = $3`},
	} {
		t.Run(tc.driver, func(t *testing.T) {
			validated := false
			m := newExprTestModel(t, tc.driver,
				WithStrictIdentifier(),
				ColumnHook(Json("profile"), Json("metadata")),
				ColumnValidator(Validate("profile", func(v *ValidInfo) error {
					validated = true
					assert.IsType(t, Expression{}, v.Row["profile"])
					assert.IsType(t, "", v.Row["metadata"])
					return nil
				})),
			)
			ok, err := m.Update(Record{
				"profile":  Expr("COALESCE(profile, ?)", `{"revision":1}`),
				"metadata": map[string]any{"source": "sync"},
			}, WhereEq("id", 7))
			require.NoError(t, err)
			assert.True(t, ok)
			assert.True(t, validated)

			compatState.Lock()
			defer compatState.Unlock()
			assert.Equal(t, tc.wantSQL, compatState.execQuery)
			require.Len(t, compatState.execArgs, 3)
			assert.JSONEq(t, `{"source":"sync"}`, compatState.execArgs[0].Value.(string))
			assert.Equal(t, `{"revision":1}`, compatState.execArgs[1].Value)
			assert.Equal(t, int64(7), compatState.execArgs[2].Value)
		})
	}
}

func TestModelUpdateExprValidatorError(t *testing.T) {
	wantErr := errors.New("update rejected")
	m := newExprTestModel(t, "mysql", ColumnValidator(Validate("profile", func(v *ValidInfo) error {
		return wantErr
	})))
	ok, err := m.Update(Record{"profile": Expr("NULL")}, WhereEq("id", 7))
	assert.ErrorIs(t, err, wantErr)
	assert.False(t, ok)
	compatState.Lock()
	defer compatState.Unlock()
	assert.Empty(t, compatState.execQuery)
}

func TestModelUpdateExprOrdinaryHookError(t *testing.T) {
	m := newExprTestModel(t, "mysql", ColumnHook(Json("metadata")))
	ok, err := m.Update(Record{"profile": Expr("NULL"), "metadata": make(chan int)}, WhereEq("id", 7))
	require.Error(t, err)
	assert.False(t, ok)
	compatState.Lock()
	defer compatState.Unlock()
	assert.Empty(t, compatState.execQuery)
}

func TestModelUpdateExprStrictIdentifier(t *testing.T) {
	m := newExprTestModel(t, "mysql", WithStrictIdentifier())
	ok, err := m.Update(Record{"profile; DROP TABLE users": Expr("NULL")}, WhereEq("id", 7))
	assert.ErrorIs(t, err, ErrInvalidIdentifier)
	assert.False(t, ok)
	compatState.Lock()
	defer compatState.Unlock()
	assert.Empty(t, compatState.execQuery)
}

func TestModelUpdateExprTransaction(t *testing.T) {
	m := newExprTestModel(t, "postgres")
	err := m.Transaction(func(_ *sql.Tx, txModel Model) error {
		ok, err := txModel.Update(Record{"score": Expr("COALESCE(score, ?) + ?", 0, 5)}, WhereRawArgs("id = ?", 7))
		assert.True(t, ok)
		return err
	})
	require.NoError(t, err)
	compatState.Lock()
	defer compatState.Unlock()
	assert.Equal(t, `update "users" set "score" = COALESCE(score, $1) + $2 where id = $3`, compatState.execQuery)
	require.Len(t, compatState.execArgs, 3)
	assert.Equal(t, int64(0), compatState.execArgs[0].Value)
	assert.Equal(t, int64(5), compatState.execArgs[1].Value)
	assert.Equal(t, int64(7), compatState.execArgs[2].Value)
}

func TestModelUpdateExprOtherWriteHooksUnchanged(t *testing.T) {
	for _, method := range []string{"insert", "ignore", "upsert"} {
		t.Run(method, func(t *testing.T) {
			m := newExprTestModel(t, "postgres", ColumnHook(Json("profile")))
			record := Record{"profile": map[string]any{"revision": 1}}
			var err error
			switch method {
			case "insert":
				_, err = m.Insert(record)
			case "ignore":
				_, err = m.InsertIgnore(record)
			case "upsert":
				_, err = m.InsertOrUpdate(record)
			}
			require.NoError(t, err)
			compatState.Lock()
			defer compatState.Unlock()
			args := compatState.execArgs
			if method == "insert" {
				args = compatState.queryArgs
			}
			require.Len(t, args, 1)
			assert.JSONEq(t, `{"revision":1}`, args[0].Value.(string))
		})
	}
}
