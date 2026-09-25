package middleware

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

const ClientGoneStatus = 499

func ClientDisconnectGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer = &clientGoneWriter{ResponseWriter: c.Writer, ctx: c.Request.Context()}

		if IsClientGone(c) {
			c.AbortWithStatus(ClientGoneStatus) // 实际写入会被包装层抑制
			return
		}

		c.Next()

		if !c.Writer.Written() && IsClientGone(c) {
			c.Abort()
			c.Writer.WriteHeader(ClientGoneStatus)
		}
	}
}

func IsClientGone(c *gin.Context) bool {
	return c != nil && c.Request != nil && c.Request.Context().Err() != nil
}

// IsRequestCanceled 判断错误是否由请求上下文取消（客户端断开 / 请求超时）引起。
func IsRequestCanceled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// clientGoneWriter 包装 gin.ResponseWriter，在请求上下文取消后抑制一切写入。
type clientGoneWriter struct {
	gin.ResponseWriter
	ctx              context.Context
	suppressed       bool
	suppressedStatus int
}

func (w *clientGoneWriter) gone() bool {
	return w.ctx.Err() != nil
}

// suppress 记录一次被抑制的写入，并保留状态码供访问日志使用。
func (w *clientGoneWriter) suppress(status int) {
	if !w.suppressed {
		w.suppressed = true
		w.suppressedStatus = status
	}
}

func (w *clientGoneWriter) WriteHeader(code int) {
	if w.gone() {
		if !w.ResponseWriter.Written() {
			w.suppress(code)
		}
		return
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *clientGoneWriter) WriteHeaderNow() {
	if w.gone() {
		if !w.ResponseWriter.Written() {
			w.suppress(w.ResponseWriter.Status())
		}
		return
	}
	w.ResponseWriter.WriteHeaderNow()
}

func (w *clientGoneWriter) Write(data []byte) (int, error) {
	if w.gone() {
		w.suppress(w.Status())
		return 0, w.ctx.Err()
	}
	return w.ResponseWriter.Write(data)
}

func (w *clientGoneWriter) WriteString(s string) (int, error) {
	if w.gone() {
		w.suppress(w.Status())
		return 0, w.ctx.Err()
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *clientGoneWriter) Flush() {
	if w.gone() {
		return
	}
	w.ResponseWriter.Flush()
}

func (w *clientGoneWriter) Status() int {
	if w.suppressed {
		return w.suppressedStatus
	}
	return w.ResponseWriter.Status()
}

func (w *clientGoneWriter) Written() bool {
	return w.suppressed || w.ResponseWriter.Written()
}

func (w *clientGoneWriter) Size() int {
	if w.suppressed {
		return 0
	}
	return w.ResponseWriter.Size()
}
