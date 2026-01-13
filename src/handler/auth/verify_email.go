package authHandler

import (
	"fmt"
	"net/http"

	"blog_api/src/config"
	"blog_api/src/model"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

// SendEmailCode handles POST /api/verify/email request.
// If code is provided, it confirms the code and returns a token; otherwise it sends a code.
func (h *VerifyHandler) SendEmailCode(c *gin.Context) {
	var req model.EmailVerifyConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid request body"))
		return
	}

	if req.Code != "" {
		if !service.ValidateEmailVerifyCode(req.Email, req.Code) {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(401, "invalid email verification code"))
			return
		}

		token, expiresAt, err := service.IssueEmailToken(req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to issue email token"))
			return
		}

		c.JSON(http.StatusOK, model.NewSuccessResponse(map[string]interface{}{
			"token":      token,
			"expires_at": expiresAt,
			"expires_in": service.EmailTokenTTLSeconds(),
		}))
		return
	}

	cfg := config.GetConfig()
	if !cfg.Email.Enable {
		c.JSON(http.StatusServiceUnavailable, model.NewErrorResponse(503, "email service is disabled"))
		return
	}

	code, expiresAt, err := service.IssueEmailVerifyCode(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to issue email code"))
		return
	}

	content := service.EmailContent{
		Subject: "Friend link verification code",
		Body:    fmt.Sprintf("Your verification code is %s. It expires in %d minutes.", code, service.EmailCodeTTLSeconds()/60),
		IsHTML:  false,
	}
	if err := service.SendEmail(cfg.Email, []string{req.Email}, content); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to send verification email"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(map[string]interface{}{
		"expires_at": expiresAt,
		"expires_in": service.EmailCodeTTLSeconds(),
	}))
}
