package xdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForceIndexStrictIdentifierValidation(t *testing.T) {
	validOpts := &Options{}
	ForceIndex("idx_users_status")(validOpts)
	require.NoError(t, (&model{strictIdentifier: true}).validateOptionsIdentifiers(validOpts))

	tests := []struct {
		name  string
		index string
	}{
		{name: "empty", index: ""},
		{name: "qualified", index: "schema.idx_status"},
		{name: "pre-quoted", index: "`idx_status`"},
		{name: "leading whitespace", index: " idx_status"},
		{name: "trailing whitespace", index: "idx_status "},
		{name: "wildcard", index: "*"},
		{name: "SQL-like", index: "idx_status; drop table users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{}
			ForceIndex(tt.index)(opts)

			err := (&model{strictIdentifier: true}).validateOptionsIdentifiers(opts)

			require.ErrorIs(t, err, ErrInvalidIdentifier)
		})
	}
}

func TestForceIndexMySQLSelectPlacesHintBeforeWhere(t *testing.T) {
	sql, args := SelectBuilderWithDialect(
		&MySQLDialect{},
		table("users"),
		ForceIndex("idx_status"),
		WhereEq("status", "active"),
	)

	assert.Equal(t, "select * from `users` force index (`idx_status`) where `status` = ?", sql)
	assert.Equal(t, []any{"active"}, args)
}

func TestForceIndexAccumulatesAcrossCalls(t *testing.T) {
	sql, args := SelectBuilderWithDialect(
		&MySQLDialect{},
		table("users"),
		ForceIndex("idx_status"),
		ForceIndex("idx_created_at", "idx_tenant_id"),
	)

	assert.Equal(t, "select * from `users` force index (`idx_status`, `idx_created_at`, `idx_tenant_id`)", sql)
	assert.Empty(t, args)
}

func TestForceIndexEmptyInputIsNoOp(t *testing.T) {
	sql, args := SelectBuilder(
		table("users"),
		ForceIndex(),
		WhereEq("id", 1),
	)

	assert.Equal(t, "select * from users where id = ?", sql)
	assert.Equal(t, []any{1}, args)
}

func TestForceIndexSelectWithoutOptionIsUnchanged(t *testing.T) {
	sql, args := SelectBuilder(
		table("users"),
		WhereEq("id", 1),
	)

	assert.Equal(t, "select * from users where id = ?", sql)
	assert.Equal(t, []any{1}, args)
}

func TestForceIndexNilDialectUsesQuotedMySQLSyntax(t *testing.T) {
	sql, args := SelectBuilder(
		table("users"),
		ForceIndex("idx`unsafe", "schema.idx_name"),
	)

	assert.Equal(t, "select * from users force index (`idx``unsafe`, `schema.idx_name`)", sql)
	assert.Empty(t, args)
}

func TestForceIndexQuotesPreQuotedNameAtomically(t *testing.T) {
	sql, args := SelectBuilderWithDialect(
		&MySQLDialect{},
		table("users"),
		ForceIndex("`idx`) WHERE 1=1 -- `"),
	)

	assert.Equal(t, "select * from `users` force index (```idx``) WHERE 1=1 -- ```)", sql)
	assert.Empty(t, args)
}

func TestForceIndexUnsupportedDialectsOmitHint(t *testing.T) {
	tests := []struct {
		name    string
		dialect Dialect
		wantSQL string
	}{
		{name: "PostgreSQL", dialect: &PostgreSQLDialect{}, wantSQL: `select * from "users" where "id" = ?`},
		{name: "SQLite", dialect: &SQLiteDialect{}, wantSQL: `select * from "users" where "id" = ?`},
		{name: "legacy", dialect: &legacyDialect{}, wantSQL: "select * from users where id = ?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := SelectBuilderWithDialect(
				tt.dialect,
				table("users"),
				ForceIndex("idx_id"),
				WhereEq("id", 1),
			)

			assert.Equal(t, tt.wantSQL, sql)
			assert.Equal(t, []any{1}, args)
		})
	}
}
