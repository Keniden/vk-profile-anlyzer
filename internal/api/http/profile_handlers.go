package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"inteam/internal/auth"
	"inteam/internal/service"
)

func parseVKID(c *gin.Context) (int64, bool) {
	vkIDStr := c.Param("vk_id")
	vkID, err := strconv.ParseInt(vkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vk_id"})
		return 0, false
	}
	return vkID, true
}

func getProfileHandler(profileSvc service.ProfileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		vkID, ok := parseVKID(c)
		if !ok {
			return
		}

		profile, err := profileSvc.GetProfile(c.Request.Context(), vkID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
			return
		}
		if profile == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}

func analyzeProfileHandler(profileSvc service.ProfileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		vkID, ok := parseVKID(c)
		if !ok {
			return
		}

		if _, exists := c.Get(auth.ContextUserIDKey); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		profile, err := profileSvc.AnalyzeProfile(c.Request.Context(), vkID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze profile"})
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}

