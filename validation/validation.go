// Package validation provides support for client input validation
// it is based on validator.v8
package validation

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/zicare/go-rpg/db"

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
	} else {
		return errors.New("Couldn't retrieve Gin's default validator engine")
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
	if err := v.RegisterValidation("json", IsJSON); err != nil {
		return err
	}
	return nil
}

//Message exported
type Message struct {
	Key string
	Msg string
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

//ErrorMessages exported
type ErrorMessages []Message

//Error exported
func (em ErrorMessages) Error() string {

	jem, _ := json.Marshal(em)
	return string(jem)
}

//GetMessages exported
func GetMessages(err error, m db.Model) ErrorMessages {

	var (
		f     = db.TAG(m, "json")
		em    ErrorMessages
		field string
		ok    bool
	)

	for _, v := range err.(validator.ValidationErrors) {
		if field, ok = f[v.Field]; !ok {
			field = v.Field
		}
		em = append(em, Message{
			Key: field,
			Msg: fmt.Sprintf("Value %v didn't pass %s(%s) validation", v.Value, v.Tag, v.Param)})
	}

	return em
}

//ValError exported
func ValError(key string, field string, value interface{}, tag string, param string) validator.ValidationErrors {
	var ve validator.ValidationErrors
	ve = make(map[string]*validator.FieldError)
	fe := validator.FieldError{Value: value, Tag: tag, Param: param, Field: field}
	ve[key] = &fe
	return ve
}
