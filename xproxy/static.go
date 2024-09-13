package xproxy

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/daodao97/xgo/xlog"
	"github.com/gin-gonic/gin"
)

type readSeekerWrapper struct {
	fs.File
}

func (rsw readSeekerWrapper) Seek(offset int64, whence int) (int64, error) {
	return rsw.File.(io.Seeker).Seek(offset, whence)
}

func ServeLocalFile(c *gin.Context, fs embed.FS, filePath string) bool {
	file, err := fs.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		xlog.Error("获取文件状态失败", xlog.String("路径", filePath), xlog.Err(err))
		return false
	}

	xlog.Debug("本地文件服务", xlog.String("路径", filePath), xlog.Int64("大小", stat.Size()))

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	http.ServeContent(c.Writer, c.Request, filePath, stat.ModTime(), readSeekerWrapper{file})
	return true
}

func ServeLocalContent(c *gin.Context, content string, contentType string) bool {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	c.DataFromReader(http.StatusOK, int64(len(content)), contentType, strings.NewReader(content), nil)
	return true
}
