package xadmin

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xhttp"
)

var (
	errEmailCodeDisabled      = errors.New("email verification code is disabled")
	errEmailSuffixNotAllowed  = errors.New("email suffix is not allowed")
	errEmailCodeInvalid       = errors.New("email verification code is invalid")
	errEmailCodeStoreFailed   = errors.New("email verification code store failed")
	errUserEmailRequired      = errors.New("user email is required")
	errEmailCodeTargetMissing = errors.New("username or email is required")
)

// EmailCodeSender sends a generated verification code to the target email.
type EmailCodeSender func(ctx context.Context, to, code string) error

// EmailCodeStore stores login email verification codes.
type EmailCodeStore interface {
	RemainingCooldown(ctx context.Context, email string, interval time.Duration, now time.Time) (time.Duration, error)
	Save(ctx context.Context, email, code string, codeExpire, sendInterval time.Duration, now time.Time) error
	Consume(ctx context.Context, email, code string, now time.Time) (bool, error)
}

type LoginEmailCodeConf struct {
	AllowedSuffixes []string
	CodeExpire      time.Duration
	SendInterval    time.Duration
	CodeLength      int
	Sender          EmailCodeSender
}

type emailCodeRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type emailCodeEntry struct {
	Code     string
	ExpireAt time.Time
	SentAt   time.Time
}

type emailCodeMemoryStore struct {
	mu    sync.Mutex
	codes map[string]emailCodeEntry
}

var (
	loginEmailCodeMu   sync.RWMutex
	loginEmailCodeConf *LoginEmailCodeConf
	emailCodeStoreMu   sync.RWMutex
	emailCodes         EmailCodeStore = newEmailCodeMemoryStore()
)

func SetLoginEmailCode(conf *LoginEmailCodeConf) {
	loginEmailCodeMu.Lock()
	defer loginEmailCodeMu.Unlock()
	loginEmailCodeConf = normalizeLoginEmailCodeConf(conf)
}

func SetLoginEmailCodeSender(sender EmailCodeSender) {
	SetLoginEmailCode(&LoginEmailCodeConf{Sender: sender})
}

func SetLoginEmailCodeStore(store EmailCodeStore) {
	emailCodeStoreMu.Lock()
	defer emailCodeStoreMu.Unlock()
	if store == nil {
		store = newEmailCodeMemoryStore()
	}
	emailCodes = store
}

func newEmailCodeMemoryStore() *emailCodeMemoryStore {
	return &emailCodeMemoryStore{codes: make(map[string]emailCodeEntry)}
}

func normalizeLoginEmailCodeConf(conf *LoginEmailCodeConf) *LoginEmailCodeConf {
	if conf == nil {
		return nil
	}

	copied := *conf
	copied.AllowedSuffixes = normalizeEmailSuffixes(conf.AllowedSuffixes)
	if copied.CodeExpire <= 0 {
		copied.CodeExpire = 5 * time.Minute
	}
	if copied.SendInterval <= 0 {
		copied.SendInterval = time.Minute
	}
	if copied.CodeLength <= 0 {
		copied.CodeLength = 6
	}
	return &copied
}

func getLoginEmailCodeConf() *LoginEmailCodeConf {
	loginEmailCodeMu.RLock()
	defer loginEmailCodeMu.RUnlock()
	if loginEmailCodeConf == nil {
		return nil
	}

	copied := *loginEmailCodeConf
	copied.AllowedSuffixes = append([]string(nil), loginEmailCodeConf.AllowedSuffixes...)
	return &copied
}

func getLoginEmailCodeStore() EmailCodeStore {
	emailCodeStoreMu.RLock()
	defer emailCodeStoreMu.RUnlock()
	return emailCodes
}

func loginEmailCodeEnabled(conf *LoginEmailCodeConf) bool {
	return conf != nil && conf.Sender != nil
}

func loginEmailCodeRequiredInCurrentEnv() bool {
	env := strings.ToLower(strings.TrimSpace(xapp.Args.AppEnv))
	return env == "prod" || env == "production"
}

func emailCodeHandler(w http.ResponseWriter, r *http.Request) {
	conf := getLoginEmailCodeConf()
	if !loginEmailCodeEnabled(conf) {
		emailCodeResponseError(w, 400, errEmailCodeDisabled.Error())
		return
	}

	req, err := xhttp.DecodeBody[emailCodeRequest](r)
	if err != nil {
		emailCodeResponseError(w, 400, "Invalid request")
		return
	}

	email, err := resolveEmailCodeTarget(req)
	if err != nil {
		emailCodeResponseError(w, 400, err.Error())
		return
	}
	if !emailAllowedBySuffix(email, conf.AllowedSuffixes) {
		emailCodeResponseError(w, 4004, errEmailSuffixNotAllowed.Error())
		return
	}

	now := time.Now()
	store := getLoginEmailCodeStore()
	remaining, err := store.RemainingCooldown(r.Context(), email, conf.SendInterval, now)
	if err != nil {
		emailCodeResponseError(w, 500, err.Error())
		return
	}
	if remaining > 0 {
		emailCodeResponseError(w, 4005, fmt.Sprintf("please retry after %d seconds", int(remaining.Seconds())+1))
		return
	}

	code, err := generateEmailCode(conf.CodeLength)
	if err != nil {
		emailCodeResponseError(w, 500, "Generate code failed")
		return
	}

	if err := conf.Sender(r.Context(), email, code); err != nil {
		emailCodeResponseError(w, 500, err.Error())
		return
	}

	if err := store.Save(r.Context(), email, code, conf.CodeExpire, conf.SendInterval, now); err != nil {
		emailCodeResponseError(w, 500, err.Error())
		return
	}
	xhttp.ResponseJson(w, Map{"code": 0})
}

