package xnotify

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

const (
	LarkScheme   = "lark"
	WeWorkScheme = "wework"
)

const (
	MessageTypeText       = "text"
	MessageTypeMarkdown   = "markdown"
	MessageTypeMarkdownV2 = "markdown_v2"
)

type Sender interface {
	Send(ctx context.Context, botID string, message string, mentions []string) error
}

// OptionSender is an optional extension for providers that support rich options.
type OptionSender interface {
	SendWithOptions(ctx context.Context, botID string, message string, mentions []string, options NotifyOptions) error
}

type NotifyOptions struct {
	// MessageType defaults to "text".
	MessageType string
	// Mentions will be merged with mentions parsed from botID.
	Mentions []string
}

type NotifyOption func(*NotifyOptions)

func WithMessageType(messageType string) NotifyOption {
	return func(o *NotifyOptions) {
		o.MessageType = strings.TrimSpace(messageType)
	}
}

func WithMentions(mentions ...string) NotifyOption {
	return func(o *NotifyOptions) {
		o.Mentions = append(o.Mentions, mentions...)
	}
}

var (
	providers   = make(map[string]Sender)
	providersMu sync.RWMutex
)

// RegisterProvider registers a sender for a scheme (e.g. "lark").
func RegisterProvider(scheme string, sender Sender) error {
	if scheme == "" {
		return errors.New("scheme is required")
	}
	if sender == nil {
		return errors.New("sender is required")
	}

	providersMu.Lock()
	defer providersMu.Unlock()
	providers[scheme] = sender
	return nil
}

func getProvider(scheme string) (Sender, bool) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	s, ok := providers[scheme]
	return s, ok
}

// Notify routes the message to the provider specified in botID.
// Supported formats:
// - lark://{bot_id}
// - lark://{bot_id}@{mention1,mention2}
// - wework://{bot_id}
// Query string `?mention=` is also supported and merged with the @ suffix.
func Notify(ctx context.Context, botID, message string) error {
	return NotifyWithOptions(ctx, botID, message)
}

// NotifyWithOptions routes the message with options (such as message type).
func NotifyWithOptions(ctx context.Context, botID, message string, opts ...NotifyOption) error {
	target, err := parseBotID(botID)
	if err != nil {
		return err
	}

	sender, ok := getProvider(target.scheme)
	if !ok {
		return fmt.Errorf("unsupported provider: %s", target.scheme)
	}

	options := buildNotifyOptions(opts...)
	mergedMentions := mergeMentions(target.mentions, options.Mentions)

	if optionSender, ok := sender.(OptionSender); ok {
		options.Mentions = mergedMentions
		return optionSender.SendWithOptions(ctx, target.botID, message, mergedMentions, options)
	}

	if options.MessageType != MessageTypeText {
		return fmt.Errorf("provider %s does not support message type: %s", target.scheme, options.MessageType)
	}
	return sender.Send(ctx, target.botID, message, mergedMentions)
}

type botTarget struct {
	scheme   string
	botID    string
	mentions []string
}

func parseBotID(raw string) (botTarget, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return botTarget{}, errors.New("bot id is required")
	}

	parts := strings.SplitN(raw, "://", 2)
	if len(parts) != 2 || parts[0] == "" {
		return botTarget{}, fmt.Errorf("invalid bot id scheme in %q", raw)
	}

	scheme := parts[0]
	body := parts[1]
	if body == "" {
		return botTarget{}, errors.New("missing bot id")
	}

	target, err := buildTarget(scheme, body)
	if err != nil {
		return botTarget{}, err
	}
	return target, nil
}

func buildTarget(scheme, body string) (botTarget, error) {
	var mentionPart string
	path := body
	if idx := strings.Index(body, "?"); idx >= 0 {
		path = body[:idx]
		query := body[idx+1:]
		values, err := url.ParseQuery(query)
		if err == nil {
			mentionPart = strings.Join(values["mention"], ",")
		}
	}

	segments := strings.SplitN(path, "@", 2)
	botID := strings.TrimSpace(segments[0])
	if botID == "" {
		return botTarget{}, errors.New("missing bot id")
	}

	mentions := collectMentions(segments, mentionPart)
	return botTarget{
		scheme:   scheme,
		botID:    botID,
		mentions: mentions,
	}, nil
}

func collectMentions(segments []string, queryMentions string) []string {
	var pathMentions []string
	if len(segments) == 2 {
		pathMentions = strings.Split(segments[1], ",")
	}
	var queryMentionList []string
	if strings.TrimSpace(queryMentions) != "" {
		queryMentionList = strings.Split(queryMentions, ",")
	}
	return mergeMentions(pathMentions, queryMentionList)
}

func buildNotifyOptions(opts ...NotifyOption) NotifyOptions {
	options := NotifyOptions{
		MessageType: MessageTypeText,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&options)
	}
	if strings.TrimSpace(options.MessageType) == "" {
		options.MessageType = MessageTypeText
	}
	options.Mentions = mergeMentions(options.Mentions)
	return options
}

func mergeMentions(mentionGroups ...[]string) []string {
	seen := make(map[string]struct{})
	var normalized []string
	for _, mentions := range mentionGroups {
		for _, m := range mentions {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			if _, ok := seen[m]; ok {
				continue
			}
			seen[m] = struct{}{}
			normalized = append(normalized, m)
		}
	}
	return normalized
}
