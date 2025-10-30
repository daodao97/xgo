# XRequest - 简单易用的 HTTP 客户端

XRequest 是一个功能丰富的 Go HTTP 客户端库，提供链式调用、重试机制、文件上传、SSE 支持等功能。

## 特性

- 🔗 **链式调用** - 流畅的 API 设计
- 🔄 **智能重试** - 可配置的重试策略
- 📁 **文件上传** - 支持多文件上传
- 🌊 **流式响应** - SSE(Server-Sent Events)支持
- 🔐 **认证支持** - Basic Auth 和自定义认证
- 🚀 **高性能** - 连接池和并发安全
- 🛡️ **错误处理** - 详细的错误信息
- 📊 **调试模式** - cURL 命令生成和响应调试

## 安装

```bash
go get github.com/daodao97/xgo/xrequest
```

## 快速开始

### 基本 GET 请求

```go
package main

import (
    "fmt"
    "github.com/daodao97/xgo/xrequest"
)

func main() {
    resp, err := xrequest.New().Get("https://httpbin.org/get")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("状态码: %d\n", resp.StatusCode())
    fmt.Printf("响应: %s\n", resp.String())
}
```

### POST 请求发送 JSON

```go
data := map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   25,
}

resp, err := xrequest.New().
    SetBody(data).
    Post("https://httpbin.org/post")

if err != nil {
    panic(err)
}

// 解析 JSON 响应
result := resp.Json()
fmt.Printf("服务器返回: %s\n", result.Get("json").String())
```

## 详细使用示例

### 1. 查询参数

```go
resp, err := xrequest.New().
    SetQueryParams(map[string]string{
        "page":     "1",
        "size":     "10",
        "keyword":  "golang",
    }).
    Get("https://api.example.com/search")
```

### 2. 自定义请求头

```go
resp, err := xrequest.New().
    SetHeaders(map[string]string{
        "User-Agent":    "MyApp/1.0",
        "Authorization": "Bearer your-token",
        "X-Request-ID":  "req-12345",
    }).
    SetHeader("Accept", "application/json").
    Get("https://api.example.com/data")
```

### 3. Cookie 管理

```go
resp, err := xrequest.New().
    SetCookies(map[string]string{
        "session_id": "abc123",
        "user_pref":  "dark_mode",
    }).
    SetCookie("language", "zh-CN").
    Get("https://example.com")
```

### 4. 表单数据提交

```go
// application/x-www-form-urlencoded
resp, err := xrequest.New().
    SetFormData(map[string]string{
        "username": "user123",
        "password": "pass456",
    }).
    Post("https://example.com/login")

// 或使用 SetFormUrlEncode (效果相同)
resp, err := xrequest.New().
    SetFormUrlEncode(map[string]string{
        "key": "value",
    }).
    Post("https://example.com/form")
```

### 5. 文件上传

```go
file, err := os.Open("document.pdf")
if err != nil {
    panic(err)
}
defer file.Close()

resp, err := xrequest.New().
    AddFile("document", "document.pdf", file).
    SetFormData(map[string]string{
        "description": "重要文档",
        "category":    "contract",
    }).
    Post("https://example.com/upload")
```

### 6. 基本认证

```go
resp, err := xrequest.New().
    SetBasicAuth("username", "password").
    Get("https://api.example.com/protected")
```

### 7. 重试机制

```go
// 基本重试：最多重试3次，每次间隔2秒
resp, err := xrequest.New().
    SetRetry(3, time.Second*2).
    Get("https://unreliable-api.com/data")

// 自定义重试条件
retryCondition := func(resp *xrequest.Response, err error) bool {
    if err != nil {
        return true // 网络错误时重试
    }
    // 5xx 错误或 429(Too Many Requests) 时重试
    return resp.StatusCode() >= 500 || resp.StatusCode() == 429
}

resp, err := xrequest.New().
    SetRetryWithCondition(3, time.Second, retryCondition).
    Get("https://api.example.com/data")
```

### 8. 超时设置

```go
resp, err := xrequest.New().
    SetTimeout(10 * time.Second).
    Get("https://slow-api.com/data")
```

### 9. Context 支持

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := xrequest.New().
    WithContext(ctx).
    Get("https://api.example.com/data")
