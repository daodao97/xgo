package xdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderByRawArgsFollowSQLClauseOrder(t *testing.T) {
	sql, args := SelectBuilder(
		table("dough"),
		OrderByRaw("MATCH(title,description,summary) AGAINST(? IN NATURAL LANGUAGE MODE) DESC", "keyword"),
		WhereEq("type", "bread"),
		WhereEq("tenant_id", int64(42)),
		WhereIn("status", []any{"ready", "draft"}),
		WhereEq("is_deleted", 0),
		Limit(20),
	)

	assert.Equal(t,
		"select * from dough where type = ? and tenant_id = ? and status in (?,?) and is_deleted = ? order by MATCH(title,description,summary) AGAINST(? IN NATURAL LANGUAGE MODE) DESC limit ? offset ?",
		sql,
	)
	assert.Equal(t, []any{"bread", int64(42), "ready", "draft", 0, "keyword", 20, 0}, args)
}

func TestOrderByRawPreservesExpressionAndArgumentOrder(t *testing.T) {
	sql, args := SelectBuilder(
		table("dough"),
		WhereRawArgs("MATCH(title) AGAINST(?)", "where-keyword"),
		OrderByAsc("priority"),
		OrderByRaw("FIELD(status, ?, ?)", "ready", "draft"),
		OrderByRaw("CASE WHEN tenant_id = ? THEN 0 ELSE 1 END", int64(42)),
		OrderByDesc("id"),
	)

	assert.Equal(t,
		"select * from dough where MATCH(title) AGAINST(?) order by priority asc, FIELD(status, ?, ?), CASE WHEN tenant_id = ? THEN 0 ELSE 1 END, id desc",
		sql,
	)
	assert.Equal(t, []any{"where-keyword", "ready", "draft", int64(42)}, args)
}

func TestOrderByRawSupportsExpressionWithoutArguments(t *testing.T) {
	sql, args := SelectBuilder(
		table("dough"),
		OrderByRaw("RAND()"),
	)

	assert.Equal(t, "select * from dough order by RAND()", sql)
	assert.Empty(t, args)
}

func TestOrderByExistingOptionsRemainCompatible(t *testing.T) {
	sql, args := SelectBuilder(
		table("dough"),
		OrderByAsc("priority"),
		OrderByDesc("id"),
	)

	assert.Equal(t, "select * from dough order by priority asc, id desc", sql)
	assert.Empty(t, args)
}

func TestOrderByRawPostgreSQLPlaceholderOrder(t *testing.T) {
	sql, args := SelectBuilder(
		table("dough"),
		OrderByRaw("CASE WHEN tenant_id = ? THEN ? ELSE ? END DESC", int64(42), 1, 0),
		WhereEq("type", "bread"),
		Limit(10),
	)

	sql = GetDialect("postgres").ConvertPlaceholders(sql)

	assert.Equal(t,
		"select * from dough where type = $1 order by CASE WHEN tenant_id = $2 THEN $3 ELSE $4 END DESC limit $5 offset $6",
		sql,
	)
	assert.Equal(t, []any{"bread", int64(42), 1, 0, 10, 0}, args)
}

func TestOrderByRawWithDialectQuotesOnlyStructuredItems(t *testing.T) {
	sql, args := SelectBuilderWithDialect(
		&MySQLDialect{},
		table("dough"),
		WhereEq("type", "bread"),
		OrderByDesc("created_at"),
		OrderByRaw("MATCH(title,description,summary) AGAINST(? IN NATURAL LANGUAGE MODE) DESC", "keyword"),
		Limit(20),
	)

	assert.Equal(t,
		"select * from `dough` where `type` = ? order by `created_at` desc, MATCH(title,description,summary) AGAINST(? IN NATURAL LANGUAGE MODE) DESC limit ? offset ?",
		sql,
	)
	assert.Equal(t, []any{"bread", "keyword", 20, 0}, args)
}

func TestOrderByRawIsAllowedByStrictIdentifierValidation(t *testing.T) {
	opts := &Options{}
	for _, opt := range []Option{
		table("dough"),
		OrderByDesc("created_at"),
		OrderByRaw("MATCH(title) AGAINST(?) DESC", "keyword"),
	} {
		opt(opts)
	}

	err := (&model{strictIdentifier: true}).validateOptionsIdentifiers(opts)

	assert.NoError(t, err)
}

func TestStructuredOrderByStillUsesStrictIdentifierValidation(t *testing.T) {
	opts := &Options{}
	for _, opt := range []Option{
		table("dough"),
		OrderByDesc("created_at; drop table dough"),
	} {
		opt(opts)
	}

	err := (&model{strictIdentifier: true}).validateOptionsIdentifiers(opts)

	assert.ErrorIs(t, err, ErrInvalidIdentifier)
}

func TestOrderByRawArgumentsDoNotLeakIntoMutationBuilders(t *testing.T) {
	insertSQL, insertArgs := InsertBuilder(
		table("dough"),
		Field("name"),
		Value("baguette"),
		OrderByRaw("FIELD(status, ?)", "ready"),
	)
	assert.Equal(t, "insert into dough (name) values (?)", insertSQL)
	assert.Equal(t, []any{"baguette"}, insertArgs)

	updateSQL, updateArgs := UpdateBuilder(
		table("dough"),
		Field("status"),
		Value("done"),
		WhereEq("id", 1),
		OrderByRaw("FIELD(status, ?)", "ready"),
	)
	assert.Equal(t, "update dough set status = ? where id = ?", updateSQL)
	assert.Equal(t, []any{"done", 1}, updateArgs)

	deleteSQL, deleteArgs := DeleteBuilder(
		table("dough"),
		WhereEq("id", 1),
		OrderByRaw("FIELD(status, ?)", "ready"),
	)
	assert.Equal(t, "delete from dough where id = ?", deleteSQL)
	assert.Equal(t, []any{1}, deleteArgs)
}
