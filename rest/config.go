package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/config"
)

//ConfigController exported
type ConfigController struct{}

//Get exported
func (ctrl ConfigController) Get(c *gin.Context) {

	c.JSON(http.StatusOK, config.Config().AllSettings())
}

//Put exported
func (ctrl ConfigController) Put(c *gin.Context) {
}
