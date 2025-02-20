package xapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/daodao97/xgo/utils"
	"github.com/daodao97/xgo/xlog"
	"github.com/gin-gonic/gin"
)

func WithBearerAuth() OpenAPIOption {
	return func(doc *OpenAPIDocument) {
		if doc.Components == nil {
			doc.Components = &Components{
				SecuritySchemes: make(map[string]SecurityScheme),
			}
		}
		doc.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		}
		doc.Security = append(doc.Security, map[string][]string{
			"bearerAuth": {},
		})
	}
}

func WithAPIKeyAuth(name, in string) OpenAPIOption {
	return func(doc *OpenAPIDocument) {
		if doc.Components == nil {
			doc.Components = &Components{
				SecuritySchemes: make(map[string]SecurityScheme),
			}
		}
		doc.Components.SecuritySchemes["apiKeyAuth"] = SecurityScheme{
			Type: "apiKey",
			In:   in,
			Name: name,
		}
		doc.Security = append(doc.Security, map[string][]string{
			"apiKeyAuth": {},
		})
	}
}

func WithBasicAuth() OpenAPIOption {
	return func(doc *OpenAPIDocument) {
		if doc.Components == nil {
			doc.Components = &Components{
				SecuritySchemes: make(map[string]SecurityScheme),
			}
		}
		doc.Components.SecuritySchemes["basicAuth"] = SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}
		doc.Security = append(doc.Security, map[string][]string{
			"basicAuth": {},
		})
	}
}

// 定义OpenAPI文档结构
type OpenAPIDocument struct {
	OpenAPI    string                `json:"openapi"`
	Info       OpenAPIInfo           `json:"info"`
	Servers    []APIServer           `json:"servers"`
	Paths      map[string]PathItem   `json:"paths"`
	Components *Components           `json:"components,omitempty"`
	Security   []map[string][]string `json:"security,omitempty"`
}

