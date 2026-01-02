# xdb

xdb 是一个轻量级的 Go 数据库 ORM，支持 MySQL、PostgreSQL 和 SQLite 多种数据库。

## 特性

- 支持多种数据库：MySQL、PostgreSQL、SQLite
- 链式查询构建器
- 自动占位符转换
- 内置缓存支持
- 钩子系统（数据转换）
- 验证器
- 关联关系（hasOne, hasMany）
- 事务支持
- 软删除

## 安装

```bash
go get github.com/daodao97/xgo/xdb
```

根据使用的数据库，还需要安装对应的驱动：

```bash
# MySQL
go get github.com/go-sql-driver/mysql

# PostgreSQL
go get github.com/lib/pq

# SQLite
go get github.com/mattn/go-sqlite3
```

## 快速开始

### 初始化连接

```go
import (
    "github.com/daodao97/xgo/xdb"
    _ "github.com/go-sql-driver/mysql"    // MySQL
    // _ "github.com/lib/pq"              // PostgreSQL
    // _ "github.com/mattn/go-sqlite3"    // SQLite
)

// MySQL
err := xdb.Init(map[string]*xdb.Config{
    "default": {
        Driver: "mysql",
        DSN:    "user:password@tcp(localhost:3306)/dbname?parseTime=true",
    },
})

// PostgreSQL
err := xdb.Init(map[string]*xdb.Config{
    "default": {
        Driver: "postgres",
        DSN:    "host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable",
    },
})

// SQLite
err := xdb.Init(map[string]*xdb.Config{
    "default": {
        Driver: "sqlite3",
        DSN:    "./database.db",
    },
})
```

### 基本 CRUD 操作

```go
// 创建 Model
m := xdb.New("users")

// 插入
lastId, err := m.Insert(xdb.Record{
    "name":  "Alice",
    "email": "alice@example.com",
})

// 查询单条
record, err := m.First(xdb.WhereEq("id", 1))

// 查询多条
records, err := m.Selects(
    xdb.WhereGt("age", 18),
    xdb.OrderByDesc("created_at"),
    xdb.Limit(10),
)

// 分页查询
total, records, err := m.Page(1, 20, xdb.WhereEq("status", 1))

// 更新
ok, err := m.Update(xdb.Record{
    "id":   1,
    "name": "Bob",
})

// 删除
ok, err := m.Delete(xdb.WhereEq("id", 1))
```

## 多数据库支持

xdb 通过 `Dialect` 接口实现多数据库兼容，自动处理不同数据库的语法差异。

### 支持的数据库

| 数据库 | Driver 值 | 占位符 | 特性 |
|--------|-----------|--------|------|
| MySQL | `mysql` | `?` | ON DUPLICATE KEY UPDATE |
| PostgreSQL | `postgres`, `postgresql`, `pgx` | `$1, $2, ...` | ON CONFLICT + RETURNING |
| SQLite | `sqlite`, `sqlite3` | `?` | ON CONFLICT |

### UPSERT (插入或更新)

```go
// 所有数据库统一 API
affected, err := m.InsertOrUpdate(xdb.Record{
    "id":    1,
    "name":  "Alice",
    "email": "alice@example.com",
})

// MySQL 生成:
// INSERT INTO users (id, name, email) VALUES (?, ?, ?)
// ON DUPLICATE KEY UPDATE name = ?, email = ?

// PostgreSQL 生成:
// INSERT INTO users (id, name, email) VALUES ($1, $2, $3)
// ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email

// SQLite 生成:
// INSERT INTO users (id, name, email) VALUES (?, ?, ?)
// ON CONFLICT (id) DO UPDATE SET name = excluded.name, email = excluded.email
```

### INSERT IGNORE (忽略冲突)

```go
// 所有数据库统一 API
lastId, err := m.InsertIgnore(xdb.Record{
    "id":   1,
    "name": "Alice",
})

// MySQL 生成:
// INSERT IGNORE INTO users (id, name) VALUES (?, ?)

// PostgreSQL 生成:
// INSERT INTO users (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING

// SQLite 生成:
// INSERT OR IGNORE INTO users (id, name) VALUES (?, ?)
```

### 批量插入

```go
lastId, err := m.InsertBatch([]xdb.Record{
    {"name": "Alice", "email": "alice@example.com"},
    {"name": "Bob", "email": "bob@example.com"},
})
// 自动处理占位符转换和 RETURNING（PostgreSQL）
```

### 自动占位符转换

xdb 内部使用 `?` 作为占位符，在执行时自动转换：

```go
// 代码中使用统一的查询方式
records, err := m.Selects(
    xdb.WhereEq("status", 1),
    xdb.WhereGt("age", 18),
)

// MySQL/SQLite 执行:
// SELECT * FROM users WHERE status = ? AND age > ?

// PostgreSQL 执行:
// SELECT * FROM users WHERE status = $1 AND age > $2
```

## 查询条件

