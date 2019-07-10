package validation

import (
	"encoding/json"
	"reflect"

	"gopkg.in/go-playground/validator.v8"
)

//IsJSON exported
func IsJSON(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {

	//return false
	var js json.RawMessage
	return json.Unmarshal([]byte(field.String()), &js) == nil
}
