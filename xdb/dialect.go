package xdb

import (
	"fmt"
	"strings"
	"sync"
)

// Dialect 定义了数据库方言接口，用于处理不同数据库的 SQL 语法差异
type Dialect interface {
	// Name 返回方言名称
	Name() string

	// Placeholder 返回参数占位符
	// index 是从 1 开始的参数索引
	// MySQL/SQLite 返回 "?", PostgreSQL 返回 "$1", "$2" 等
	Placeholder(index int) string

	// Placeholders 返回多个占位符，用逗号分隔
	Placeholders(count int) string

	// ConvertPlaceholders 将 SQL 中的 ? 占位符转换为对应数据库的格式
	ConvertPlaceholders(sql string) string

	// InsertReturning 返回带 RETURNING 子句的 INSERT SQL
	// 对于 MySQL 返回原始 SQL，PostgreSQL 返回带 RETURNING 的 SQL
	InsertReturning(sql string, primaryKey string) string

	// SupportsLastInsertId 返回数据库是否支持 LastInsertId
	SupportsLastInsertId() bool

	// InsertIgnore 返回忽略冲突的 INSERT SQL
	InsertIgnore(table string, fields []string, placeholders string) string

	// Upsert 返回 UPSERT SQL（插入或更新）
	// fields: 所有字段
	// placeholders: 值占位符
	// primaryKey: 主键字段
	// updateFields: 需要更新的字段
	// updateValues: 用于 MySQL 的更新值占位符
	Upsert(table string, fields []string, placeholders string, primaryKey string, updateFields []string) (sql string, needExtraValues bool)

	// LimitOffset 返回分页 SQL 片段
	LimitOffset(limit, offset int) string
}

type identifierQuoter interface {
	QuoteIdentifier(identifier string) string
}

type forUpdateDialect interface {
	ForUpdate() string
}

func forUpdateSQL(dialect Dialect) string {
	if dialect == nil {
		return " for update"
	}
	if d, ok := dialect.(forUpdateDialect); ok {
		return d.ForUpdate()
	}
	return " for update"
}

// 确保所有实现都满足 Dialect 接口
var (
	_ Dialect = (*MySQLDialect)(nil)
	_ Dialect = (*PostgreSQLDialect)(nil)
	_ Dialect = (*SQLiteDialect)(nil)
)

func convertQuestionPlaceholders(sql string, placeholder func(index int) string) string {
	if sql == "" {
		return sql
	}

	var builder strings.Builder
	builder.Grow(len(sql))
	argIndex := 1

	for i := 0; i < len(sql); i++ {
		switch sql[i] {
		case '\'':
			i = copyQuotedSQL(&builder, sql, i, '\'')
		case '"':
			i = copyQuotedSQL(&builder, sql, i, '"')
		case '`':
			i = copyQuotedSQL(&builder, sql, i, '`')
		case '-':
			if i+1 < len(sql) && sql[i+1] == '-' {
				i = copyLineComment(&builder, sql, i)
			} else {
				builder.WriteByte(sql[i])
			}
		case '/':
			if i+1 < len(sql) && sql[i+1] == '*' {
				i = copyBlockComment(&builder, sql, i)
			} else {
				builder.WriteByte(sql[i])
			}
		case '?':
			if i+1 < len(sql) && sql[i+1] == '?' {
				builder.WriteByte('?')
				i++
				continue
			}
			builder.WriteString(placeholder(argIndex))
			argIndex++
		default:
			builder.WriteByte(sql[i])
		}
	}

	return builder.String()
}

func copyQuotedSQL(builder *strings.Builder, sql string, start int, quote byte) int {
	builder.WriteByte(sql[start])
	for i := start + 1; i < len(sql); i++ {
		builder.WriteByte(sql[i])
		if sql[i] == quote {
			if i+1 < len(sql) && sql[i+1] == quote {
				i++
				builder.WriteByte(sql[i])
				continue
			}
			if quote != '`' && i > 0 && sql[i-1] == '\\' {
				continue
			}
			return i
		}
	}
	return len(sql) - 1
}

