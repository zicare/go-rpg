package validation

import (
	"reflect"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/zicare/go-rpg/db"
	"gopkg.in/go-playground/validator.v8"
)

//Count exported
func Count(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {

	var (
		f     = strings.Split(param, ".")
		sb    = sqlbuilder.PostgreSQL.NewSelectBuilder()
		count = 0
	)

	sb.Select(sb.As("COUNT(*)", "count"))
	sb.From(f[0])
	sb.Where(sb.Equal(f[1], field.Interface()))

	sql, args := sb.Build()
	//fmt.Println(sql, args)
	if err := db.Db().QueryRow(sql, args...).Scan(&count); err != nil || count < 1 {
		return false
	}
	return true
}
