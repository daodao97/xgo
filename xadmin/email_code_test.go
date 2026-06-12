package xadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

type emailCodeStatusResp struct {
	Code int `json:"code"`
	Data struct {
		Enabled bool `json:"enabled"`
	} `json:"data"`
}

type emailCodeSpyStore struct {
	*emailCodeMemoryStore
	saveCount int
}

func (s *emailCodeSpyStore) Save(ctx context.Context, email, code string, codeExpire, sendInterval time.Duration, now time.Time) error {
	s.saveCount++
	return s.emailCodeMemoryStore.Save(ctx, email, code, codeExpire, sendInterval, now)
}

func resetEmailCodeTest(t *testing.T) {
	t.Helper()

	oldLoginEmailCodeConf := loginEmailCodeConf
	oldEmailCodes := emailCodes

	loginEmailCodeMu.Lock()
	loginEmailCodeConf = nil
	loginEmailCodeMu.Unlock()
	SetLoginEmailCodeStore(newEmailCodeMemoryStore())

	t.Cleanup(func() {
		loginEmailCodeMu.Lock()
		loginEmailCodeConf = oldLoginEmailCodeConf
		loginEmailCodeMu.Unlock()
		SetLoginEmailCodeStore(oldEmailCodes)
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

func getEmailCodeStatus() emailCodeStatusResp {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	emailCodeStatusHandler(w, req)

	var resp emailCodeStatusResp
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

func TestEmailCodeStatusReflectsSenderRegistration(t *testing.T) {
	resetEmailCodeTest(t)

	resp := getEmailCodeStatus()
	require.Equal(t, 0, resp.Code)
	assert.False(t, resp.Data.Enabled)

	SetLoginEmailCode(&LoginEmailCodeConf{
		Sender: func(ctx context.Context, to, code string) error {
			return nil
		},
	})

	resp = getEmailCodeStatus()
	require.Equal(t, 0, resp.Code)
	assert.True(t, resp.Data.Enabled)
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

func TestSetLoginEmailCodeStoreUsesRegisteredStore(t *testing.T) {
	resetEmailCodeTest(t)

	store := &emailCodeSpyStore{emailCodeMemoryStore: newEmailCodeMemoryStore()}
	SetLoginEmailCodeStore(store)
	SetLoginEmailCode(&LoginEmailCodeConf{
		Sender: func(ctx context.Context, to, code string) error {
			return nil
		},
	})

	resp := postJSON(emailCodeHandler, `{"email":"user@example.com"}`)
	require.Equal(t, 0, resp.Code)
	assert.Equal(t, 1, store.saveCount)

	SetLoginEmailCodeStore(nil)
	_, ok := getLoginEmailCodeStore().(*emailCodeMemoryStore)
	assert.True(t, ok)
}

func TestLoginEmailCodeOptionalWithoutSender(t *testing.T) {
	resetEmailCodeTest(t)

	assert.NoError(t, validateLoginEmailCode(context.Background(), xdb.Record{}, ""))

	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
	})
	assert.NoError(t, validateLoginEmailCode(context.Background(), xdb.Record{}, ""))
}

func TestLoginEmailCodeRequiredWhenSenderRegistered(t *testing.T) {
	resetEmailCodeTest(t)
	t.Setenv("APP_ENV", "prod")

	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
		Sender: func(ctx context.Context, to, code string) error {
			return nil
		},
	})

	row := xdb.Record{"email": "USER@Example.COM"}
	require.NoError(t, getLoginEmailCodeStore().Save(context.Background(), "user@example.com", "123456", time.Minute, time.Minute, time.Now()))

	err := validateLoginEmailCode(context.Background(), row, "")
	require.ErrorIs(t, err, errEmailCodeInvalid)

	err = validateLoginEmailCode(context.Background(), row, "000000")
	require.ErrorIs(t, err, errEmailCodeInvalid)

	err = validateLoginEmailCode(context.Background(), row, "123456")
	require.NoError(t, err)

	err = validateLoginEmailCode(context.Background(), row, "123456")
	require.ErrorIs(t, err, errEmailCodeInvalid)
}

func TestLoginEmailCodeBypassedOutsideProduction(t *testing.T) {
	resetEmailCodeTest(t)
	t.Setenv("APP_ENV", "dev")

	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
		Sender: func(ctx context.Context, to, code string) error {
			return nil
		},
	})

	err := validateLoginEmailCode(context.Background(), xdb.Record{}, "anything")
	require.NoError(t, err)

	err = validateLoginEmailCode(context.Background(), xdb.Record{"email": "user@other.com"}, "wrong")
	require.NoError(t, err)
}

func TestLoginEmailCodeProductionEnvAliases(t *testing.T) {
	t.Setenv("APP_ENV", " production ")
	assert.True(t, loginEmailCodeRequiredInCurrentEnv())

	t.Setenv("APP_ENV", "PROD")
	assert.True(t, loginEmailCodeRequiredInCurrentEnv())

	t.Setenv("APP_ENV", "pre")
	assert.False(t, loginEmailCodeRequiredInCurrentEnv())
}

func TestLoginEmailCodeEnvFromAppEnvArg(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	t.Setenv("APP_ENV", "")
	os.Args = []string{"app", "--app-env=prod"}
	assert.True(t, loginEmailCodeRequiredInCurrentEnv())

	os.Args = []string{"app", "--app-env", "production"}
	assert.True(t, loginEmailCodeRequiredInCurrentEnv())

	os.Args = []string{"app", "--app-env", "test"}
	assert.False(t, loginEmailCodeRequiredInCurrentEnv())
}

func TestLoginEmailCodeRejectsMissingAndDisallowedEmail(t *testing.T) {
	resetEmailCodeTest(t)
	t.Setenv("APP_ENV", "prod")

	SetLoginEmailCode(&LoginEmailCodeConf{
		AllowedSuffixes: []string{"@example.com"},
		Sender: func(ctx context.Context, to, code string) error {
			return nil
		},
	})

	err := validateLoginEmailCode(context.Background(), xdb.Record{}, "123456")
	require.ErrorIs(t, err, errUserEmailRequired)

	err = validateLoginEmailCode(context.Background(), xdb.Record{"email": "user@other.com"}, "123456")
	require.ErrorIs(t, err, errEmailSuffixNotAllowed)
}