```

### 10. 请求钩子

```go
resp, err := xrequest.New().
    AddReqHook(func(req *http.Request) error {
        // 添加自定义头部
        req.Header.Set("X-Custom-Header", "custom-value")
        return nil
    }).
    AddReqHook(func(req *http.Request) error {
        // 记录请求日志
        fmt.Printf("发送请求: %s %s\n", req.Method, req.URL)
        return nil
    }).
    Get("https://api.example.com/data")
```

### 11. 代理设置

```go
resp, err := xrequest.New().
    SetProxy("http://proxy.example.com:8080").
    Get("https://api.example.com/data")
```

### 12. 自定义 HTTP 客户端

```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     30 * time.Second,
    },
}

resp, err := xrequest.New().
    SetClient(client).
    Get("https://api.example.com/data")
```

## 响应处理

### 基本响应处理

```go
resp, err := xrequest.New().Get("https://api.example.com/user/123")
if err != nil {
    panic(err)
}

// 检查状态码
if resp.IsError() {
    fmt.Printf("请求失败: %d\n", resp.StatusCode())
    return
}

// 获取响应内容
fmt.Println("原始响应:", resp.String())
fmt.Println("响应字节:", len(resp.Bytes()))
```

### JSON 响应解析

```go
resp, err := xrequest.New().Get("https://api.example.com/user/123")

// 方法1: 使用内置 JSON 解析器
jsonData := resp.Json()
name := jsonData.Get("name").String()
age := jsonData.Get("age").Int()

// 方法2: 解析到结构体
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var user User
err = resp.Scan(&user)
if err != nil {
    panic(err)
}
```

### XML 响应解析

```go
type Config struct {
    Version string `xml:"version"`
    Name    string `xml:"name"`
}

var config Config
err = resp.XML(&config)
if err != nil {
    panic(err)
}
```

### 获取响应头

```go
headers := resp.Headers()
contentType := headers.Get("Content-Type")
server := headers.Get("Server")
```

## Server-Sent Events (SSE) 支持

```go
resp, err := xrequest.New().
    SetHeader("Accept", "text/event-stream").
    Get("https://api.example.com/events")

if err != nil {
    panic(err)
}

// 获取事件流
eventChan, err := resp.SSE()
if err != nil {
    panic(err)
}

// 处理事件
for event := range eventChan {
    fmt.Printf("收到事件: %s\n", event)
}
```

## 调试模式

### 全局调试

```go
// 开启全局调试模式
xrequest.SetRequestDebug(true)

// 所有请求都会打印 cURL 命令和响应信息
resp, err := xrequest.New().Get("https://api.example.com/data")
```

### 单个请求调试

```go
resp, err := xrequest.New().
    SetDebug(true).
    Post("https://api.example.com/data")

// 会打印类似以下的调试信息:
// -------request curl command start-------
// curl -X POST 'https://api.example.com/data' -H 'Content-Type: application/json' -d '{"key":"value"}'
// response status: 200
// response body: {"result":"success"}
// -------request curl command end-------
```

## 错误处理

```go
resp, err := xrequest.New().Get("https://api.example.com/data")

if err != nil {
    // 检查是否为 xrequest 特定错误
    if reqErr, ok := err.(*xrequest.RequestError); ok {
        fmt.Printf("请求错误: %s\n", reqErr.Message)
        fmt.Printf("原始错误: %v\n", reqErr.Err)
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
    return
}

// 检查 HTTP 错误状态
if resp.IsError() {
    fmt.Printf("HTTP 错误: %d\n", resp.StatusCode())
    fmt.Printf("错误内容: %s\n", resp.String())
}
```

## 高级用法

### 复杂工作流

```go
// 1. 登录获取 token
loginResp, err := xrequest.New().
    SetBody(map[string]string{
        "username": "user",
        "password": "pass",
    }).
    Post("https://api.example.com/login")

if err != nil || loginResp.IsError() {
    panic("登录失败")
}

token := loginResp.Json().Get("token").String()

// 2. 使用 token 获取数据
dataResp, err := xrequest.New().
    SetHeader("Authorization", "Bearer "+token).
    Get("https://api.example.com/user/profile")

// 3. 上传文件
file, _ := os.Open("avatar.jpg")
defer file.Close()

uploadResp, err := xrequest.New().
    SetHeader("Authorization", "Bearer "+token).
    AddFile("avatar", "avatar.jpg", file).
    Post("https://api.example.com/user/avatar")
```

### 并发请求

```go
urls := []string{
    "https://api.example.com/endpoint1",
    "https://api.example.com/endpoint2", 
    "https://api.example.com/endpoint3",
}

var wg sync.WaitGroup
results := make(chan string, len(urls))

for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        
        resp, err := xrequest.New().
            SetTimeout(5 * time.Second).
            Get(u)
            
        if err != nil {
            results <- fmt.Sprintf("错误 %s: %v", u, err)
        } else {
            results <- fmt.Sprintf("成功 %s: %d", u, resp.StatusCode())
        }
    }(url)
}

