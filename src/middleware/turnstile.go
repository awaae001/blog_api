package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"blog_api/src/config"
	"blog_api/src/model"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

type turnstileResponse struct {
	Success bool `json:"success"`
}

const maxTurnstileRequestBytes = int64(1 << 20)

// TurnstileVerify validates a Turnstile token before continuing.
func TurnstileVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		if !cfg.Verify.Turnstile.Enable {
			c.Next()
			return
		}

		token := extractTurnstileToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "turnstile token is required"})
			return
		}

		if ok := verifyTurnstile(c, cfg.Verify.Turnstile.Secret, token); !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "turnstile verification failed"})
			return
		}

		c.Next()
	}
}

// AntiBotAuth validates the short-lived anti-bot token.
func AntiBotAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractAntiBotToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "antibot token is required"})
			return
		}

		if !service.ValidateAntiBotToken(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "antibot token is invalid"})
			return
		}

		c.Set("antibot_token", token)
		c.Next()
	}
}

func extractAntiBotToken(c *gin.Context) string {
	if token := c.GetHeader("X-Antibot-Token"); token != "" {
		return token
	}

	scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
	if ok && strings.EqualFold(scheme, "Bearer") {
		return token
	}

	return ""
}

func verifyTurnstile(c *gin.Context, secret, token string) bool {
	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if ip := c.ClientIP(); ip != "" {
		form.Set("remoteip", ip)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", form)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result turnstileResponse
	return json.NewDecoder(resp.Body).Decode(&result) == nil && result.Success
}

func extractTurnstileToken(c *gin.Context) string {
	if token := c.GetHeader("X-Turnstile-Token"); token != "" {
		return token
	}
	if token := c.GetHeader("CF-Turnstile-Token"); token != "" {
		return token
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTurnstileRequestBytes)
	body, err := c.GetRawData()
	if err != nil || len(body) == 0 {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if token, ok := payload["turnstile_token"].(string); ok {
		return token
	}
	if token, ok := payload["token"].(string); ok {
		return token
	}

	return ""
}
