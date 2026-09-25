package handler

import (
	"blog_api/src/middleware"

	"github.com/gin-gonic/gin"
)

// handleCanceledQuery 判断查询错误是否由客户端断开（请求上下文取消）引起。
func handleCanceledQuery(c *gin.Context, err error) bool {
	if middleware.IsRequestCanceled(err) || middleware.IsClientGone(c) {
		c.AbortWithStatus(middleware.ClientGoneStatus) // 写入会被 ClientDisconnectGuard 抑制
		return true
	}
	return false
}