wg.Wait()
close(results)

for result := range results {
    fmt.Println(result)
}
```

### 中间件模式

```go
// 创建带有公共配置的请求构建器
func newAPIRequest() *xrequest.Request {
    return xrequest.New().
        SetHeader("User-Agent", "MyApp/1.0").
        SetHeader("Accept", "application/json").
        SetTimeout(10 * time.Second).
        SetRetry(2, time.Second).
        AddReqHook(func(req *http.Request) error {
            // 添加请求追踪ID
            req.Header.Set("X-Request-ID", generateRequestID())
            return nil
        })
}

// 使用公共配置
resp1, err := newAPIRequest().Get("https://api.example.com/users")
resp2, err := newAPIRequest().
    SetBody(userData).
    Post("https://api.example.com/users")
```

## API 参考

### Request 方法

| 方法 | 描述 |
|------|------|
| `New()` | 创建新的请求实例 |
| `SetMethod(method)` | 设置 HTTP 方法 |
| `SetURL(url)` | 设置请求 URL |
| `SetBody(body)` | 设置请求体 |
| `SetHeaders(headers)` | 设置多个请求头 |
| `SetHeader(key, value)` | 设置单个请求头 |
| `SetCookies(cookies)` | 设置多个 Cookie |
| `SetCookie(key, value)` | 设置单个 Cookie |
| `SetQueryParams(params)` | 设置查询参数 |
| `SetQueryParam(key, value)` | 设置单个查询参数 |
| `SetFormData(data)` | 设置表单数据 |
| `SetFormUrlEncode(data)` | 设置 URL 编码表单数据 |
| `SetBasicAuth(user, pass)` | 设置基本认证 |
| `SetTimeout(duration)` | 设置超时时间 |
| `SetRetry(attempts, delay)` | 设置重试策略 |
| `SetRetryWithCondition(attempts, delay, condition)` | 设置自定义重试条件 |
| `SetProxy(proxy)` | 设置代理 |
| `SetClient(client)` | 设置自定义 HTTP 客户端 |
| `SetDebug(debug)` | 设置调试模式 |
| `AddFile(fieldName, fileName, content)` | 添加上传文件 |
| `AddReqHook(hook)` | 添加请求钩子 |
| `WithContext(ctx)` | 设置上下文 |
| `WithRequest(req)` | 使用现有 HTTP 请求 |

### HTTP 方法快捷方式

| 方法 | 描述 |
|------|------|
| `Get(url)` | 发送 GET 请求 |
| `Post(url)` | 发送 POST 请求 |
| `Put(url)` | 发送 PUT 请求 |
| `Delete(url)` | 发送 DELETE 请求 |
| `Patch(url)` | 发送 PATCH 请求 |
| `Do()` | 执行请求 |

### Response 方法

| 方法 | 描述 |
|------|------|
| `StatusCode()` | 获取状态码 |
| `String()` | 获取响应字符串 |
| `Bytes()` | 获取响应字节数组 |
| `Json()` | 获取 JSON 解析器 |
| `Scan(dest)` | 解析到结构体 |
| `XML(dest)` | 解析 XML |
| `Headers()` | 获取响应头 |
| `IsError()` | 检查是否为错误状态 |
| `Error()` | 获取错误信息 |
| `SSE()` | 获取 SSE 事件流 |
| `Stream()` | 获取流式响应(SSE 别名) |

## 最佳实践

### 1. 错误处理

```go
resp, err := xrequest.New().Get("https://api.example.com/data")

