package xnotify

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/daodao97/xgo/xrequest"
)

const barkDefaultBaseURL = "https://api.day.app"

type barkSender struct{}

type barkPayload struct {
	Title      string `json:"title,omitempty"`
	Subtitle   string `json:"subtitle,omitempty"`
	Body       string `json:"body,omitempty"`
	DeviceKey  string `json:"device_key"`
	Level      string `json:"level,omitempty"`
	Volume     string `json:"volume,omitempty"`
	Badge      *int   `json:"badge,omitempty"`
	Call       string `json:"call,omitempty"`
	AutoCopy   string `json:"autoCopy,omitempty"`
	Copy       string `json:"copy,omitempty"`
	Sound      string `json:"sound,omitempty"`
	Icon       string `json:"icon,omitempty"`
	Group      string `json:"group,omitempty"`
	Ciphertext string `json:"ciphertext,omitempty"`
	IsArchive  string `json:"isArchive,omitempty"`
	URL        string `json:"url,omitempty"`
	Action     string `json:"action,omitempty"`
}

func (barkSender) Send(ctx context.Context, botID, message string, mentions []string) error {
	return barkSender{}.SendWithOptions(ctx, botID, message, mentions, NotifyOptions{})
}

func (barkSender) SendWithOptions(ctx context.Context, botID, message string, _ []string, options NotifyOptions) error {
	if options.MessageType != "" && options.MessageType != MessageTypeText {
		return fmt.Errorf("unsupported bark message type: %s", options.MessageType)
	}
	return SendBarkWithOptions(ctx, botID, message, options)
}

func init() {
	if err := RegisterProvider(BarkScheme, barkSender{}); err != nil {
		panic(err)
	}
}

// SendBarkText sends a text notification to Bark.
//
// Supported botID formats:
// - "{device_key}" uses the official Bark host https://api.day.app
// - "{device_key1,device_key2}" sends one request per device key to the official Bark host
// - "{host}/{device_key}" uses https://{host}
// - "{host}/{device_key1,device_key2}" sends one request per device key to https://{host}
// - "{http(s)://host[:port][/base-path]}/{device_key}" uses the explicit self-hosted address
// - "{http(s)://host[:port][/base-path]}/{device_key1,device_key2}" sends one request per device key
func SendBarkText(ctx context.Context, botID, content string) error {
	return SendBarkWithOptions(ctx, botID, content, NotifyOptions{})
}

// SendBarkWithOptions sends a Bark notification with optional rich fields.
func SendBarkWithOptions(ctx context.Context, botID, content string, options NotifyOptions) error {
	baseURL, deviceKeys, err := parseBarkTarget(botID)
	if err != nil {
		return err
	}

	var sendErrs []error
	for _, deviceKey := range deviceKeys {
		if ctx != nil && ctx.Err() != nil {
			sendErrs = append(sendErrs, ctx.Err())
			break
		}

		payload := buildBarkPayload(deviceKey, content, options)
		resp, err := xrequest.New().
			WithContext(ctx).
			SetBody(payload).
			Post(strings.TrimRight(baseURL, "/") + "/push")
		if err != nil {
			sendErrs = append(sendErrs, fmt.Errorf("bark key %s: %w", deviceKey, err))
			continue
		}

		if resp.StatusCode() >= 400 {
			sendErrs = append(sendErrs, fmt.Errorf("bark key %s: %w", deviceKey, barkResponseError(resp)))
			continue
		}

		code := resp.Json().Get("code").Int64()
		if code != 0 && code != 200 {
			sendErrs = append(sendErrs, fmt.Errorf("bark key %s: %w", deviceKey, barkResponseError(resp)))
		}
	}

	return errors.Join(sendErrs...)
}