```go
// 等于
xdb.WhereEq("field", value)

// 不等于
xdb.WhereNotEq("field", value)

// 大于/大于等于
xdb.WhereGt("field", value)
xdb.WhereGte("field", value)

// 小于/小于等于
xdb.WhereLt("field", value)
xdb.WhereLte("field", value)

// IN / NOT IN
xdb.WhereIn("field", []any{1, 2, 3})
xdb.WhereNotIn("field", []any{1, 2, 3})

// LIKE
xdb.WhereLike("name", "%alice%")

// BETWEEN
xdb.WhereBetween("age", 18, 30)

// IS NULL / IS NOT NULL
xdb.WhereIsNil("deleted_at")
xdb.WhereNotNil("deleted_at")

// OR 条件
xdb.WhereOrEq("status", 1)
xdb.WhereOrLike("name", "%test%")

// 条件组合
xdb.WhereGroup(
    xdb.WhereEq("a", 1),
    xdb.WhereOrEq("b", 2),
)
// 生成: (a = ? OR b = ?)

// 原始条件
xdb.WhereRaw("age > 18 AND status = 1")

// FIND_IN_SET (仅 MySQL)
xdb.WhereFindInSet("tags", "golang")
```

## 排序和分页

```go
// 排序
xdb.OrderByDesc("created_at")
xdb.OrderByAsc("id")

// 分页
xdb.Limit(10)
xdb.Offset(20)

// 或使用 Pagination 快捷方式
xdb.Pagination(page, pageSize)

// 分组
xdb.GroupBy("category")
```

## 事务

```go
err := m.Transaction(func(tx *sql.Tx, m xdb.Model) error {
    _, err := m.Insert(xdb.Record{"name": "Alice"})
    if err != nil {
        return err // 自动回滚
    }

    _, err = m.Update(xdb.Record{"id": 1, "status": 1})
    if err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})
```

## 钩子系统

```go
m := xdb.New("users",
    xdb.ColumnHook(
        // JSON 字段自动序列化/反序列化
        xdb.Json("profile"),
        // 逗号分隔的整数
        xdb.CommaInt("role_ids"),
    ),
)

// 插入时 profile 会自动 JSON 序列化
m.Insert(xdb.Record{
    "name": "Alice",
    "profile": map[string]any{"hobby": "coding"},
    "role_ids": []int{1, 2, 3},
})

// 查询时自动反序列化
record, _ := m.First(xdb.WhereEq("id", 1))
// record["profile"] = map[string]any{"hobby": "coding"}
// record["role_ids"] = []int{1, 2, 3}
```

## 验证器

```go
m := xdb.New("users",
    xdb.ColumnValidator(
        xdb.Validate("name",
            xdb.Required(),
            xdb.Unique(xdb.WithMsg("名称已存在")),
        ),
        xdb.Validate("email",
            xdb.Required(),
        ),
    ),
)
```

## 关联关系

```go
m := xdb.New("users",
    // 一对一
    xdb.HasOne(xdb.HasOpts{
        Table:     "profiles",
        OtherKeys: []string{"bio", "avatar"},
    }),
    // 一对多
    xdb.HasMany(xdb.HasOpts{
        Table:      "orders",
        ForeignKey: "user_id",
        OtherKeys:  []string{"order_no", "amount"},
    }),
)
```

## 软删除

```go
m := xdb.New("users",
    xdb.WithFakeDelKey("is_deleted"),
)

// Delete 实际执行 UPDATE ... SET is_deleted = 1
// Select 自动添加 WHERE is_deleted = 0
```

## 缓存

```go
m := xdb.New("users",
    xdb.WithCacheKey("id", "email"), // 指定缓存键
)

// 查询会自动使用缓存
record, _ := m.First(xdb.WhereEq("id", 1))

// 更新/删除会自动清除缓存
m.Update(xdb.Record{"id": 1, "name": "new"})
```

## 多连接

```go
// 初始化多个连接
xdb.Init(map[string]*xdb.Config{
    "default": {Driver: "mysql", DSN: "..."},
    "readonly": {Driver: "mysql", DSN: "..."},
    "analytics": {Driver: "postgres", DSN: "..."},
})

// 使用指定连接
m := xdb.New("users", xdb.WithConnection("analytics"))
```

## 读写分离

```go
xdb.Init(map[string]*xdb.Config{
    "default": {
        Driver:  "mysql",
        DSN:     "user:pass@tcp(master:3306)/db",
        ReadDsn: "user:pass@tcp(slave:3306)/db", // 读库
    },
})

// 读操作自动使用 ReadDsn
// 写操作使用主库 DSN
```

## 原始 SQL

```go
// 执行原始 SQL
result, err := m.Exec("UPDATE users SET status = ? WHERE id = ?", 1, 100)

// 查询原始 SQL
rows, err := m.Query("SELECT * FROM users WHERE status = ?", 1)
```

## 方言接口

如需支持其他数据库，可实现 `Dialect` 接口：

```go
type Dialect interface {
    Name() string
    Placeholder(index int) string
    Placeholders(count int) string
    ConvertPlaceholders(sql string) string
    InsertReturning(sql string, primaryKey string) string
    SupportsLastInsertId() bool
    InsertIgnore(table string, fields []string, placeholders string) string
    Upsert(table string, fields []string, placeholders string,
           primaryKey string, updateFields []string) (sql string, needExtraValues bool)
    LimitOffset(limit, offset int) string
}
```

## 注意事项

### MySQL 特有函数

以下函数仅在 MySQL 中可用：

- `WhereFindInSet()` - 使用 MySQL 的 `FIND_IN_SET` 函数
- 如需跨数据库兼容，建议使用 JSON 列或关联表替代

### PostgreSQL 特性

- 不支持 `LastInsertId()`，xdb 自动使用 `RETURNING` 语句
- 占位符使用 `$1, $2, $3` 格式，xdb 自动转换

### SQLite 特性

- 使用 `INSERT OR IGNORE` 语法
- 使用小写的 `excluded` 表引用

## License

MIT
