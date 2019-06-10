package rest

import (
	"github.com/gin-gonic/gin"
)

//ControllerInterface exported
type ControllerInterface interface {
	Index(c *gin.Context)
	Get(c *gin.Context)
}
