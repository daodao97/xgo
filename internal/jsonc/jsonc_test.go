package jsonc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToJSONRemovesCommentsAndTrailingCommas(t *testing.T) {
	input := []byte(`{
		// line comment
		"name": "http://example.com/#keep",
		"items": [1, 2,],
		/* block
		   comment */
		"ok": true,
		# hash comment
	}`)

	cleaned := ToJSON(input)
	require.True(t, json.Valid(cleaned), string(cleaned))

	var got map[string]any
	require.NoError(t, json.Unmarshal(cleaned, &got))
	require.Equal(t, "http://example.com/#keep", got["name"])
	require.Equal(t, true, got["ok"])
}

func TestValidRejectsInvalidJSONC(t *testing.T) {
	require.False(t, Valid([]byte(`{"name":}`)))
}
