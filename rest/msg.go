package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/msg"
)

//MsgController exported
type MsgController struct{}

//Index exported
func (ctrl MsgController) Index(c *gin.Context) {

	var data []interface{}
	for _, v := range msg.GetAll() {
		data = append(data, v)
	}
	c.JSON(http.StatusOK, data)
}
