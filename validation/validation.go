// Package validation provides support for client input validation
// it is based on validator.v8
package validation

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/msg"

	validator "gopkg.in/go-playground/validator.v8"
)

var validate *validator.Validate

// Init method initializes a validator instance
// additionally to the one built-in on go-gin.
// Init also registers custom validation rules.
func Init() error {

	//initialize my validator and register custom fn
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)
	if err := register(validate); err != nil {
		return err
	}

	//register custom validation fn with gin built-in validator
	if bindingValidator, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := register(bindingValidator); err != nil {
			return err
		}
		//Registers an alias
		//nickname is the alias for hostname,max=128
		//bindingValidator.RegisterAliasValidation("nickname", "hostname,min=5,max=128")
	} else {
		//Couldn't retrieve Gin's default validator engine
		return msg.Get("27").M2E()
	}

	return nil
}

//register custom validations
func register(v *validator.Validate) error {

	if err := v.RegisterValidation("auto", Auto); err != nil {
		return err
	}
	if err := v.RegisterValidation("unique", Unique); err != nil {
		return err
	}
	if err := v.RegisterValidation("count", Count); err != nil {
		return err
	}
	if err := v.RegisterValidation("json", IsJSON); err != nil {
		return err
	}
	return nil
}

//Struct exported
func Struct(m db.Model) error {

	validate.RegisterStructValidation(m.Validation, m.Val())
	if ve := validate.Struct(m); ve != nil {
		return ve
		//return GetMessages(ve)
	}
	return nil
}

//GetMessages exported
func GetMessages(err error, m db.Model) (eml msg.MessageList) {

	var (
		f     = db.TAG(m, "json")
		field string
		ok    bool
	)

	switch err.(type) {
	case *time.ParseError:
		//Time %s has a wrong format, required format is %s
		e := err.(*time.ParseError)
		m := msg.Get("22").SetArgs(lib.TrimQuotes(e.Value), "2006-01-02T15:04:05-07:00")
		eml = append(eml, m)
	case *json.UnmarshalTypeError:
		//Value is a %s, required type is %s
		e := err.(*json.UnmarshalTypeError)
		m := msg.Get("23").SetArgs(e.Value, e.Type.String()).SetField(e.Field)
		eml = append(eml, m)
	case validator.ValidationErrors:
		for _, v := range err.(validator.ValidationErrors) {
			if field, ok = f[v.Field]; !ok {
				field = v.Field
			}
			key := strings.Split(field, ",")
			for _, jv := range key {
				if jv == "omitempty" {
					v.Value = ""
				}
			}
			//Value %s didn't pass %s(%s) validation
			m := msg.Get("24").SetArgs(v.Value, v.Tag, v.Param).SetField(key[0])
			eml = append(eml, m)
		}
	}

	return eml
}

//ValError exported
func ValError(key string, field string, value interface{}, tag string, param string) validator.ValidationErrors {
	var ve validator.ValidationErrors
	ve = make(map[string]*validator.FieldError)
	fe := validator.FieldError{Value: value, Tag: tag, Param: param, Field: field}
	ve[key] = &fe
	return ve
}
