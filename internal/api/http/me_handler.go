package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"inteam/internal/auth"
	"inteam/internal/service"
)

func meHandler(authSvc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get(auth.ContextUserIDKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, ok := userIDVal.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}

		user, err := authSvc.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}
		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":    user.ID,
			"email": user.Email,
			"vk_id": user.VKID,
		})
	}
}

