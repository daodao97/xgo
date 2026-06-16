package xdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereRawArgs_MixedWithWhereEq(t *testing.T) {
	sql, args := SelectBuilder(
		table("todos"),
		WhereEq("type", "todo"),
		WhereRawArgs("attributes->>'$.channel_subject_type' = ?", "user"),
		WhereRawArgs("attributes->>'$.channel_subject_id' = ?", "42"),
	)
	assert.Equal(t,
		"select * from todos where type = ? and attributes->>'$.channel_subject_type' = ? and attributes->>'$.channel_subject_id' = ?",
		sql)
	assert.Equal(t, []any{"todo", "user", "42"}, args)
}

func TestWhereRawArgs_MultiPlaceholder(t *testing.T) {
	sql, args := SelectBuilder(
		table("todos"),
		WhereRawArgs("coalesce(retry_at, ?) <= ?", "fallback", "deadline"),
	)
	assert.Equal(t, "select * from todos where coalesce(retry_at, ?) <= ?", sql)
	assert.Equal(t, []any{"fallback", "deadline"}, args)
}

func TestWhereOrRawArgs_UsesOrLogic(t *testing.T) {
	sql, args := SelectBuilder(
		table("t"),
		WhereEq("a", 1),
		WhereOrRawArgs("b = ?", 2),
	)
	assert.Equal(t, "select * from t where a = ? or b = ?", sql)
	assert.Equal(t, []any{1, 2}, args)
}

func TestWhereRawArgs_InsideWhereGroup(t *testing.T) {
	sql, args := SelectBuilder(
		table("t"),
		WhereEq("a", 1),
		WhereGroup(
			WhereRawArgs("json_contains(tags, ?)", "x"),
			WhereOrEq("b", 2),
		),
	)
	assert.Equal(t, "select * from t where a = ? and (json_contains(tags, ?) or b = ?)", sql)
	assert.Equal(t, []any{1, "x", 2}, args)
}

func TestWhereRawArgs_InUpdateAndDelete(t *testing.T) {
	usql, uargs := UpdateBuilder(
		table("t"),
		Field("status"),
		Value("done"),
		WhereRawArgs("attributes->>'$.k' = ?", "v"),
	)
	assert.Equal(t, "update t set status = ? where attributes->>'$.k' = ?", usql)
	assert.Equal(t, []any{"done", "v"}, uargs)

	dsql, dargs := DeleteBuilder(
		table("t"),
		WhereRawArgs("attributes->>'$.k' = ?", "v"),
	)
	assert.Equal(t, "delete from t where attributes->>'$.k' = ?", dsql)
	assert.Equal(t, []any{"v"}, dargs)
}

// WhereRaw 旧行为不变：raw 片段原样进入，不绑定任何参数。
func TestWhereRaw_BackwardCompat(t *testing.T) {
	sql, args := SelectBuilder(
		table("t"),
		WhereEq("a", 1),
		WhereRaw("b is null"),
	)
	assert.Equal(t, "select * from t where a = ? and b is null", sql)
	assert.Equal(t, []any{1}, args)
}
