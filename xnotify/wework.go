package xnotify

import (
	"context"
	"errors"
	"fmt"

	"github.com/daodao97/xgo/xrequest"
)

const wecomWebhookURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="
const (
	wecomMsgTypeText       = MessageTypeText
	wecomMsgTypeMarkdown   = MessageTypeMarkdown
	wecomMsgTypeMarkdownV2 = MessageTypeMarkdownV2
)

type weworkSender struct{}

func SendWeComText(ctx context.Context, botID, content string, mentionedMobiles []string) error {
	return sendWeCom(ctx, botID, buildWeComTextPayload(content, mentionedMobiles))
}

// SendWeComMarkdown sends a markdown message to WeCom webhook.
func SendWeComMarkdown(ctx context.Context, botID, content string) error {
	return sendWeCom(ctx, botID, buildWeComMarkdownPayload(content))
}

// SendWeComMarkdownV2 sends a markdown_v2 message to WeCom webhook.
func SendWeComMarkdownV2(ctx context.Context, botID, content string) error {
	return sendWeCom(ctx, botID, buildWeComMarkdownV2Payload(content))
}

func buildWeComTextPayload(content string, mentionedMobiles []string) map[string]any {
	text := map[string]any{
		"content": content,
	}
	if len(mentionedMobiles) > 0 {
		text["mentioned_mobile_list"] = mentionedMobiles
	}
	return map[string]any{
		"msgtype": wecomMsgTypeText,
		"text":    text,
	}
}

func buildWeComMarkdownPayload(content string) map[string]any {
	return map[string]any{
		"msgtype":  wecomMsgTypeMarkdown,
		"markdown": map[string]any{"content": content},
	}
}

func buildWeComMarkdownV2Payload(content string) map[string]any {
	return map[string]any{
		"msgtype":     wecomMsgTypeMarkdownV2,
		"markdown_v2": map[string]any{"content": content},
	}
}

func sendWeCom(ctx context.Context, botID string, data map[string]any) error {
	if botID == "" {
		return errors.New("missing wecom bot id")
	}

	resp, err := xrequest.New().WithContext(ctx).SetBody(data).Post(wecomWebhookURL + botID)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("wecom send failed, status: %d", resp.StatusCode())
	}
	return nil
}

func (weworkSender) Send(ctx context.Context, botID, message string, mentions []string) error {
	return weworkSender{}.SendWithOptions(ctx, botID, message, mentions, NotifyOptions{})
}

func (weworkSender) SendWithOptions(ctx context.Context, botID, message string, mentions []string, options NotifyOptions) error {
	switch options.MessageType {
	case "", MessageTypeText:
		return SendWeComText(ctx, botID, message, mentions)
	case MessageTypeMarkdown:
		return SendWeComMarkdown(ctx, botID, message)
	case MessageTypeMarkdownV2:
		return SendWeComMarkdownV2(ctx, botID, message)
	default:
		return fmt.Errorf("unsupported wecom message type: %s", options.MessageType)
	}
}

func init() {
	if err := RegisterProvider(WeWorkScheme, weworkSender{}); err != nil {
		panic(err)
	}
}
