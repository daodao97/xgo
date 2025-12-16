package xnotify

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/daodao97/xgo/xrequest"
)

const (
	larkWebhookURL  = "https://open.larksuite.com/open-apis/bot/v2/hook/"
	respSuccessCode = int64(0)
)

type larkSender struct{}

// 飞书文本消息结构
type textMsg struct {
	MsgType string         `json:"msg_type"`
	Content textMsgContent `json:"content"`
}

type textMsgContent struct {
	Text string `json:"text"`
}

func (larkSender) Send(ctx context.Context, botID, message string, mentions []string) error {
	return SendLarkText(ctx, botID, message, mentions)
}

func init() {
	if err := RegisterProvider(LarkScheme, larkSender{}); err != nil {
		panic(err)
	}
}

// SendLarkText sends a text message to Lark with optional user mentions.
func SendLarkText(ctx context.Context, botID, content string, mentions []string) error {
	if botID == "" {
		return errors.New("missing lark bot id")
	}

	withMentions := content
	if len(mentions) > 0 {
		var tokens []string
		for _, m := range mentions {
			if m == "" {
				continue
			}
			tokens = append(tokens, fmt.Sprintf("<at user_id=\"%s\">%s</at>", m, m))
		}
		if len(tokens) > 0 {
			withMentions = strings.TrimSpace(content + "\n" + strings.Join(tokens, " "))
		}
	}

	msgInfo := textMsg{MsgType: "text", Content: textMsgContent{Text: withMentions}}

	resp, err := xrequest.New().WithContext(ctx).SetBody(msgInfo).Post(larkWebhookURL + botID)
	if err != nil {
		return err
	}
	if resp.Json().Get("code").Int64() != respSuccessCode {
		return errors.New(resp.Json().Get("msg").String())
	}
	return nil
}
