# xnotify

`xnotify` 是一个基于 provider 的通知组件，当前内置支持：

- `lark` / 飞书机器人
- `wework` / 企业微信机器人
- `bark` / Bark 推送

统一入口：

- `Notify(ctx, botID, message)`
- `NotifyWithOptions(ctx, botID, message, opts...)`

---

## 1. 快速开始

```go
package main

import (
	"context"

	"github.com/daodao97/xgo/xnotify"
)

func main() {
	ctx := context.Background()

	_ = xnotify.Notify(ctx, "lark://bot_id", "hello")
	_ = xnotify.Notify(ctx, "wework://bot_key", "hello")
	_ = xnotify.Notify(ctx, "bark://device_key", "hello")
}
```

---

## 2. botID 格式

### Lark

```text
lark://{bot_id}
lark://{bot_id}@{mention1,mention2}
```

示例：

```go
err := xnotify.Notify(ctx, "lark://bot_abc123", "hello")
err := xnotify.Notify(ctx, "lark://bot_abc123@user1,user2", "hello")
```

---

### WeWork

```text
wework://{bot_key}
wework://{bot_key}@13800001111,@all
wework://{bot_key}?mention=13800001111,@all
```

示例：

```go
err := xnotify.Notify(ctx, "wework://key123", "hello")
err := xnotify.Notify(ctx, "wework://key123@13800001111,@all", "hello")
```

---

### Bark

```text
bark://{device_key}
bark://{device_key1,device_key2}
bark://{host}/{device_key}
bark://{host}/{device_key1,device_key2}
bark://http://127.0.0.1:8080/{device_key}
bark://https://bark.example.com/{device_key1,device_key2}
```

说明：

- `bark://device_key` 默认发送到官方地址 `https://api.day.app`
- 多个 key 使用英文逗号 `,` 分隔
- 多 key 场景下，内部会拆成 **多个独立请求**
- 自建 Bark 服务可传完整 `http(s)://` 地址

示例：

```go
err := xnotify.Notify(ctx, "bark://device_key", "hello")
err := xnotify.Notify(ctx, "bark://key1,key2,key3", "hello")
err := xnotify.Notify(ctx, "bark://bark.example.com/key1,key2", "hello")
err := xnotify.Notify(ctx, "bark://http://127.0.0.1:8080/device_key", "hello")
```

---

## 3. NotifyWithOptions

### 通用写法

```go
err := xnotify.NotifyWithOptions(
	ctx,
	"wework://key123",
	"hello",
	xnotify.WithMessageType(xnotify.MessageTypeText),
)
```

---

## 4. WeWork 用例

### 文本消息

```go
err := xnotify.NotifyWithOptions(
	ctx,
	"wework://key123@13800001111,@all",
	"hello wework",
)
```

### Markdown 消息

```go
err := xnotify.NotifyWithOptions(
	ctx,
	"wework://key123",
	"**hello wework**",
	xnotify.WithMessageType(xnotify.MessageTypeMarkdown),
)
```

### Markdown V2 消息

```go
err := xnotify.NotifyWithOptions(
	ctx,
	"wework://key123",
	"# hello wework",
	xnotify.WithMessageType(xnotify.MessageTypeMarkdownV2),
)
```

---

## 5. Bark 用例

`Notify(message)` 中的 `message` 会映射到 Bark 的 `body` 字段。

### 基础推送

```go
err := xnotify.Notify(ctx, "bark://device_key", "这是一条测试消息")
```

### 多 key 推送

```go
err := xnotify.Notify(ctx, "bark://key1,key2,key3", "这是一条多 key 测试消息")
```

### 自建 Bark

```go
err := xnotify.Notify(ctx, "bark://bark.example.com/device_key", "hello")
err := xnotify.Notify(ctx, "bark://http://127.0.0.1:8080/device_key", "hello")
```

### 带标题、副标题、分组

```go
err := xnotify.NotifyWithOptions(
	ctx,
	"bark://device_key",
	"body text",
	xnotify.WithTitle("标题"),
	xnotify.WithSubtitle("副标题"),
	xnotify.WithGroup("system-alert"),
)
```

### 常用 Bark 参数

```go
err := xnotify.NotifyWithOptions(
	ctx,
	"bark://key1,key2",
	"服务异常，请尽快处理",
	xnotify.WithTitle("告警通知"),
	xnotify.WithSubtitle("生产环境"),
	xnotify.WithURL("https://example.com"),
	xnotify.WithGroup("ops"),
	xnotify.WithSound("alarm"),
	xnotify.WithIcon("https://example.com/icon.png"),
	xnotify.WithLevel("timeSensitive"),
	xnotify.WithVolume("8"),
	xnotify.WithCopy("服务异常，请尽快处理"),
	xnotify.WithAction("none"),
	xnotify.WithBadge(1),
	xnotify.WithAutoCopy(true),
	xnotify.WithCall(true),
	xnotify.WithArchive(true),
)
```

说明：

- Bark 当前仅支持 `text` 发送入口，不支持 `markdown`
- `mentions` 对 Bark 无效
- 多 key 发送时，若部分 key 失败，函数会返回聚合错误

---

## 6. Lark 用例

### 文本消息

```go
err := xnotify.Notify(ctx, "lark://bot_abc123", "hello lark")
```

### @用户

```go
err := xnotify.Notify(ctx, "lark://bot_abc123@user1,user2", "hello lark")
```

说明：

- Lark 当前仅支持文本消息
- `mentions` 会被转换为飞书 `<at ...>` 格式

---

## 7. 直接调用底层方法

```go
err := xnotify.SendLarkText(ctx, "bot_id", "hello", []string{"user1"})
err := xnotify.SendWeComText(ctx, "key123", "hello", []string{"13800001111", "@all"})
err := xnotify.SendWeComMarkdown(ctx, "key123", "**hello**")
err := xnotify.SendWeComMarkdownV2(ctx, "key123", "# hello")
err := xnotify.SendBarkText(ctx, "device_key", "hello")
err := xnotify.SendBarkWithOptions(ctx, "key1,key2", "hello", xnotify.NotifyOptions{
	Title: "标题",
	Group: "test",
})
```

---

## 8. 自定义 Provider

```go
type mySender struct{}

func (mySender) Send(ctx context.Context, botID string, message string, mentions []string) error {
	return nil
}

func init() {
	_ = xnotify.RegisterProvider("my", mySender{})
}

// 使用
err := xnotify.Notify(ctx, "my://bot_id", "hello")
```

