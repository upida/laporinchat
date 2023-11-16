package admin

import (
	"laporinchat/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username"`
}

func Register(c *gin.Context) {
	var request RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminRequest, _ := models.SetAdminRequest(request.Username)
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": adminRequest})
	return
}
