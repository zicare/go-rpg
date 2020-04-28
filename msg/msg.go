package msg

import (
	"encoding/json"
	"fmt"
)

var msg map[string]Message

//Message exported
type Message struct {
	Key   string
	Msg   string
	Args  []interface{}
	Field string
}

//New exported
func New(k string, m string) Message {
	return Message{k, m, nil, ""}
}

//SetArgs exported
func (m Message) SetArgs(args ...interface{}) Message {
	m.Args = args
	return m
}

//SetField exported
func (m Message) SetField(field string) Message {
	m.Field = field
	return m
}

//M2E exported
func (m Message) M2E() *Message {
	return &m
}

//Error exported
func (m *Message) Error() string {
	return m.String()
}

//String exported
func (m Message) String() string {

	if m.Args != nil && len(m.Args) > 0 {
		return fmt.Sprintf(m.Msg, m.Args)
	}
	return m.Msg
}

//MessageList exported
type MessageList []Message

//Error exported
func (ml MessageList) Error() string {

	jml, _ := json.Marshal(ml)
	return string(jml)
}

//Init exported
func Init(m []Message) (err error) {

	_init()

	//add client messages
	for _, v := range m {
		if _, ok := msg[v.Key]; ok {
			m := Get("1") //Invalid message key
			return &m
		}
		msg[v.Key] = v
	}

	return nil
}

//GetAll exported
func GetAll() map[string]Message {
	return msg
}

//Get exported
func Get(key string) Message {

	if message, ok := msg[key]; ok {
		return message
	}
	return New(key, fmt.Sprintf("Message %v", key))
}

func _init() {

	msg = make(map[string]Message)

	msg["1"] = New("1", "Invalid message key")
	msg["2"] = New("2", "ACL tags are not properly set")
	msg["3"] = New("3", "HTTP basic authentication required")
	msg["4"] = New("4", "Invalid credentials")
	msg["5"] = New("5", "Something went wrong verifying your credentials")
	msg["6"] = New("6", "Credentials expired or not yet valid")
	msg["7"] = New("7", "JWT authorization header malformed")
	msg["8"] = New("8", "Not enough permissions")
	msg["9"] = New("9", "Role access expired or not yet valid")
	msg["10"] = New("10", "TPS limit exceeded")
	msg["11"] = New("11", "Unauthorized")
	msg["12"] = New("12", "Invalid token")
	msg["13"] = New("13", "Invalid payload")
	msg["14"] = New("14", "Token tampered")
	msg["15"] = New("15", "Token expired")
	msg["16"] = New("16", "Read only model")
	msg["17"] = New("17", "Decoding Error %s")
	msg["18"] = New("18", "No found!")
	msg["19"] = New("19", "There are validation errors")
	msg["20"] = New("20", "A minimum of %s calls/chcap required to calculate TPS")
	msg["21"] = New("21", "TPS data clean up cycles must be %s seconds or longer")
	msg["22"] = New("22", "Time %s has a wrong format, required format is %s")
	msg["23"] = New("23", "Value is a %s, required type is %s")
	msg["24"] = New("24", "Value %s didn't pass %s(%s) validation")
	msg["25"] = New("25", "Server error: %s")
	msg["26"] = New("26", "Composite key missuse")
	msg["27"] = New("27", "Couldn't retrieve Gin's default validator engine")
}