func emailCodeStatusHandler(w http.ResponseWriter, r *http.Request) {
	xhttp.ResponseJson(w, Map{
		"code": 0,
		"data": Map{
			"enabled": loginEmailCodeEnabled(getLoginEmailCodeConf()),
		},
	})
}

func resolveEmailCodeTarget(req *emailCodeRequest) (string, error) {
	username := strings.TrimSpace(req.Username)
	if username != "" {
		row, err := xdb.New(operatorTable).Single(xdb.WhereEq("username", username))
		if err != nil {
			return "", err
		}
		if !row.GetBool("status") {
			return "", errors.New("user is disabled")
		}
		return normalizeUserEmail(row)
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		return "", errEmailCodeTargetMissing
	}
	return normalizeEmailAddress(email)
}

func validateLoginEmailCode(ctx context.Context, row xdb.Record, code string) error {
	conf := getLoginEmailCodeConf()
	if !loginEmailCodeEnabled(conf) {
		return nil
	}
	if !loginEmailCodeRequiredInCurrentEnv() {
		return nil
	}

	email, err := normalizeUserEmail(row)
	if err != nil {
		return err
	}
	if !emailAllowedBySuffix(email, conf.AllowedSuffixes) {
		return errEmailSuffixNotAllowed
	}
	ok, err := getLoginEmailCodeStore().Consume(ctx, email, strings.TrimSpace(code), time.Now())
	if err != nil {
		return fmt.Errorf("%w: %v", errEmailCodeStoreFailed, err)
	}
	if !ok {
		return errEmailCodeInvalid
	}
	return nil
}

func loginEmailCodeFromUser(user *User) string {
	if user == nil {
		return ""
	}
	if strings.TrimSpace(user.EmailCode) != "" {
		return user.EmailCode
	}
	return user.Code
}

func normalizeUserEmail(row xdb.Record) (string, error) {
	email, err := normalizeEmailAddress(row.GetString("email"))
	if err != nil {
		return "", errUserEmailRequired
	}
	return email, nil
}

func emailCodeErrorCode(err error) int {
	switch {
	case errors.Is(err, errEmailSuffixNotAllowed):
		return 4004
	case errors.Is(err, errUserEmailRequired):
		return 4007
	case errors.Is(err, errEmailCodeInvalid):
		return 4006
	case errors.Is(err, errEmailCodeStoreFailed):
		return 500
	default:
		return 400
	}
}

func emailCodeResponseError(w http.ResponseWriter, code int, message string) {
	xhttp.ResponseJson(w, Map{
		"code":    code,
		"message": message,
	})
}

func normalizeEmailAddress(raw string) (string, error) {
	email := strings.TrimSpace(raw)
	if email == "" {
		return "", errors.New("email is required")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(addr.Address, email) {
		return "", errors.New("email must not include display name")
	}
	return strings.ToLower(addr.Address), nil
}

func normalizeEmailSuffixes(suffixes []string) []string {
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(suffixes))
	for _, suffix := range suffixes {
		suffix = strings.ToLower(strings.TrimSpace(suffix))
		if suffix == "" {
			continue
		}
		if !strings.HasPrefix(suffix, "@") {
			suffix = "@" + suffix
		}
		if _, ok := seen[suffix]; ok {
			continue
		}
		seen[suffix] = struct{}{}
		normalized = append(normalized, suffix)
	}
	return normalized
}

func emailAllowedBySuffix(email string, suffixes []string) bool {
	if len(suffixes) == 0 {
		return true
	}
	email = strings.ToLower(email)
	for _, suffix := range suffixes {
		if strings.HasSuffix(email, suffix) {
			return true
		}
	}
	return false
}

func generateEmailCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	var b strings.Builder
	b.Grow(length)
	max := big.NewInt(10)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String(), nil
}

func (s *emailCodeMemoryStore) RemainingCooldown(ctx context.Context, email string, interval time.Duration, now time.Time) (time.Duration, error) {
	if interval <= 0 {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.codes[email]
	if !ok {
		return 0, nil
	}

	remaining := entry.SentAt.Add(interval).Sub(now)
	if remaining <= 0 {
		return 0, nil
	}
	return remaining, nil
}

func (s *emailCodeMemoryStore) Save(ctx context.Context, email, code string, codeExpire, sendInterval time.Duration, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[email] = emailCodeEntry{
		Code:     code,
		ExpireAt: now.Add(codeExpire),
		SentAt:   now,
	}
	return nil
}

func (s *emailCodeMemoryStore) Consume(ctx context.Context, email, code string, now time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.codes[email]
	if !ok {
		return false, nil
	}
	if now.After(entry.ExpireAt) {
		delete(s.codes, email)
		return false, nil
	}
	if subtle.ConstantTimeCompare([]byte(entry.Code), []byte(code)) != 1 {
		return false, nil
	}

	delete(s.codes, email)
	return true, nil
}
