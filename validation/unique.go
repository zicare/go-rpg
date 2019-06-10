package validation

import (
	"reflect"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/slice"
	"gopkg.in/go-playground/validator.v8"
)

//Unique exported
func Unique(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {

	var (
		f     = strings.Split(param, " ")
		m     = currentStructOrField.Interface().(db.Model)
		t     = reflect.Indirect(reflect.ValueOf(m))
		sb    = sqlbuilder.PostgreSQL.NewSelectBuilder()
		count = 0
	)

	sb.Select(sb.As("COUNT(*)", "count"))
	sb.From(m.Table())

	for i := 0; i < t.NumField(); i++ {
		if tag, ok := t.Type().Field(i).Tag.Lookup("db"); ok {
			if fv, _, ok := v.GetStructFieldOK(currentStructOrField, t.Type().Field(i).Name); ok {
				if pk, _ := t.Type().Field(i).Tag.Lookup("primary"); pk == "1" {
					sb.Where(sb.NotEqual(tag, fv.Interface()))
				} else if slice.Contains(f, tag) {
					sb.Where(sb.Equal(tag, fv.Interface()))
				}
			}
		}
	}

	sql, args := sb.Build()
	//log.Println(param, f)
	//log.Println(sql, args)
	if err := db.Db().QueryRow(sql, args...).Scan(&count); err != nil || count > 0 {
		return false
	}
	return true
}

/*
//Only check uniqueness in one column
func Unique(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {

	var (
		m     = currentStructOrField.Interface().(db.Model)
		t     = reflect.Indirect(reflect.ValueOf(m))
		sb    = sqlbuilder.PostgreSQL.NewSelectBuilder()
		count = 0
	)

	sb.Select(sb.As("COUNT(*)", "count"))
	sb.From(m.Table())
	sb.Where(sb.Equal(param, field.Interface()))

	for i := 0; i < t.NumField(); i++ {
		pk, _ := t.Type().Field(i).Tag.Lookup("primary")
		if pk == "1" {
			tag, ok := t.Type().Field(i).Tag.Lookup("db")
			if ok {
				fv, _, _ := v.GetStructFieldOK(currentStructOrField, t.Type().Field(i).Name)
				sb.Where(sb.NotEqual(tag, fv.Interface()))
			}
		}
	}

	sql, args := sb.Build()
	log.Println(sql, args)
	if err := db.Db().QueryRow(sql, args...).Scan(&count); err != nil || count > 0 {
		return false
	}
	return true
}
*/
