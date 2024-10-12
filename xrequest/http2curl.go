package xrequest

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
)

var groupSplit = "<http-curl-group-xrequest>"

// CurlCommand contains exec.Command compatible slice + helpers
type CurlCommand []string

// append appends a string to the CurlCommand
func (c *CurlCommand) append(newSlice ...string) {
	*c = append(*c, newSlice...)
}

// String returns a ready to copy/paste command
func (c *CurlCommand) String() string {
	groups := strings.Join(*c, " ")
	parts := strings.Split(groups, groupSplit)
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return strings.Join(parts, " \\\n")
}

func bashEscape(str string) string {
	return `'` + strings.Replace(str, `'`, `'\''`, -1) + `'`
}

// GetCurlCommand returns a CurlCommand corresponding to an http.Request
func GetCurlCommand(req *http.Request) (*CurlCommand, error) {
	if req.URL == nil {
		return nil, fmt.Errorf("getCurlCommand: invalid request, req.URL is nil")
	}

	command := CurlCommand{}

	command.append("curl")

	schema := req.URL.Scheme
	requestURL := req.URL.String()
	if schema == "" {
		schema = "http"
		if req.TLS != nil {
			schema = "https"
		}
		requestURL = schema + "://" + req.Host + req.URL.Path
	}

	if schema == "https" {
		command.append("-k")
	}

	command.append("-X", bashEscape(req.Method), groupSplit)

	if req.Body != nil {
		contentType := req.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// 保存原始的请求体
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("getCurlCommand: read body error: %w", err)
			}
			// 重置 body 以便后续读取
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 解析 multipart form 数据
			_, params, err := mime.ParseMediaType(contentType)
			if err != nil {
				return nil, fmt.Errorf("getCurlCommand: parse content-type error: %w", err)
			}
			boundary, ok := params["boundary"]
			if !ok {
				return nil, fmt.Errorf("getCurlCommand: no boundary found in content-type")
			}

			reader := multipart.NewReader(bytes.NewReader(bodyBytes), boundary)
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("getCurlCommand: read multipart error: %w", err)
				}

				if part.FileName() != "" {
					command.append("-F", fmt.Sprintf("%s=@%s", part.FormName(), part.FileName()), groupSplit)
				} else {
					// 对于非文件字段，我们可以读取并包含其值
					value, _ := io.ReadAll(part)
					command.append("-F", fmt.Sprintf("%s=%s", part.FormName(), string(value)), groupSplit)
				}
			}

			// 再次重置 body，确保原始内容被保留
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		} else {
			var buff bytes.Buffer
			_, err := buff.ReadFrom(req.Body)
			if err != nil {
				return nil, fmt.Errorf("getCurlCommand: buffer read from body error: %w", err)
			}
			// reset body for potential re-reads
			req.Body = io.NopCloser(bytes.NewBuffer(buff.Bytes()))
			if len(buff.String()) > 0 {
				bodyEscaped := bashEscape(buff.String())
				command.append("-d", bodyEscaped, groupSplit)
			}
		}
	}

	var keys []string

	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		command.append("-H", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " "))), groupSplit)
	}

	command.append(bashEscape(requestURL))

	// command.append("--compressed")

	return &command, nil
}
