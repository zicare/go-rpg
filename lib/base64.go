package lib

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/zicare/go-rpg/msg"
)

//Encode exported
func Encode(s interface{}) string {

	b, _ := json.Marshal(s)
	return strings.TrimRight(base64.StdEncoding.EncodeToString(b), "=")
}

//Decode exported
func Decode(src string) (string, error) {

	if l := len(src) % 4; l > 0 {
		src += strings.Repeat("=", 4-l)
	}
	decoded, err := base64.URLEncoding.DecodeString(src)
	if err != nil {
		//Decoding Error %s
		return "", msg.Get("17").SetArgs(err.Error()).M2E()
	}
	return string(decoded), nil
}
