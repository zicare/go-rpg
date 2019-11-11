package mw

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/acl"
)

//IBadRequestLogger exported
type IBadRequestLogger interface {
	Write(status int, t acl.JwtPayload, r *http.Request)
}

type badRequestLogWriter struct {
	gin.ResponseWriter
}

//func (w badRequestLogWriter) Write(b []byte) (int, error) {
//	return w.ResponseWriter.Write(b)
//}

//BadRequestLogger exported
func BadRequestLogger(brl IBadRequestLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer = &badRequestLogWriter{ResponseWriter: c.Writer}
		c.Next()
		if code := c.Writer.Status(); code < 400 {
			return
		} else if tpi, exists := c.Get("Auth"); !exists {
			return
		} else if tp, ok := tpi.(acl.JwtPayload); !ok {
			return
		} else {
			brl.Write(code, tp, c.Request)
		}
	}
}
