package authHandler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"blog_api/src/config"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

// SendEmailCode handles POST /api/verify/email request.
// If code is provided, it confirms the code and returns a token; otherwise it sends a code.
func (h *VerifyHandler) SendEmailCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return
	}

	if req.Code != "" {
		if !service.ValidateEmailVerifyCode(req.Email, req.Code) {
			c.JSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "invalid email verification code"})
			return
		}

		token, expiresAt, err := service.IssueEmailToken(req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to issue email token"})
			return
		}

		result, err := friendsRepositories.QueryFriendLinks(h.DB, model.FriendLinkQueryOptions{
			Email: req.Email,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to retrieve friend link"})
			return
		}
		friendLinks := make([]model.FriendLinkDTO, 0, len(result.Links))
		for _, link := range result.Links {
			friendLinks = append(friendLinks, link.ToDTO(false))
		}

		c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: map[string]interface{}{
			"token":        token,
			"expires_at":   expiresAt,
			"expires_in":   service.EmailTokenTTLSeconds,
			"friend_links": friendLinks,
		}})
		return
	}

	cfg := config.GetConfig()
	code, expiresAt, err := service.IssueEmailVerifyCode(req.Email)
	if err != nil {
		if errors.Is(err, service.ErrEmailCodeResendTooSoon) {
			c.JSON(http.StatusTooManyRequests, model.ApiResponse{Code: 429, Message: "verification code requested too frequently"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to issue email code"})
		return
	}

	content := service.EmailContent{
		Subject: "blog api 登录验证",
		Body:    fmt.Sprintf("你的验证码是： %s 。 它将在 %d 后过期。", code, service.EmailCodeTTLSeconds/60),
		IsHTML:  false,
	}
	if !cfg.Email.Enable {
		log.Printf("[email][disabled] To=%s Subject=%s Body=%s", req.Email, content.Subject, content.Body)
		if !cfg.IsDev {
			c.JSON(http.StatusServiceUnavailable, model.ApiResponse{Code: 503, Message: "email service is disabled"})
			return
		}
	} else if err := service.SendEmail(cfg.Email, []string{req.Email}, content); err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: map[string]interface{}{
		"expires_at": expiresAt,
		"expires_in": service.EmailCodeTTLSeconds,
	}})
}
