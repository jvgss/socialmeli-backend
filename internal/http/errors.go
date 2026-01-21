package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func okNoBody(c *gin.Context) {
	c.Status(http.StatusOK)
}
