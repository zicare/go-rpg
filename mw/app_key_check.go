package mw

import (
	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/msg"

	"github.com/zicare/go-rpg/cors"
)

//AppKeyCheck exported
func AppKeyCheck() gin.HandlerFunc {

	return func(c *gin.Context) {

		cors := cors.CORS()
		//fmt.Println("mw", cors)

		key := c.GetHeader("X-App-Key")
		origin := c.GetHeader("Origin")

		if val, ok := cors[key]; !ok {
			abort(c, 401, msg.Get("28")) //Unauthorized app
			return
		} else if val != origin {
			abort(c, 401, msg.Get("28")) //Unauthorized app
			return
		}
		c.Next()
	}
}

func abort(c *gin.Context, code int, msg msg.Message) {

	c.JSON(
		code,
		gin.H{"message": msg},
	)
	c.Abort()
}
