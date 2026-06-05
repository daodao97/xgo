package xadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daodao97/xgo/xdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAPIResp struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
}

func resetEmailCodeTest(t *testing.T) {
	t.Helper()

	oldLoginEmailCodeConf := loginEmailCodeConf
	oldEmailCodes := emailCodes

	loginEmailCodeMu.Lock()
	loginEmailCodeConf = nil
	loginEmailCodeMu.Unlock()
	emailCodes = newEmailCodeMemoryStore()

	t.Cleanup(func() {
		loginEmailCodeMu.Lock()
		loginEmailCodeConf = oldLoginEmailCodeConf
		loginEmailCodeMu.Unlock()
		emailCodes = oldEmailCodes
	})
}

func postJSON(handler http.HandlerFunc, body string) testAPIResp {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	var resp testAPIResp
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

func TestEmailSuffixValidation(t *testing.T) {
	conf := normalizeLoginEmailCodeConf(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"Example.COM", "@TEAM.org", " @team.org "},
	})

	assert.Equal(t, []string{"@example.com", "@team.org"}, conf.AllowedSuffixes)

	email, err := normalizeEmailAddress("USER@Example.COM")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", email)
	assert.True(t, emailAllowedBySuffix(email, conf.AllowedSuffixes))
	assert.True(t, emailAllowedBySuffix("member@team.org", conf.AllowedSuffixes))
	assert.False(t, emailAllowedBySuffix("member@other.org", conf.AllowedSuffixes))
	assert.True(t, emailAllowedBySuffix("member@other.org", nil))

	_, err = normalizeEmailAddress("not-email")
	assert.Error(t, err)
	_, err = normalizeEmailAddress("User <user@example.com>")
	assert.Error(t, err)
}

func TestEmailCodeHandlerUsesInjectedSenderAndEnforcesCooldown(t *testing.T) {
	resetEmailCodeTest(t)

	var sentTo string
	var sentCode string
	sendCount := 0
	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
		SendInterval:    time.Minute,
		Sender: func(ctx context.Context, to, code string) error {
			sendCount++
			sentTo = to
			sentCode = code
			return nil
		},
	})

	resp := postJSON(emailCodeHandler, `{"email":"User@Example.COM"}`)
	require.Equal(t, 0, resp.Code)
	assert.Equal(t, "user@example.com", sentTo)
	assert.Len(t, sentCode, 6)
	assert.Equal(t, 1, sendCount)

	resp = postJSON(emailCodeHandler, `{"email":"user@example.com"}`)
	assert.Equal(t, 4005, resp.Code)
	assert.Equal(t, 1, sendCount)

	resp = postJSON(emailCodeHandler, `{"email":"user@other.com"}`)
	assert.Equal(t, 4004, resp.Code)

	SetLoginEmailCode(nil)
	resp = postJSON(emailCodeHandler, `{"email":"user@example.com"}`)
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, errEmailCodeDisabled.Error(), resp.Message)
}

func TestLoginEmailCodeOptionalWithoutSender(t *testing.T) {
	resetEmailCodeTest(t)

	assert.NoError(t, validateLoginEmailCode(xdb.Record{}, ""))

	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
	})
	assert.NoError(t, validateLoginEmailCode(xdb.Record{}, ""))
}

func TestLoginEmailCodeRequiredWhenSenderRegistered(t *testing.T) {
	resetEmailCodeTest(t)

	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
		Sender: func(ctx context.Context, to, code string) error {
			return nil
		},
	})

	row := xdb.Record{"email": "USER@Example.COM"}
	emailCodes.save("user@example.com", "123456", time.Now().Add(time.Minute), time.Now())

	err := validateLoginEmailCode(row, "")
	require.ErrorIs(t, err, errEmailCodeInvalid)

	err = validateLoginEmailCode(row, "000000")
	require.ErrorIs(t, err, errEmailCodeInvalid)

	err = validateLoginEmailCode(row, "123456")
	require.NoError(t, err)

	err = validateLoginEmailCode(row, "123456")
	require.ErrorIs(t, err, errEmailCodeInvalid)
}

func TestLoginEmailCodeRejectsMissingAndDisallowedEmail(t *testing.T) {
	resetEmailCodeTest(t)

	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
		Sender: func(ctx context.Context, to, code string) error {
			return nil
		},
	})

	err := validateLoginEmailCode(xdb.Record{}, "123456")
	require.ErrorIs(t, err, errUserEmailRequired)

	err = validateLoginEmailCode(xdb.Record{"email": "user@other.com"}, "123456")
	require.ErrorIs(t, err, errEmailSuffixNotAllowed)
}
