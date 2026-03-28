package xdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSoftDeleteWhereOptions(t *testing.T) {
	t.Run("fake delete only", func(t *testing.T) {
		m := &model{fakeDelKey: "is_deleted"}

		opts := m.softDeleteWhereOptions()
		require.Len(t, opts, 1)

		parsed := &Options{}
		opts[0](parsed)
		require.Len(t, parsed.where, 1)
		assert.Equal(t, "is_deleted", parsed.where[0].field)
		assert.Equal(t, "=", parsed.where[0].operator)
		assert.Equal(t, 0, parsed.where[0].value)
	})

	t.Run("deleted at only", func(t *testing.T) {
		m := &model{deletedAtKey: "deleted_at"}

		opts := m.softDeleteWhereOptions()
		require.Len(t, opts, 1)

		parsed := &Options{}
		opts[0](parsed)
		require.Len(t, parsed.where, 1)
		assert.Equal(t, "deleted_at is null", parsed.where[0].raw)
	})

	t.Run("both soft delete styles", func(t *testing.T) {
		m := &model{
			fakeDelKey:   "is_deleted",
			deletedAtKey: "deleted_at",
		}

		opts := m.softDeleteWhereOptions()
		require.Len(t, opts, 2)

		parsed := &Options{}
		for _, opt := range opts {
			opt(parsed)
		}

		require.Len(t, parsed.where, 2)
		assert.Equal(t, "is_deleted", parsed.where[0].field)
		assert.Equal(t, 0, parsed.where[0].value)
		assert.Equal(t, "deleted_at is null", parsed.where[1].raw)
	})
}

func TestSoftDeleteRecord(t *testing.T) {
	t.Run("fake delete only", func(t *testing.T) {
		m := &model{fakeDelKey: "is_deleted"}

		record := m.softDeleteRecord()
		assert.Equal(t, 1, record["is_deleted"])
	})

	t.Run("deleted at only", func(t *testing.T) {
		m := &model{deletedAtKey: "deleted_at"}

		before := time.Now()
		record := m.softDeleteRecord()
		after := time.Now()

		value, ok := record["deleted_at"]
		require.True(t, ok)

		deletedAt, ok := value.(time.Time)
		require.True(t, ok)
		assert.False(t, deletedAt.Before(before))
		assert.False(t, deletedAt.After(after))
	})

	t.Run("both soft delete styles", func(t *testing.T) {
		m := &model{
			fakeDelKey:   "is_deleted",
			deletedAtKey: "deleted_at",
		}

		record := m.softDeleteRecord()
		assert.Equal(t, 1, record["is_deleted"])
		_, ok := record["deleted_at"]
		assert.True(t, ok)
	})
}