type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	In           string `json:"in,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
}

type OpenAPIOption func(*OpenAPIDocument)

func WithServer(url, description string) OpenAPIOption {
	return func(doc *OpenAPIDocument) {
		doc.Servers = append(doc.Servers, APIServer{
			URL:         url,
			Description: description,
		})
	}
}

func WithInfo(title, version string) OpenAPIOption {
	return func(doc *OpenAPIDocument) {
		doc.Info = OpenAPIInfo{
			Title:   title,
			Version: version,
		}
	}
}

type APIServer struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
	// 其他HTTP方法...
}

type Operation struct {
	Tags        []string            `json:"tags,omitempty"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Deprecated  bool                `json:"deprecated,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

type RequestBody struct {
	Content map[string]MediaType `json:"content"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type                 string            `json:"type,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	AdditionalProperties *Schema           `json:"additionalProperties,omitempty"`
	Description          string            `json:"description,omitempty"`
	Required             []string          `json:"required,omitempty"`
	Format               string            `json:"format,omitempty"`
	Enum                 []any             `json:"enum,omitempty"`
	Minimum              *float64          `json:"minimum,omitempty"`
	Maximum              *float64          `json:"maximum,omitempty"`
	MinLength            *int              `json:"minLength,omitempty"`
	MaxLength            *int              `json:"maxLength,omitempty"`
	Pattern              string            `json:"pattern,omitempty"`
}

func generateSchema(t reflect.Type) Schema {
	schema := Schema{
		Properties: make(map[string]Schema),
	}

	switch t.Kind() {
	case reflect.Struct:
		// 检查是否为 decimal.Decimal 类型
		if t.String() == "decimal.Decimal" {
			schema.Type = "string"
			schema.Format = "decimal"
			return schema
		}

		if t.String() == "time.Time" {
			schema.Type = "string"
			schema.Format = "date-time"
			return schema
		}

		schema.Type = "object"
		var required []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Anonymous {
				// 处理匿名字段（嵌入式结构体）
				embeddedSchema := generateSchema(field.Type)
				for k, v := range embeddedSchema.Properties {
					schema.Properties[k] = v
				}
				required = append(required, embeddedSchema.Required...)
			} else {
				if uri := field.Tag.Get("uri"); uri != "" {
					continue
				}
				jsonTag := field.Tag.Get("json")
				if jsonTag == "-" {
					continue // 跳过被标记为忽略的字段
				}
				fieldName := strings.Split(jsonTag, ",")[0]
				if fieldName == "" {
					fieldName = field.Name
				}
				fieldSchema := generateSchema(field.Type)
				fieldSchema.Description = generateDescription(field)
				schema.Properties[fieldName] = fieldSchema

				if isRequired(field) {
					required = append(required, fieldName)
				}
			}
		}
		if len(required) > 0 {
			schema.Required = required
		}
	case reflect.Ptr:
		return generateSchema(t.Elem())
	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		schema.Items = &Schema{
			Properties: generateSchema(t.Elem()).Properties,
		}
	case reflect.Map:
		schema.Type = "object"
		schema.AdditionalProperties = &Schema{
			Properties: generateSchema(t.Elem()).Properties,
		}
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	default:
		schema.Type = "string" // 默认处理为字符串
	}

	return schema
}

func generateDescription(field reflect.StructField) string {
	var desc []string

	// 获取 comment 标签
	if comment := field.Tag.Get("comment"); comment != "" {
		desc = append(desc, comment)
	}

	// 解析 default 标签
	if defaultVal := field.Tag.Get("default"); defaultVal != "" {
		desc = append(desc, fmt.Sprintf("默认值: %s", defaultVal))
	}

	// 解析 binding 标签
	if binding := field.Tag.Get("binding"); binding != "" {
		bindingRules := strings.Split(binding, ",")
		for _, rule := range bindingRules {
			switch {
			case rule == "required":
				desc = append(desc, "此字段是必需的")
			case rule == "email":
				desc = append(desc, "必须是有效的电子邮件地址")
			case rule == "url":
				desc = append(desc, "必须是有效的URL")
			case strings.HasPrefix(rule, "min="):
				desc = append(desc, fmt.Sprintf("最小值为 %s", strings.TrimPrefix(rule, "min=")))
			case strings.HasPrefix(rule, "max="):
				desc = append(desc, fmt.Sprintf("最大值为 %s", strings.TrimPrefix(rule, "max=")))
			case strings.HasPrefix(rule, "len="):
				desc = append(desc, fmt.Sprintf("长度必须为 %s", strings.TrimPrefix(rule, "len=")))
			case strings.HasPrefix(rule, "oneof="):
				desc = append(desc, fmt.Sprintf("枚举值: %s", strings.TrimPrefix(rule, "oneof=")))
			case rule == "eq":
				desc = append(desc, fmt.Sprintf("必须等于 %s", strings.TrimPrefix(rule, "eq=")))
			case rule == "ne":
				desc = append(desc, fmt.Sprintf("不能等于 %s", strings.TrimPrefix(rule, "ne=")))
			case rule == "lt":
				desc = append(desc, fmt.Sprintf("必须小于 %s", strings.TrimPrefix(rule, "lt=")))
			case rule == "lte":
				desc = append(desc, fmt.Sprintf("必须小于或等于 %s", strings.TrimPrefix(rule, "lte=")))
			case rule == "gt":
				desc = append(desc, fmt.Sprintf("必须大于 %s", strings.TrimPrefix(rule, "gt=")))
			case rule == "gte":
				desc = append(desc, fmt.Sprintf("必须大于或等于 %s", strings.TrimPrefix(rule, "gte=")))
			case rule == "alpha":
				desc = append(desc, "只能包含字母")
			case rule == "alphanum":
				desc = append(desc, "只能包含字母和数字")
			case rule == "numeric":
				desc = append(desc, "必须是数字")
			case strings.HasPrefix(rule, "eqfield="):
				desc = append(desc, fmt.Sprintf("必须等于 %s 字段", strings.TrimPrefix(rule, "eqfield=")))
			case strings.HasPrefix(rule, "nefield="):
				desc = append(desc, fmt.Sprintf("不能等于 %s 字段", strings.TrimPrefix(rule, "nefield=")))
			case strings.HasPrefix(rule, "gtfield="):
				desc = append(desc, fmt.Sprintf("必须大于 %s 字段", strings.TrimPrefix(rule, "gtfield=")))
			case strings.HasPrefix(rule, "gtefield="):
				desc = append(desc, fmt.Sprintf("必须大于或等于 %s 字段", strings.TrimPrefix(rule, "gtefield=")))
			case strings.HasPrefix(rule, "ltfield="):
				desc = append(desc, fmt.Sprintf("必须小于 %s 字段", strings.TrimPrefix(rule, "ltfield=")))
			case strings.HasPrefix(rule, "ltefield="):
				desc = append(desc, fmt.Sprintf("必须小于或等于 %s 字段", strings.TrimPrefix(rule, "ltefield=")))
			case rule == "isdefault":
				desc = append(desc, "必须是默认值")
			case rule == "unique":
				desc = append(desc, "必须是唯一的")
			case rule == "alphaunicode":
				desc = append(desc, "只能包含 unicode 字符")
			case rule == "alphanumunicode":
				desc = append(desc, "只能包含 unicode 字母和数字")
			case rule == "lowercase":
				desc = append(desc, "只能包含小写字符")
			case rule == "uppercase":
				desc = append(desc, "只能包含大写字符")
			case rule == "json":
				desc = append(desc, "必须是有效的 JSON")
			case rule == "file":
				desc = append(desc, "必须是有效的文件路径")
			case rule == "uri":
				desc = append(desc, "必须是有效的 URI")
			case rule == "base64":
				desc = append(desc, "必须是有效的 base64 值")
			case strings.HasPrefix(rule, "contains="):
				desc = append(desc, fmt.Sprintf("必须包含 %s", strings.TrimPrefix(rule, "contains=")))
			case strings.HasPrefix(rule, "containsany="):
				desc = append(desc, fmt.Sprintf("必须包含 %s 中的任何字符", strings.TrimPrefix(rule, "containsany=")))
			case strings.HasPrefix(rule, "excludes="):
				desc = append(desc, fmt.Sprintf("不能包含 %s", strings.TrimPrefix(rule, "excludes=")))
			case strings.HasPrefix(rule, "excludesall="):
				desc = append(desc, fmt.Sprintf("不能包含 %s 中的任何字符", strings.TrimPrefix(rule, "excludesall=")))
			case strings.HasPrefix(rule, "startswith="):
				desc = append(desc, fmt.Sprintf("必须以 %s 开始", strings.TrimPrefix(rule, "startswith=")))
			case strings.HasPrefix(rule, "endswith="):
				desc = append(desc, fmt.Sprintf("必须以 %s 结束", strings.TrimPrefix(rule, "endswith=")))
			case rule == "ip":
				desc = append(desc, "必须是有效的 IP 地址")
			case rule == "ipv4":
				desc = append(desc, "必须是有效的 IPv4 地址")
			case rule == "datetime":
				desc = append(desc, "必须是有效的日期时间")
			case rule == "omitempty":
				desc = append(desc, "非必须")
			case rule == "oneof":
				desc = append(desc, fmt.Sprintf("枚举值: %s", strings.TrimPrefix(rule, "oneof=")))
			}
		}
	}

	return strings.Join(desc, ", ")
}

func isRequired(field reflect.StructField) bool {
	binding := field.Tag.Get("binding")
	return strings.Contains(binding, "required")
}

var apiRegistry []APIInfo

type APIInfo struct {
	Path         string
	Method       string
	RequestType  reflect.Type
	ResponseType reflect.Type
	Handler      gin.HandlerFunc
	Summary      string
	Description  string
	Tags         []string
	Deprecated   bool
	Accept       string
	Produce      string
}

func CollectRouteInfo(engine *gin.Engine) {
	routes := engine.Routes()
	for _, route := range routes {
		for i, info := range apiRegistry {
			if info.Path == "" && reflect.ValueOf(route.HandlerFunc).Pointer() == reflect.ValueOf(info.Handler).Pointer() {
				apiRegistry[i].Path = route.Path
				apiRegistry[i].Method = route.Method
				break
			}
		}
	}
}

func RegisterAPI[Req any, Resp any](handler func(*gin.Context, Req) (*Resp, error)) gin.HandlerFunc {
	wrappedHandler := HanderFunc(handler)

	if !Args.EnableOpenAPI {
		return wrappedHandler
	}

	if utils.IsGoRun() {
		reqType := reflect.TypeOf((*Req)(nil)).Elem()
		respType := reflect.TypeOf((*Resp)(nil)).Elem()

		// 获取handler函数的信息
		handlerValue := reflect.ValueOf(handler)
		handlerPtr := handlerValue.Pointer()
		handlerFunc := runtime.FuncForPC(handlerPtr)
		if handlerFunc == nil {
			xlog.Error("无法获取handler函数信息")
		} else {
			fileName, _ := handlerFunc.FileLine(handlerPtr)
			// fmt.Printf("Handler函数位置: %s:%d\n", fileName, lineNumber)

			fileContent, err := os.ReadFile(fileName)
			if err != nil {
				xlog.Error("read file error", xlog.Err(err))
			} else {
				lines := strings.Split(string(fileContent), "\n")
				funcName := handlerFunc.Name()
				shortFuncName := funcName[strings.LastIndex(funcName, ".")+1:]
				lineNumber := findFunctionDefinition(lines, shortFuncName)
				// fmt.Println("funcName", funcName)
				// fmt.Println("shortFuncName", shortFuncName)
				// fmt.Println("lineNumber", lineNumber)
				// fmt.Println("fileName", fileName)

				var comments []string
				for i := lineNumber - 2; i >= 0; i-- {
					line := strings.TrimSpace(lines[i])
					if strings.HasPrefix(line, "//") {
						comments = append([]string{strings.TrimPrefix(line, "//")}, comments...)
					} else {
						break
					}
				}

				commentResult := parseComments(strings.Join(comments, "\n"))
				// fmt.Printf("提取的注释: summary=%s, description=%s\n", commentResult.Summary, commentResult.Description)

				apiRegistry = append(apiRegistry, APIInfo{
					RequestType:  reqType,
					ResponseType: respType,
					Handler:      wrappedHandler,
					Summary:      commentResult.Summary,
					Description:  commentResult.Description,
					Tags:         commentResult.Tags,
					Deprecated:   commentResult.Deprecated,
					Accept:       commentResult.Accept,
					Produce:      commentResult.Produce,
				})
			}
		}
	}

	return wrappedHandler
}

func findFunctionDefinition(lines []string, funcName string) int {
	for i, line := range lines {
		// 查找函数定义的模式
		if strings.Contains(line, "func "+funcName) {
			return i + 1
		}
	}
	return 0
}

type APIComment struct {
	Summary     string
	Description string
	Tags        []string
	Deprecated  bool
	Accept      string
	Produce     string
}

func parseComments(comments string) APIComment {
	result := APIComment{}
	lines := strings.Split(comments, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "@summary"):
			result.Summary = strings.TrimSpace(strings.TrimPrefix(line, "@summary"))
		case strings.HasPrefix(line, "@description"):
			result.Description = strings.TrimSpace(strings.TrimPrefix(line, "@description"))
		case strings.HasPrefix(line, "@tags"):
			tags := strings.TrimSpace(strings.TrimPrefix(line, "@tags"))
			result.Tags = strings.Split(tags, ",")
		case strings.HasPrefix(line, "@deprecated"):
			result.Deprecated = true
		case strings.HasPrefix(line, "@accept"):
			result.Accept = strings.TrimSpace(strings.TrimPrefix(line, "@accept"))
		case strings.HasPrefix(line, "@produce"):
			result.Produce = strings.TrimSpace(strings.TrimPrefix(line, "@produce"))
		}
	}
	return result
}

func GenerateOpenAPIDoc(engine *gin.Engine, options ...OpenAPIOption) ([]byte, error) {
	if !Args.EnableOpenAPI {
		return nil, fmt.Errorf("please use --enable-openapi true flag")
	}

	var doc OpenAPIDocument
	var jsonDoc []byte
	var err error

	if utils.IsGoRun() {
		// 开发模式：解析并生成文档
		doc, err = genDoc(engine, options...)
		if err != nil {
			return nil, err
		}

		// 将文档保存到文件
		jsonDoc, err = json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return nil, err
		}
		err = os.WriteFile("openapi.json", jsonDoc, 0644)
		if err != nil {
			xlog.Error("保存 openapi.json 失败", xlog.Err(err))
		}
	} else {
		// 构建模式：直接读取保存的文件
		jsonDoc, err = os.ReadFile("openapi.json")
		if err != nil {
			return nil, fmt.Errorf("读取 openapi.json 失败: %w", err)
		}
		err = json.Unmarshal(jsonDoc, &doc)
		if err != nil {
			return nil, fmt.Errorf("解析 openapi.json 失败: %w", err)
		}
	}

	// 注册路由
	engine.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", jsonDoc)
	})

	// 注册 /docs 路由
	engine.GET("/docs", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(`<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="SwaggerUI" />
  <title>SwaggerUI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '/openapi.json',
      dom_id: '#swagger-ui',
    });
  };
