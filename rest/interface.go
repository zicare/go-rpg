package rest

import (
	"github.com/gin-gonic/gin"
)

//ControllerInterface exported
type ControllerInterface interface {
	Index(c *gin.Context)
	IndexHead(c *gin.Context)
	Get(c *gin.Context)
	Post(c *gin.Context)
	Put(c *gin.Context)
	Delete(c *gin.Context)
}
