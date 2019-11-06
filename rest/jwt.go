package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/acl"
	"github.com/zicare/go-rpg/config"
	"github.com/zicare/go-rpg/jwt"
)

//JwtController exported
type JwtController struct{}

//Get exported
func (ctrl JwtController) Get(c *gin.Context) {

	var u acl.User
	var ok bool

	if user, exists := c.Get("User"); !exists {
		c.AbortWithStatus(500)
		return
	} else if u, ok = user.(acl.User); !ok {
		c.AbortWithStatus(500)
		return
	}

	var (
		secret      = config.Config().GetString("hmac_key")
		duration, _ = time.ParseDuration(config.Config().GetString("jwt_duration"))
	)

	//adjust token duration for users with credentials to expire before
	//the default token duration
	now := time.Now()
	if u.GetSystemAccessTo().Before(now.Add(duration)) {
		duration = u.GetSystemAccessTo().Sub(now)
	}

	id := u.GetUserID()
	token, expiration := jwt.Token(&id, u.GetParentID(), u.GetRoleID(), duration, secret)
	c.JSON(http.StatusOK, gin.H{"token": token, "expiration": expiration})
}
