package xdb

import (
	"fmt"
	"testing"
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