// 始终检查网络错误
if err != nil {
    log.Printf("网络错误: %v", err)
    return
}

// 检查 HTTP 状态错误
if resp.IsError() {
    log.Printf("HTTP 错误: %d, 响应: %s", resp.StatusCode(), resp.String())
    return
}
```

### 2. 资源管理

```go
// 文件上传时记得关闭文件
file, err := os.Open("large-file.zip")
if err != nil {
    return err
}
defer file.Close() // 重要!

resp, err := xrequest.New().
    AddFile("upload", "large-file.zip", file).
    Post("https://upload.example.com")
```

### 3. 超时设置

```go
// 根据 API 特性设置合理的超时时间
resp, err := xrequest.New().
    SetTimeout(30 * time.Second). // 文件上传需要更长时间
    Post("https://upload.example.com")
```

### 4. 重试策略

```go
// 只对幂等操作使用重试
resp, err := xrequest.New().
    SetRetry(3, time.Second*2).
    Get("https://api.example.com/data") // GET 是幂等的

// POST/PUT 需要谨慎使用重试
```

### 5. 调试模式

```go
// 在开发环境启用调试
if os.Getenv("ENV") == "development" {
    xrequest.SetRequestDebug(true)
}
```

### 6. 透传上游响应及错误排查

```go
import (
    "github.com/daodao97/xgo/xlog"
    "github.com/daodao97/xgo/xrequest"
)

func relayHandler(ctx *gin.Context, apiURL string, hooks ...xrequest.ResponseHook) {
    resp, err := xrequest.New().
        SetHeaders(map[string]string{
            "Authorization": ctx.GetHeader("Authorization"),
        }).
        SetBody(ctx.Request.Body).
        Post(apiURL)
    if err != nil {
        xlog.ErrorCtx(ctx, "上游请求失败", xlog.Any("error", err))
        ctx.JSON(http.StatusBadGateway, gin.H{"msg": "upstream error"})
        return
    }

    // 透传上游响应
    totalBytes, writeErr := resp.ToHttpResponseWriter(ctx.Writer, hooks...)
    if writeErr != nil {
        if xrequest.IsClientDisconnected(writeErr) {
            xlog.WarnCtx(ctx, "客户端已断开", xlog.Any("error", writeErr))
            return // 下游已断开，不必继续写
        }

        xlog.ErrorCtx(ctx, "透传响应失败", xlog.Any("error", writeErr))
        ctx.Status(http.StatusBadGateway)
        return
    }

    // totalBytes 为成功写入下游的字节数
    if totalBytes == 0 {
        // 结合上游声明的 Content-Length 与实际内容判断
        contentLen := resp.RawResponse.ContentLength
        xlog.WarnCtx(ctx, "上游无内容或被 hook 丢弃",
            xlog.Int("status", resp.StatusCode()),
            xlog.Int64("upstream_content_length", contentLen),
            xlog.Bool("body_is_empty", resp.BodyIsEmpty()),
        )
    }
}
```

如果业务需要在写入前检查上游内容，可通过 `resp.BodyIsEmpty()`、`resp.Bytes()` 等方法获取原始数据，再决定是否透传或根据需求定制处理。

## 与其他 HTTP 客户端对比

| 特性 | xrequest | net/http | resty | req |
|------|----------|----------|-------|-----|
| 链式调用 | ✅ | ❌ | ✅ | ✅ |
| 重试机制 | ✅ | ❌ | ✅ | ✅ |
| 文件上传 | ✅ | 手动 | ✅ | ✅ |
| SSE 支持 | ✅ | 手动 | ❌ | ❌ |
| 调试模式 | ✅ (cURL) | ❌ | ✅ | ✅ |
| 中间件 | ✅ (Hooks) | 手动 | ✅ | ✅ |
| JSON 解析 | ✅ | 手动 | ✅ | ✅ |
| Context 支持 | ✅ | ✅ | ✅ | ✅ |

## 许可证

MIT License - 详见 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.0.0
- 基本 HTTP 客户端功能
- 链式调用支持
- 重试机制
- 文件上传
- SSE 支持
- 调试模式

---

如有问题或建议，请访问 [GitHub 仓库](https://github.com/daodao97/xgo) 提交 Issue。
