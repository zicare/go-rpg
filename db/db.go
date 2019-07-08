package db

import (
	"database/sql"
	"fmt"
	"reflect"

	//required
	_ "github.com/lib/pq"
	"github.com/zicare/go-rpg/config"
	"github.com/zicare/go-rpg/lib"
	"gopkg.in/go-playground/validator.v8"
)

//SelectOpt exported
type SelectOpt struct {
	Offset   int
	Limit    int
	Column   []string
	Filter   map[string][]lib.Pair
	Null     []string
	NotNull  []string
	Order    []string
	Checksum int
}

//ResultSetMeta exported
type ResultSetMeta struct {
	Range    string
	Checksum string
}

//Model exported
type Model interface {
	New() Model
	Table() string
	Val() interface{}
	Xfrm() Model
	FilterInput(pIDs []lib.Pair) error
	StructLevelValidation(*validator.Validate, *validator.StructLevel)
}

/*
 * Model interface implementation example
 *
 * type Person struct {
 *	 PersonID  *int64     `db:"person_id"  json:"person_id,omitempty" binding:"omitempty,serial" primary:"1"`
 *	 FirstName *string    `db:"first_name" json:"first_name"          binding:"-"`
 *	 LastName  *string    `db:"last_name"  json:"last_name"           binding:"-"`
 * }
 *
 * func (*Person) New() db.Model {
 *	 return new(Person)
 * }
 *
 * func (*Person) Table() string {
 *   return "persons"
 * }
 *
 * func (person *Person) Val() interface{} {
 * 	return *person
 * }
 *
 *
 * func (person *Person) Xfrm() db.Model { //make changes to person before sending the output on GET/HEAD requests
 * 	return person
 * }
 *
 * func (person *Person) FilterInput(pIDs []lib.Pair) error { //make changes to submitted person before struct level validation on POST/PUT requests
 *
 * 	 if len(pIDs) > 0 {
 * 		id, _ := strconv.ParseInt(pIDs[0].B.(string), 10, 64)
 * 		person.PersonID = &id
 * 	 } else {
 * 		id := int64(0)
 * 		person.PersonID = &id
 * 	 }
 *
 * 	 return validation.Struct(person)
 * }
 *
 * func (*Person) StructLevelValidation(v *validator.Validate, structLevel *validator.StructLevel) { //make final struct level validation before trying to persist person to db
 *
 * 	person := structLevel.CurrentStruct.Interface().(Person)
 *
 * 	if person.FirstName == nil && person.LastName == nil {
 * 		structLevel.ReportError(reflect.ValueOf(person.FirstName), "FirstName", "first_name", "FirstNameOrLastName")
 * 		structLevel.ReportError(reflect.ValueOf(person.LastName), "LastName", "last_name", "FirstNameOrLastName")
 * 	}
 *
 * }
 *
 */

//Cols exported
func Cols(m Model) (cols map[string]interface{}, colsOrdered []string, pk []string) {

	var (
		v = reflect.ValueOf(m.Val())
		t = reflect.Indirect(v)
	)

	cols = make(map[string]interface{})
	colsOrdered = []string{}
	pk = []string{}

	for i := 0; i < t.NumField(); i++ {
		k, ok := t.Type().Field(i).Tag.Lookup("db")
		if ok && k != "-" {
			cols[k] = v.Field(i).Interface()
			colsOrdered = append(colsOrdered, k)
			p, _ := t.Type().Field(i).Tag.Lookup("primary")
			if p == "1" {
				pk = append(pk, k)
			}
		}
	}

	return
}

var db *sql.DB

//Init tests the db connection and saves the db handler
func Init() (err error) {

	var (
		c = config.Config()
		//conn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		conn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			c.GetString("db.user"),
			c.GetString("db.password"),
			c.GetString("db.host"),
			c.GetString("db.port"),
			c.GetString("db.name"))
	)

	//db, err = sql.Open("mysql", conn)
	db, err = sql.Open("postgres", conn)
	if err != nil {
		return
	}

	err = db.Ping()
	if err != nil {
		return
	}

	db.SetMaxOpenConns(c.GetInt("db.max_open_conns"))

	return
}

//Db returns the db handler
func Db() *sql.DB {

	return db
}
