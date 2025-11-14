package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"inteam/internal/auth"
	"inteam/internal/config"
	"inteam/internal/service"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func loginHandler(authSvc service.AuthService, jwtManager *auth.JWTManager, authCfg config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		access, refresh, err := authSvc.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if authCfg.CookieHTTPOnly {
			c.SetCookie(authCfg.AccessTokenCookie, access, int(authCfg.AccessTokenTTL.Seconds()), "/", authCfg.CookieDomain, authCfg.CookieSecure, authCfg.CookieHTTPOnly)
			c.SetCookie(authCfg.RefreshTokenCookie, refresh, int(authCfg.RefreshTokenTTL.Seconds()), "/", authCfg.CookieDomain, authCfg.CookieSecure, authCfg.CookieHTTPOnly)
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  access,
			"refresh_token": refresh,
		})
	}
}

func vkLoginHandler(authCfg config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		v := url.Values{}
		v.Set("client_id", authCfg.VKClientID)
		v.Set("redirect_uri", authCfg.VKRedirectURL)
		v.Set("response_type", "code")
		v.Set("scope", "friends,wall,offline")
		v.Set("display", "page")

		redirectURL := fmt.Sprintf("https://oauth.vk.com/authorize?%s", v.Encode())
		c.Redirect(http.StatusFound, redirectURL)
	}
}

func vkCallbackHandler(authCfg config.AuthConfig, jwtManager *auth.JWTManager, authSvc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}

		values := url.Values{}
		values.Set("client_id", authCfg.VKClientID)
		values.Set("client_secret", authCfg.VKClientSecret)
		values.Set("redirect_uri", authCfg.VKRedirectURL)
		values.Set("code", code)

		resp, err := http.Get("https://oauth.vk.com/access_token?" + values.Encode())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "vk oauth request failed"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			c.JSON(http.StatusBadGateway, gin.H{"error": "vk oauth error"})
			return
		}

		var body struct {
			AccessToken string `json:"access_token"`
			UserID      int64  `json:"user_id"`
			Email       string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid vk oauth response"})
			return
		}

		if body.Email == "" {
			body.Email = fmt.Sprintf("vk_%d@example.com", body.UserID)
		}

		user, err := authSvc.Register(c.Request.Context(), body.Email, "")
		if err != nil && err.Error() != "user already exists" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
		if user == nil {
			// user already exists; login
			access, refresh, err := authSvc.Login(c.Request.Context(), body.Email, "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login user"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"access_token":  access,
				"refresh_token": refresh,
			})
			return
		}

		access, refresh, err := jwtManager.GenerateTokens(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  access,
			"refresh_token": refresh,
		})
	}
}

