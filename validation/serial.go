package validation

import (
	"reflect"

	"gopkg.in/go-playground/validator.v8"
)

//Serial exported
func Serial(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {

	//log.Println("field", field.Interface())
	return false
}