func copyLineComment(builder *strings.Builder, sql string, start int) int {
	builder.WriteString("--")
	for i := start + 2; i < len(sql); i++ {
		builder.WriteByte(sql[i])
		if sql[i] == '\n' {
			return i
		}
	}
	return len(sql) - 1
}

func copyBlockComment(builder *strings.Builder, sql string, start int) int {
	builder.WriteString("/*")
	for i := start + 2; i < len(sql); i++ {
		builder.WriteByte(sql[i])
		if sql[i] == '*' && i+1 < len(sql) && sql[i+1] == '/' {
			i++
			builder.WriteByte(sql[i])
			return i
		}
	}
	return len(sql) - 1
}

func quoteIdentifierWith(identifier, quote string) string {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" || identifier == "*" {
		return identifier
	}
	if isQuotedIdentifier(identifier) {
		return identifier
	}
	parts := strings.Split(identifier, ".")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "*" || isQuotedIdentifier(part) {
			parts[i] = part
			continue
		}
		parts[i] = quote + strings.ReplaceAll(part, quote, quote+quote) + quote
	}
	return strings.Join(parts, ".")
}

func isQuotedIdentifier(identifier string) bool {
	if len(identifier) < 2 {
		return false
	}
	switch {
	case identifier[0] == '`' && identifier[len(identifier)-1] == '`':
		return true
	case identifier[0] == '"' && identifier[len(identifier)-1] == '"':
		return true
	case identifier[0] == '[' && identifier[len(identifier)-1] == ']':
		return true
	default:
		return false
	}
}

// MySQLDialect MySQL 方言实现
type MySQLDialect struct{}

func (d *MySQLDialect) Name() string {
	return "mysql"
}

func (d *MySQLDialect) Placeholder(index int) string {
	return "?"
}

func (d *MySQLDialect) Placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat("?, ", count-1) + "?"
}

func (d *MySQLDialect) ConvertPlaceholders(sql string) string {
	return sql // MySQL 使用 ? 占位符，无需转换
}

func (d *MySQLDialect) QuoteIdentifier(identifier string) string {
	return quoteIdentifierWith(identifier, "`")
}

func (d *MySQLDialect) InsertReturning(sql string, primaryKey string) string {
	return sql // MySQL 不需要 RETURNING
}

func (d *MySQLDialect) SupportsLastInsertId() bool {
	return true
}

func (d *MySQLDialect) InsertIgnore(table string, fields []string, placeholders string) string {
	return fmt.Sprintf("INSERT IGNORE INTO %s (%s) VALUES (%s)",
		table, strings.Join(fields, ", "), placeholders)
}

func (d *MySQLDialect) Upsert(table string, fields []string, placeholders string, primaryKey string, updateFields []string) (string, bool) {
	// 构建更新子句，MySQL 需要 field = ? 格式
	var updates []string
	for _, field := range updateFields {
		updates = append(updates, fmt.Sprintf("%s = ?", field))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		table, strings.Join(fields, ", "), placeholders, strings.Join(updates, ", "))

	return sql, true // MySQL 需要额外的更新值参数
}

func (d *MySQLDialect) LimitOffset(limit, offset int) string {
	return " limit ? offset ?"
}

func (d *MySQLDialect) ForUpdate() string {
	return " for update"
}

// PostgreSQLDialect PostgreSQL 方言实现
type PostgreSQLDialect struct{}

func (d *PostgreSQLDialect) Name() string {
	return "postgres"
}

func (d *PostgreSQLDialect) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (d *PostgreSQLDialect) Placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(placeholders, ", ")
}

func (d *PostgreSQLDialect) ConvertPlaceholders(sql string) string {
	return convertQuestionPlaceholders(sql, d.Placeholder)
}

func (d *PostgreSQLDialect) QuoteIdentifier(identifier string) string {
	return quoteIdentifierWith(identifier, `"`)
}

func (d *PostgreSQLDialect) InsertReturning(sql string, primaryKey string) string {
	return sql + " RETURNING " + primaryKey
}

func (d *PostgreSQLDialect) SupportsLastInsertId() bool {
	return false
}

