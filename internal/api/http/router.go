package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"inteam/internal/auth"
	"inteam/internal/config"
	"inteam/internal/service"
)

func RegisterRoutes(
	router *gin.Engine,
	cfg *config.Config,
	profileSvc service.ProfileService,
	authSvc service.AuthService,
	jwtManager *auth.JWTManager,
) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", loginHandler(authSvc, jwtManager, cfg.Auth))
		authGroup.GET("/vk/login", vkLoginHandler(cfg.Auth))
		authGroup.GET("/vk/callback", vkCallbackHandler(cfg.Auth, jwtManager, authSvc))
	}

	protected := router.Group("/")
	protected.Use(auth.JWTMiddleware(jwtManager))
	{
		protected.GET("/me", meHandler(authSvc))
		protected.GET("/profiles/:vk_id", getProfileHandler(profileSvc))
		protected.POST("/profiles/:vk_id/analyze", analyzeProfileHandler(profileSvc))
	}

	router.Static("/static", "./internal/frontend")
}