</script>
	</body>
</html>`))
	})

	xlog.Debug(fmt.Sprintf("openapi started at %s", "http://"+Args.Bind+"/docs"))

	return jsonDoc, nil
}

func genDoc(engine *gin.Engine, options ...OpenAPIOption) (OpenAPIDocument, error) {
	CollectRouteInfo(engine)

	doc := OpenAPIDocument{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:   "API Documentation",
			Version: "1.0.0",
		},
		Servers: []APIServer{},
		Paths:   make(map[string]PathItem),
	}

	for _, option := range options {
		option(&doc)
	}

	// 如果没有设置服务器，添加一个默认的本地服务器
	if len(doc.Servers) == 0 {
		doc.Servers = append(doc.Servers, APIServer{
			URL:         "http://" + Args.Bind,
			Description: "Local development server",
		})
	}

	for _, api := range apiRegistry {
		// 解析路径参数
		path, pathParams := parsePathParameters(api.Path)

		// 根据路径前缀生成标签
		tags := api.Tags
		if len(tags) == 0 {
			tags = generateTags(path)
		}

		pathItem, ok := doc.Paths[path]
		if !ok {
			pathItem = PathItem{}
		}

		operation := Operation{
			Tags:        tags,
			Summary:     api.Summary,
			Description: api.Description,
			Deprecated:  api.Deprecated,
			Responses: map[string]Response{
				"200": {
					Description: "Successful response",
					Content: map[string]MediaType{
						"application/json": {
							Schema: generateSchema(api.ResponseType),
						},
					},
				},
			},
		}

		// 设置请求内容类型
		if api.Accept != "" {
			if operation.RequestBody != nil {
				newContent := make(map[string]MediaType)
				for _, acceptType := range strings.Split(api.Accept, ",") {
					acceptType = strings.TrimSpace(acceptType)
					newContent[acceptType] = operation.RequestBody.Content["application/json"]
				}
				operation.RequestBody.Content = newContent
			}
		}

		// 设置响应内容类型
		if api.Produce != "" {
			resp := operation.Responses["200"]
			newContent := make(map[string]MediaType)
			for _, produceType := range strings.Split(api.Produce, ",") {
				produceType = strings.TrimSpace(produceType)
				newContent[produceType] = resp.Content["application/json"]
			}
			resp.Content = newContent
			operation.Responses["200"] = resp
		}

		// 添加路径参数
		operation.Parameters = append(operation.Parameters, pathParams...)

		// 根据 HTTP 方法决定使用 Parameters 还是 RequestBody
		switch api.Method {
		case "GET", "DELETE":
			operation.Parameters = append(operation.Parameters, generateParameters(api.RequestType)...)
		case "POST", "PUT", "PATCH":
			operation.RequestBody = &RequestBody{
				Content: map[string]MediaType{
					"application/json": {
						Schema: generateSchema(api.RequestType),
					},
				},
			}
		}

		switch api.Method {
		case "GET":
			pathItem.Get = &operation
		case "POST":
			pathItem.Post = &operation
		case "PUT":
			pathItem.Put = &operation
		case "DELETE":
			pathItem.Delete = &operation
		case "PATCH":
			pathItem.Patch = &operation
		}

		doc.Paths[path] = pathItem
	}

	return doc, nil
}

func generateParameters(t reflect.Type) []Parameter {
	var parameters []Parameter
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		if field.Tag.Get("uri") != "" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}

		param := Parameter{
			Name:        name,
			In:          "query", // 默认为查询参数
			Description: generateDescription(field),
			Required:    isRequired(field),
			Schema:      generateSchema(field.Type),
		}
		parameters = append(parameters, param)
	}
	return parameters
}

func parsePathParameters(path string) (string, []Parameter) {
	parts := strings.Split(path, "/")
	var params []Parameter
	var newParts []string

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params = append(params, Parameter{
				Name:        paramName,
				In:          "path",
				Description: "Path parameter " + paramName,
				Required:    true,
				Schema: Schema{
					Type: "string",
				},
			})
			newParts = append(newParts, "{"+paramName+"}")
		} else {
			newParts = append(newParts, part)
		}
	}

	return strings.Join(newParts, "/"), params
}

// 根据路径生成标签
func generateTags(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		return []string{parts[0]}
	}
	return []string{"default"}
}