func (d *PostgreSQLDialect) InsertIgnore(table string, fields []string, placeholders string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
		table, strings.Join(fields, ", "), placeholders)
}

func (d *PostgreSQLDialect) Upsert(table string, fields []string, placeholders string, primaryKey string, updateFields []string) (string, bool) {
	// PostgreSQL 使用 EXCLUDED 表引用新值
	var updates []string
	for _, field := range updateFields {
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", field, field))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		table, strings.Join(fields, ", "), placeholders, primaryKey, strings.Join(updates, ", "))

	return sql, false // PostgreSQL 使用 EXCLUDED，不需要额外值
}

func (d *PostgreSQLDialect) LimitOffset(limit, offset int) string {
	return " limit ? offset ?"
}

func (d *PostgreSQLDialect) ForUpdate() string {
	return " for update"
}

// SQLiteDialect SQLite 方言实现
type SQLiteDialect struct{}

func (d *SQLiteDialect) Name() string {
	return "sqlite"
}

func (d *SQLiteDialect) Placeholder(index int) string {
	return "?"
}

func (d *SQLiteDialect) Placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat("?, ", count-1) + "?"
}

func (d *SQLiteDialect) ConvertPlaceholders(sql string) string {
	return sql // SQLite 使用 ? 占位符，无需转换
}

func (d *SQLiteDialect) QuoteIdentifier(identifier string) string {
	return quoteIdentifierWith(identifier, `"`)
}

func (d *SQLiteDialect) InsertReturning(sql string, primaryKey string) string {
	return sql // SQLite 3.35+ 支持 RETURNING，但为兼容性不使用
}

func (d *SQLiteDialect) SupportsLastInsertId() bool {
	return true
}

func (d *SQLiteDialect) InsertIgnore(table string, fields []string, placeholders string) string {
	return fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES (%s)",
		table, strings.Join(fields, ", "), placeholders)
}

func (d *SQLiteDialect) Upsert(table string, fields []string, placeholders string, primaryKey string, updateFields []string) (string, bool) {
	// SQLite 使用 excluded 表引用新值（小写）
	var updates []string
	for _, field := range updateFields {
		updates = append(updates, fmt.Sprintf("%s = excluded.%s", field, field))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		table, strings.Join(fields, ", "), placeholders, primaryKey, strings.Join(updates, ", "))

	return sql, false // SQLite 使用 excluded，不需要额外值
}

func (d *SQLiteDialect) LimitOffset(limit, offset int) string {
	return " limit ? offset ?"
}

func (d *SQLiteDialect) ForUpdate() string {
	return ""
}

// 方言实例（单例模式）
var (
	dialectMySQL      = &MySQLDialect{}
	dialectPostgreSQL = &PostgreSQLDialect{}
	dialectSQLite     = &SQLiteDialect{}

	dialectsMu sync.RWMutex
	dialects   = map[string]Dialect{
		"":           dialectMySQL,
		"mysql":      dialectMySQL,
		"postgres":   dialectPostgreSQL,
		"postgresql": dialectPostgreSQL,
		"pgx":        dialectPostgreSQL,
		"sqlite":     dialectSQLite,
		"sqlite3":    dialectSQLite,
	}
)

func RegisterDialect(driver string, dialect Dialect) {
	if driver == "" || dialect == nil {
		return
	}
	dialectsMu.Lock()
	defer dialectsMu.Unlock()
	dialects[strings.ToLower(driver)] = dialect
}

// GetDialect 根据驱动名称获取对应的方言实现
func GetDialect(driver string) Dialect {
	dialectsMu.RLock()
	dialect, ok := dialects[strings.ToLower(driver)]
	dialectsMu.RUnlock()
	if ok {
		return dialect
	}
	return dialectMySQL
}

// IsPostgres 检查是否为 PostgreSQL 驱动
func IsPostgres(driver string) bool {
	switch driver {
	case "postgres", "postgresql", "pgx":
		return true
	default:
		return false
	}
}

// IsSQLite 检查是否为 SQLite 驱动
func IsSQLite(driver string) bool {
	switch driver {
	case "sqlite", "sqlite3":
		return true
	default:
		return false
	}
}
