package xdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectBuilder(t *testing.T) {
	sql, args := SelectBuilder(
		table("user"),
		Field("id", "name", "avatar"),
		WhereEq("id", 1),
		WhereGt("age", 20),
		WhereLike("name", "dd"),
		WhereGroup(
			WhereEq("sex", 1),
			WhereOrEq("class", 2),
			WhereOrGroup(
				WhereEq("sex1", 1),
				WhereEq("class2", 2),
			),
		),
		WhereFindInSet("role", 3),
		OrderByDesc("id"),
	)

	fmt.Println(sql, args)
}

func TestInsertBuilder(t *testing.T) {
	sql, args := InsertBuilder(
		table("user"),
		Field("id", "name"),
		Value("1", "daodao"),
	)
	fmt.Println(sql, args)
}

func TestUpdateBuilder(t *testing.T) {
	sql, args := UpdateBuilder(
		table("user"),
		Field("id", "name", "age"),
		Value("1", "daodao", UpdateValue{Value: 18, Op: OpAdd}),
		WhereEq("id", 1),
	)
	fmt.Println(sql, args)
}

func TestDeleteBuilder(t *testing.T) {
	sql, args := DeleteBuilder(
		table("user"),
		WhereEq("id", 1),
	)
	fmt.Println(sql, args)
}

func TestSelectBuilderWithDialect(t *testing.T) {
	sql, args := SelectBuilderWithDialect(
		&PostgreSQLDialect{},
		table("app_user"),
		WhereEq("id", 1),
		Limit(10),
		Offset(20),
		ForUpdate(),
	)

	assert.Equal(t, `select * from "app_user" where "id" = ? limit ? offset ? for update`, sql)
	assert.Equal(t, []any{1, 10, 20}, args)
	assert.Equal(t, `select * from "app_user" where "id" = $1 limit $2 offset $3 for update`, (&PostgreSQLDialect{}).ConvertPlaceholders(sql))
}

func TestSelectBuilderWithSQLiteSkipsForUpdate(t *testing.T) {
	sql, args := SelectBuilderWithDialect(
		&SQLiteDialect{},
		table("user"),
		WhereEq("id", 1),
		ForUpdate(),
	)

	assert.Equal(t, `select * from "user" where "id" = ?`, sql)
	assert.Equal(t, []any{1}, args)
}

func TestSelectBuilderWithDialectDefaultsForUpdateForLegacyDialect(t *testing.T) {
	sql, args := SelectBuilderWithDialect(
		&legacyDialect{},
		table("user"),
		WhereEq("id", 1),
		ForUpdate(),
	)

	assert.Equal(t, `select * from user where id = ? for update`, sql)
	assert.Equal(t, []any{1}, args)
}

func TestBuilderWithDialectQuotesIdentifiers(t *testing.T) {
	dialect := &PostgreSQLDialect{}

	sql, args := SelectBuilderWithDialect(
		dialect,
		table("public.users"),
		Field("id", "profile_id as profileId", "count(*) as total"),
		WhereEq("profile.id", 1),
		OrderByDesc("created_at"),
		GroupBy("profile_id"),
	)
	assert.Equal(t, `select "id", "profile_id" as "profileId", count(*) as "total" from "public"."users" where "profile"."id" = ? group by "profile_id" order by "created_at" desc`, sql)
	assert.Equal(t, []any{1}, args)

	sql, args = InsertBuilderWithDialect(
		dialect,
		table("public.users"),
		Field("id", "name"),
		Value(1, "alice"),
	)
	assert.Equal(t, `insert into "public"."users" ("id", "name") values (?,?)`, sql)
	assert.Equal(t, []any{1, "alice"}, args)

	sql, args = UpdateBuilderWithDialect(
		dialect,
		table("public.users"),
		Field("login_count"),
		Value(SelfAdd(1)),
		WhereEq("id", 1),
	)
	assert.Equal(t, `update "public"."users" set "login_count" = "login_count" + ? where "id" = ?`, sql)
	assert.Equal(t, []any{1, 1}, args)

	sql, args = DeleteBuilderWithDialect(
		dialect,
		table("public.users"),
		WhereEq("id", 1),
	)
	assert.Equal(t, `delete from "public"."users" where "id" = ?`, sql)
	assert.Equal(t, []any{1}, args)
}

func TestWhereInEmptySlice(t *testing.T) {
	sql, args := SelectBuilder(table("user"), WhereIn("id", []any{}))
	assert.Equal(t, "select * from user where 1 = 0", sql)
	assert.Empty(t, args)

	sql, args = SelectBuilder(table("user"), WhereNotIn("id", []any{}))
	assert.Equal(t, "select * from user where 1 = 1", sql)
	assert.Empty(t, args)
}
