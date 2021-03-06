package acl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/msg"
)

type jwt struct {
	header     header
	payload    JwtPayload
	token      string
	expiration string
}

type header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

//JwtPayload exported
type JwtPayload struct {
	Issuer     string  `json:"iss"`
	IssuedAt   int64   `json:"iat"`
	Expiration int64   `json:"exp"`
	Audience   string  `json:"aud"`
	Subject    string  `json:"sub"`
	UserID     int64   `json:"user_id"`
	ParentID   int64   `json:"parent_id"`
	RoleID     int64   `json:"role_id"`
	TPS        float32 `json:"tps"`
}

//JwtToken exported
//func Token(user *int64, parent *int64, role *int64, tps *float32, duration time.Duration, secret string) (string, string) {
func JwtToken(u User, duration time.Duration, secret string) (string, string) {

	var (
		user   = u.GetUserID()
		parent = u.GetParentID()
		role   = u.GetRoleID()
		tps    = u.GetTPS()
	)

	if parent == nil {
		parent = &user
	}

	now := time.Now()
	j := new(JwtPayload{
		IssuedAt:   now.Unix(),
		Expiration: now.Add(duration).Unix(),
		UserID:     user,
		ParentID:   *parent,
		RoleID:     *role,
		TPS:        *tps,
	}, secret)

	return j.token, j.expiration

}

//JwtAuth exported
func JwtAuth(token string, secret string) (JwtPayload, *msg.Message) {

	var payload JwtPayload

	t := strings.Split(token, ".")
	if len(t) != 3 {
		//Invalid token
		return payload, msg.Get("12").M2E()
	}

	decodedPayload, PayloadErr := lib.Decode(t[1])
	if PayloadErr != nil {
		//Invalid payload
		return payload, msg.Get("13").M2E()
	}

	ParseErr := json.Unmarshal([]byte(decodedPayload), &payload)
	if ParseErr != nil {
		//Invalid payload
		return payload, msg.Get("13").M2E()
	}

	j := new(payload, secret)

	if token != j.token {
		//Token tampered
		return payload, msg.Get("14").M2E()
	}

	if j.payload.Expiration < time.Now().Unix() {
		//Token expired
		return payload, msg.Get("15").M2E()
	}

	return payload, nil

}

func new(payload JwtPayload, secret string) jwt {

	j := jwt{}
	j.header = header{
		Typ: "JWT",
		Alg: "HS256",
	}
	j.payload = payload
	j.token = j.getToken(secret)
	j.expiration = time.Unix(j.payload.Expiration, 0).Format(time.RFC3339)
	return j
}

func (j jwt) getToken(secret string) string {

	var (
		src       = encode(j.header) + "." + encode(j.payload)
		signature = hash(src, secret)
	)
	return src + "." + signature
}

func encode(s interface{}) string {

	b, _ := json.Marshal(s)
	return strings.TrimRight(base64.StdEncoding.EncodeToString(b), "=")
}

func hash(src string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(src))
	return strings.TrimRight(base64.StdEncoding.EncodeToString(h.Sum(nil)), "=")
}
