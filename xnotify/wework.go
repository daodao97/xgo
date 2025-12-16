package xnotify

import (
	"context"
	"errors"
	"fmt"

	"github.com/daodao97/xgo/xrequest"
)

const wecomWebhookURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="

type weworkSender struct{}

func SendWeComText(ctx context.Context, botID, content string, mentionedMobiles []string) error {
	if botID == "" {
		return errors.New("missing wecom bot id")
	}

	text := map[string]any{
		"content": content,
	}

	if len(mentionedMobiles) > 0 {
		text["mentioned_mobile_list"] = mentionedMobiles
	}

	data := map[string]any{
		"msgtype": "text",
		"text":    text,
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
	return SendWeComText(ctx, botID, message, mentions)
}

func init() {
	if err := RegisterProvider(WeWorkScheme, weworkSender{}); err != nil {
		panic(err)
	}
}
