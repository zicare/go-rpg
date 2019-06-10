package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/zicare/go-rpg/lib"
)

type jwt struct {
	header     header
	payload    Payload
	token      string
	expiration string
}

type header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

//Payload exported
type Payload struct {
	Issuer     string `json:"iss"`
	IssuedAt   int64  `json:"iat"`
	Expiration int64  `json:"exp"`
	Audience   string `json:"aud"`
	Subject    string `json:"sub"`
	UserID     int64  `json:"user_id"`
	RoleID     int64  `json:"role_id"`
}

//Token exported
func Token(user *int64, role *int64, duration time.Duration, secret string) (string, string) {

	now := time.Now()
	j := new(Payload{
		IssuedAt:   now.Unix(),
		Expiration: now.Add(duration).Unix(),
		UserID:     *user,
		RoleID:     *role,
	}, secret)

	return j.token, j.expiration

}

//Auth exported
func Auth(token string, secret string) (Payload, error) {

	var payload Payload

	t := strings.Split(token, ".")
	if len(t) != 3 {
		return payload, errors.New("Invalid token")
	}

	decodedPayload, PayloadErr := lib.Decode(t[1])
	if PayloadErr != nil {
		return payload, errors.New("Invalid payload")
	}

	ParseErr := json.Unmarshal([]byte(decodedPayload), &payload)
	if ParseErr != nil {
		return payload, errors.New("Invalid payload")
	}

	j := new(payload, secret)

	if token != j.token {
		return payload, errors.New("Token tampered")
	}

	if j.payload.Expiration < time.Now().Unix() {
		return payload, errors.New("Token expired")
	}

	return payload, nil

}

func new(payload Payload, secret string) jwt {

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
