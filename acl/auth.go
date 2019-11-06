package acl

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"golang.org/x/crypto/bcrypt"

	"github.com/zicare/go-rpg/config"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/jwt"
)

var auth jwt.Payload

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
}

//BasicAuth executes HTTP basic authentication
//email and password must be correct
//also the user's system access date range
func BasicAuth(m User) gin.HandlerFunc {

	return func(c *gin.Context) {

		email, password, ok := c.Request.BasicAuth()
		if ok == false {
			abort(c, 401, "HTTP basic authentication required")
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
			abort(c, 401, "Invalid credentials")
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
			abort(c, 401, "Invalid credentials")
			return
		default:
			//log.Println("Something went wrong")
			abort(c, 500, "Something went wrong verifying your credentials")
			return
		}

		if now.Before(m.GetSystemAccessFrom()) || now.After(m.GetSystemAccessTo()) {
			abort(c, 401, "Credentials expired or not yet valid")
			return
		}

		c.Set("User", m)
		c.Next()
	}
}

//Auth exported
//Jwt must be correct and not expired
//Also, authorization towards the ACL must be passed
func Auth(route string) gin.HandlerFunc {

	return func(c *gin.Context) {

		//authenticate jwt
		secret := config.Config().GetString("hmac_key")

		token := strings.Split(c.GetHeader("Authorization"), " ")
		if (len(token) != 2) || (token[0] != "JWT") {
			abort(c, 401, "JWT authorization header malformed")
			return
		}

		var err error
		auth, err = jwt.Auth(token[1], secret)
		if err != nil {
			abort(c, 401, err.Error())
			return
		}

		a := ACL()
		g := Grant{RoleID: auth.RoleID, Route: route, Method: c.Request.Method}
		r, ok := a[g]
		if !ok {
			abort(c, 401, "Not enough permissions")
			return
		}

		now := time.Now()
		if now.Before(r.From) || now.After(r.To) {
			abort(c, 401, "Role access expired or not yet valid")
			return
		}

		c.Set("Auth", auth)
		c.Next()
	}
}

//Token exported
func Token() jwt.Payload {
	return auth
}

func abort(c *gin.Context, code int, msg string) {

	c.JSON(
		code,
		gin.H{"message": msg},
	)
	c.Abort()
}

//TsAndUserID exported
func TsAndUserID() (*time.Time, *int64) {

	var (
		ts  = time.Now()
		uid = auth.UserID
	)

	return &ts, &uid
}

//ParentID exported
func ParentID() int64 {

	return auth.ParentID
}