func parseBarkTarget(botID string) (baseURL string, deviceKeys []string, err error) {
	raw := strings.TrimSpace(botID)
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" {
		return "", nil, errors.New("missing bark device key")
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return parseBarkAbsoluteTarget(raw)
	}

	if !strings.Contains(raw, "/") {
		deviceKeys, err = splitBarkDeviceKeys(raw)
		if err != nil {
			return "", nil, err
		}
		return barkDefaultBaseURL, deviceKeys, nil
	}

	lastSlash := strings.LastIndex(raw, "/")
	baseURL = strings.TrimSpace(raw[:lastSlash])
	deviceKeys, err = splitBarkDeviceKeys(raw[lastSlash+1:])
	if baseURL == "" || err != nil {
		if err != nil {
			return "", nil, err
		}
		return "", nil, errors.New("invalid bark target")
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	return strings.TrimRight(baseURL, "/"), deviceKeys, nil
}

func parseBarkAbsoluteTarget(raw string) (baseURL string, deviceKeys []string, err error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", nil, err
	}

	cleanPath := strings.Trim(parsed.Path, "/")
	if cleanPath == "" {
		return "", nil, errors.New("missing bark device key")
	}

	lastSlash := strings.LastIndex(cleanPath, "/")
	deviceKeyPart := cleanPath
	parsed.Path = ""
	if lastSlash >= 0 {
		deviceKeyPart = cleanPath[lastSlash+1:]
		parsed.Path = "/" + strings.Trim(cleanPath[:lastSlash], "/")
	}

	deviceKeys, err = splitBarkDeviceKeys(deviceKeyPart)
	if err != nil {
		return "", nil, err
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), deviceKeys, nil
}

func splitBarkDeviceKeys(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("missing bark device key")
	}

	parts := strings.Split(raw, ",")
	deviceKeys := make([]string, 0, len(parts))
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key == "" {
			return nil, errors.New("invalid bark device key")
		}
		deviceKeys = append(deviceKeys, key)
	}
	return deviceKeys, nil
}

func buildBarkPayload(deviceKey, content string, options NotifyOptions) barkPayload {
	payload := barkPayload{
		DeviceKey: deviceKey,
		Body:      content,
	}

	if strings.TrimSpace(options.Title) != "" {
		payload.Title = options.Title
	}
	if strings.TrimSpace(options.Subtitle) != "" {
		payload.Subtitle = options.Subtitle
	}
	if strings.TrimSpace(options.Level) != "" {
		payload.Level = options.Level
	}
	if strings.TrimSpace(options.Volume) != "" {
		payload.Volume = options.Volume
	}
	if options.Badge != nil {
		payload.Badge = options.Badge
	}
	if options.Call {
		payload.Call = "1"
	}
	if options.AutoCopy {
		payload.AutoCopy = "1"
	}
	if strings.TrimSpace(options.Copy) != "" {
		payload.Copy = options.Copy
	}
	if strings.TrimSpace(options.Sound) != "" {
		payload.Sound = options.Sound
	}
	if strings.TrimSpace(options.Icon) != "" {
		payload.Icon = options.Icon
	}
	if strings.TrimSpace(options.Group) != "" {
		payload.Group = options.Group
	}
	if strings.TrimSpace(options.Ciphertext) != "" {
		payload.Ciphertext = options.Ciphertext
	}
	if options.IsArchive {
		payload.IsArchive = "1"
	}
	if strings.TrimSpace(options.URL) != "" {
		payload.URL = options.URL
	}
	if strings.TrimSpace(options.Action) != "" {
		payload.Action = options.Action
	}

	return payload
}

func barkResponseError(resp *xrequest.Response) error {
	message := strings.TrimSpace(resp.Json().Get("message").String())
	if message == "" {
		message = strings.TrimSpace(resp.Json().Get("msg").String())
	}
	if message == "" {
		message = strings.TrimSpace(resp.String())
	}
	if message == "" {
		message = fmt.Sprintf("bark send failed, status: %d", resp.StatusCode())
	}
	return errors.New(message)
}
