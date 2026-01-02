package xdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDialect(t *testing.T) {
	tests := []struct {
		driver   string
		expected string
	}{
		{"mysql", "mysql"},
		{"", "mysql"}, // 默认是 MySQL
		{"postgres", "postgres"},
		{"postgresql", "postgres"},
		{"pgx", "postgres"},
		{"sqlite", "sqlite"},
		{"sqlite3", "sqlite"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			dialect := GetDialect(tt.driver)
			assert.Equal(t, tt.expected, dialect.Name())
		})
	}
}

func TestMySQLDialect(t *testing.T) {
	dialect := &MySQLDialect{}

	t.Run("Placeholder", func(t *testing.T) {
		assert.Equal(t, "?", dialect.Placeholder(1))
		assert.Equal(t, "?", dialect.Placeholder(5))
	})

	t.Run("Placeholders", func(t *testing.T) {
		assert.Equal(t, "", dialect.Placeholders(0))
		assert.Equal(t, "?", dialect.Placeholders(1))
		assert.Equal(t, "?, ?", dialect.Placeholders(2))
		assert.Equal(t, "?, ?, ?", dialect.Placeholders(3))
	})

	t.Run("ConvertPlaceholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ? AND name = ?"
		assert.Equal(t, sql, dialect.ConvertPlaceholders(sql))
	})

	t.Run("InsertReturning", func(t *testing.T) {
		sql := "INSERT INTO users (name) VALUES (?)"
		assert.Equal(t, sql, dialect.InsertReturning(sql, "id"))
	})

	t.Run("SupportsLastInsertId", func(t *testing.T) {
		assert.True(t, dialect.SupportsLastInsertId())
	})

	t.Run("InsertIgnore", func(t *testing.T) {
		sql := dialect.InsertIgnore("users", []string{"id", "name"}, "?, ?")
		assert.Equal(t, "INSERT IGNORE INTO users (id, name) VALUES (?, ?)", sql)
	})

	t.Run("Upsert", func(t *testing.T) {
		sql, needExtra := dialect.Upsert("users", []string{"id", "name", "email"}, "?, ?, ?", "id", []string{"name", "email"})
		assert.Equal(t, "INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name = ?, email = ?", sql)
		assert.True(t, needExtra)
	})
}

func TestPostgreSQLDialect(t *testing.T) {
	dialect := &PostgreSQLDialect{}

	t.Run("Placeholder", func(t *testing.T) {
		assert.Equal(t, "$1", dialect.Placeholder(1))
		assert.Equal(t, "$5", dialect.Placeholder(5))
	})

	t.Run("Placeholders", func(t *testing.T) {
		assert.Equal(t, "", dialect.Placeholders(0))
		assert.Equal(t, "$1", dialect.Placeholders(1))
		assert.Equal(t, "$1, $2", dialect.Placeholders(2))
		assert.Equal(t, "$1, $2, $3", dialect.Placeholders(3))
	})

	t.Run("ConvertPlaceholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ? AND name = ?"
		expected := "SELECT * FROM users WHERE id = $1 AND name = $2"
		assert.Equal(t, expected, dialect.ConvertPlaceholders(sql))
	})

	t.Run("InsertReturning", func(t *testing.T) {
		sql := "INSERT INTO users (name) VALUES ($1)"
		expected := "INSERT INTO users (name) VALUES ($1) RETURNING id"
		assert.Equal(t, expected, dialect.InsertReturning(sql, "id"))
	})

	t.Run("SupportsLastInsertId", func(t *testing.T) {
		assert.False(t, dialect.SupportsLastInsertId())
	})

	t.Run("InsertIgnore", func(t *testing.T) {
		sql := dialect.InsertIgnore("users", []string{"id", "name"}, "$1, $2")
		assert.Equal(t, "INSERT INTO users (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING", sql)
	})

	t.Run("Upsert", func(t *testing.T) {
		sql, needExtra := dialect.Upsert("users", []string{"id", "name", "email"}, "$1, $2, $3", "id", []string{"name", "email"})
		assert.Equal(t, "INSERT INTO users (id, name, email) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email", sql)
		assert.False(t, needExtra)
	})
}

func TestSQLiteDialect(t *testing.T) {
	dialect := &SQLiteDialect{}

	t.Run("Placeholder", func(t *testing.T) {
		assert.Equal(t, "?", dialect.Placeholder(1))
		assert.Equal(t, "?", dialect.Placeholder(5))
	})

	t.Run("Placeholders", func(t *testing.T) {
		assert.Equal(t, "", dialect.Placeholders(0))
		assert.Equal(t, "?", dialect.Placeholders(1))
		assert.Equal(t, "?, ?", dialect.Placeholders(2))
		assert.Equal(t, "?, ?, ?", dialect.Placeholders(3))
	})

	t.Run("ConvertPlaceholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ? AND name = ?"
		assert.Equal(t, sql, dialect.ConvertPlaceholders(sql))
	})

	t.Run("SupportsLastInsertId", func(t *testing.T) {
		assert.True(t, dialect.SupportsLastInsertId())
	})

	t.Run("InsertIgnore", func(t *testing.T) {
		sql := dialect.InsertIgnore("users", []string{"id", "name"}, "?, ?")
		assert.Equal(t, "INSERT OR IGNORE INTO users (id, name) VALUES (?, ?)", sql)
	})

	t.Run("Upsert", func(t *testing.T) {
		sql, needExtra := dialect.Upsert("users", []string{"id", "name", "email"}, "?, ?, ?", "id", []string{"name", "email"})
		assert.Equal(t, "INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON CONFLICT (id) DO UPDATE SET name = excluded.name, email = excluded.email", sql)
		assert.False(t, needExtra)
	})
}

func TestIsPostgres(t *testing.T) {
	assert.True(t, IsPostgres("postgres"))
	assert.True(t, IsPostgres("postgresql"))
	assert.True(t, IsPostgres("pgx"))
	assert.False(t, IsPostgres("mysql"))
	assert.False(t, IsPostgres("sqlite"))
}

func TestIsSQLite(t *testing.T) {
	assert.True(t, IsSQLite("sqlite"))
	assert.True(t, IsSQLite("sqlite3"))
	assert.False(t, IsSQLite("mysql"))
	assert.False(t, IsSQLite("postgres"))
}
