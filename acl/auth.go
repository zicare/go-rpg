package acl

import (
	"strings"
	"time"

	"github.com/zicare/go-rpg/msg"

	"github.com/gin-gonic/gin"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"golang.org/x/crypto/bcrypt"

	"github.com/zicare/go-rpg/config"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/tps"
)

//User interface exported
type User interface {
	db.Model
	GetUserID() int64
	GetParentID() *int64
	GetEmail() string
	GetPassword() []byte
	GetRoleID() *int64
	GetSystemAccessFrom() time.Time
	GetSystemAccessTo() time.Time
	GetTPS() *float32
}

//BasicAuth executes HTTP basic authentication
//email and password must be correct
//also the user's system access date range
func BasicAuth(m User) gin.HandlerFunc {

	return func(c *gin.Context) {

		email, password, ok := c.Request.BasicAuth()
		if ok == false {
			abort(c, 401, msg.Get("3")) //HTTP basic authentication required
			return
		}

		var (
			now       = time.Now()
			table     = m.View()
			fields, _ = db.Fields(m)
			ms        = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
			sb        = ms.SelectFrom(table)
			pepper    = config.Config().GetString("pepper")
		)

		sb.Select(fields.Ordered...)
		sb.Where(
			sb.Equal("email", email),
			//sb.Equal("password", password),
		)

		sql, args := sb.Build()
		//fmt.Println(sql, args)
		err := db.Db().QueryRow(sql, args...).Scan(ms.Addr(&m)...)
		if err != nil {
			//email not registered
			abort(c, 401, msg.Get("4")) //Invalid credentials
			return
		}

		//pwd creation
		//hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password+pepper), bcrypt.DefaultCost)
		//hpwd := string(hashedBytes)
		//fmt.Println(hpwd)

		//err = bcrypt.CompareHashAndPassword([]byte(*m.Password), []byte(password+pepper))
		err = bcrypt.CompareHashAndPassword(m.GetPassword(), []byte(password+pepper))
		switch err {
		case nil:
			//fmt.Println("pwd okay")
		case bcrypt.ErrMismatchedHashAndPassword:
			//log.Println("Invalid password")
			abort(c, 401, msg.Get("4")) //Invalid credentials
			return
		default:
			//log.Println("Something went wrong")
			abort(c, 500, msg.Get("5")) //Something went wrong verifying your credentials
			return
		}

		if now.Before(m.GetSystemAccessFrom()) || now.After(m.GetSystemAccessTo()) {
			abort(c, 401, msg.Get("6")) //Credentials expired or not yet valid
			return
		}

		c.Set("User", m)
		c.Next()
	}
}

//Auth exported
//Jwt must be correct and not expired
//Also, authorization towards the ACL must be passed
//Also, can't be found in acl.DeleteUsers registry
func Auth(route string) gin.HandlerFunc {

	return func(c *gin.Context) {

		var auth JwtPayload

		//authenticate jwt
		secret := config.Config().GetString("hmac_key")

		token := strings.Split(c.GetHeader("Authorization"), " ")
		if (len(token) != 2) || (token[0] != "JWT") {
			abort(c, 401, msg.Get("7")) //JWT authorization header malformed
			return
		}

		var e *msg.Message
		auth, e = JwtAuth(token[1], secret)
		if e != nil {
			abort(c, 401, *e)
			return
		}

		a := ACL()
		g := Grant{RoleID: auth.RoleID, Route: route, Method: c.Request.Method}
		r, ok := a[g]
		if !ok {
			abort(c, 401, msg.Get("8")) //Not enough permissions
			return
		}

		now := time.Now()
		if now.Before(r.From) || now.After(r.To) {
			abort(c, 401, msg.Get("9")) //Role access expired or not yet valid
			return
		}

		if tps.IsEnabled() && tps.Transaction(auth.UserID, auth.TPS) != nil {
			abort(c, 401, msg.Get("10")) //TPS limit exceeded
			return
		}

		if _, ok := DeletedUsersMap[auth.UserID]; ok {
			abort(c, 401, msg.Get("11")) //Unauthorized
			return
		}

		c.Set("Auth", auth)
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

//TsAndParentID exported
func TsAndParentID(c *gin.Context) (*time.Time, *int64) {

	ts := time.Now()
	if jp, exists := c.Get("Auth"); !exists {
		return &ts, nil
	} else if py, ok := jp.(JwtPayload); ok {
		return &ts, &py.ParentID
	}
	return &ts, nil
}

//TsAndUserID exported
func TsAndUserID(c *gin.Context) (*time.Time, *int64) {

	ts := time.Now()
	if jp, exists := c.Get("Auth"); !exists {
		return &ts, nil
	} else if py, ok := jp.(JwtPayload); ok {
		return &ts, &py.UserID
	}
	return &ts, nil
}

//UserID exported
func UserID(c *gin.Context) int64 {

	if jp, exists := c.Get("Auth"); !exists {
		return 0
	} else if py, ok := jp.(JwtPayload); ok {
		return py.UserID
	}
	return 0
}

//ParentID exported
func ParentID(c *gin.Context) int64 {

	if jp, exists := c.Get("Auth"); !exists {
		return 0
	} else if py, ok := jp.(JwtPayload); ok {
		return py.ParentID
	}
	return 0
}

//RoleID exported
func RoleID(c *gin.Context) int64 {

	if jp, exists := c.Get("Auth"); !exists {
		return 0
	} else if py, ok := jp.(JwtPayload); ok {
		return py.RoleID
	}
	return 0
}

//IsParent exported
func IsParent(c *gin.Context) bool {

	if jp, exists := c.Get("Auth"); !exists {
		return false
	} else if py, ok := jp.(JwtPayload); !ok {
		return false
	} else if py.UserID > 0 {
		return py.ParentID == py.UserID
	}
	return false
}

//Session exported
func Session(c *gin.Context) (JwtPayload, bool) {

	if jp, exists := c.Get("Auth"); !exists {
		return JwtPayload{}, false
	} else if session, ok := jp.(JwtPayload); ok {
		return session, ok
	}
	return JwtPayload{}, false
}
