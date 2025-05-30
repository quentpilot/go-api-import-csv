package handlers

import (
	"go-csv-import/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	logger.Info("Call endpoint /ping")
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "API is running",
	})
	logger.Info("Health check response sent")
}
