package xdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateExprMixedAssignments(t *testing.T) {
	query, args := UpdateBuilder(
		table("conversations"),
		WhereEq("tenant_id", 42),
		WhereRawArgs("COALESCE(JSON_EXTRACT(state, ?), 0) = ?", "$.support.revision", 3),
		Field("name", "state", "attempts", "remaining", "updated_at", "deleted_at", "score"),
		Value(
			"CURRENT_TIMESTAMP",
			Expr("JSON_SET(COALESCE(state, JSON_OBJECT()), ?, JSON_OBJECT('revision', ?, 'operation', ?))", "$.support", 4, "op-1"),
			SelfAdd(1), SelfSub(2), Expr("CURRENT_TIMESTAMP"), nil,
			Expr("COALESCE(score, ?) + ?", 0, 5),
		),
	)

	assert.Equal(t, "update conversations set name = ?,state = JSON_SET(COALESCE(state, JSON_OBJECT()), ?, JSON_OBJECT('revision', ?, 'operation', ?)),attempts = attempts + ?,remaining = remaining - ?,updated_at = CURRENT_TIMESTAMP,deleted_at = ?,score = COALESCE(score, ?) + ? where tenant_id = ? and COALESCE(JSON_EXTRACT(state, ?), 0) = ?", query)
	assert.Equal(t, []any{"CURRENT_TIMESTAMP", "$.support", 4, "op-1", 1, 2, nil, 0, 5, 42, "$.support.revision", 3}, args)
}

func TestUpdateExprDialects(t *testing.T) {
	for _, tc := range []struct {
		name    string
		dialect Dialect
		wantSQL string
	}{
		{"default", nil, "update users set score = COALESCE(score, ?) + ?,label = COALESCE(label, '?') where id = ?"},
		{"mysql", &MySQLDialect{}, "update `users` set `score` = COALESCE(score, ?) + ?,`label` = COALESCE(label, '?') where `id` = ?"},
		{"postgres", &PostgreSQLDialect{}, `update "users" set "score" = COALESCE(score, $1) + $2,"label" = COALESCE(label, '?') where "id" = $3`},
		{"sqlite", &SQLiteDialect{}, `update "users" set "score" = COALESCE(score, ?) + ?,"label" = COALESCE(label, '?') where "id" = ?`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, args := UpdateBuilderWithDialect(tc.dialect,
				table("users"), Field("score", "label"),
				Value(Expr("COALESCE(score, ?) + ?", 0, 5), Expr("COALESCE(label, '?')")),
				WhereEq("id", 7),
			)
			if tc.dialect != nil {
				query = tc.dialect.ConvertPlaceholders(query)
			}
			assert.Equal(t, tc.wantSQL, query)
			assert.Equal(t, []any{0, 5, 7}, args)
		})
	}
}

func TestUpdateExprReusableValues(t *testing.T) {
	expressionArgs := []any{0, 5}
	values := []any{SelfAdd(1), Expr("COALESCE(score, ?) + ?", expressionArgs...), SelfSub(2)}
	opts := []Option{table("users"), Field("attempts", "score", "remaining"), Value(values...), WhereEq("id", 7)}

	for i := 0; i < 2; i++ {
		query, args := UpdateBuilder(opts...)
		assert.Equal(t, "update users set attempts = attempts + ?,score = COALESCE(score, ?) + ?,remaining = remaining - ? where id = ?", query)
		assert.Equal(t, []any{1, 0, 5, 2, 7}, args)
	}
	assert.Equal(t, SelfAdd(1), values[0])
	assert.Equal(t, SelfSub(2), values[2])
	assert.Equal(t, []any{0, 5}, expressionArgs)
}

func TestUpdateExprWithoutArguments(t *testing.T) {
	query, args := UpdateBuilder(table("users"), Field("updated_at"), Value(Expr("CURRENT_TIMESTAMP")))
	assert.Equal(t, "update users set updated_at = CURRENT_TIMESTAMP", query)
	assert.Empty(t, args)
}
