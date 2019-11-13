package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"

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

//ScopedModel interface exported
type ScopedModel interface {
	Model
	ScopeOk(c *gin.Context) bool
}

//Model exported
type Model interface {
	New() Model
	Table() string
	View() string
	Val() interface{}
	Xfrm() Model
	Bind(*gin.Context, []lib.Pair) error
	Validation(*validator.Validate, *validator.StructLevel)
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
 * func (*Person) View() string {
 *   return "view_persons"
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
 * func (person *Person) Bind(c *gin.Context, pIDs []lib.Pair) error {
 *
 * 	if err := c.ShouldBind(person); err != nil {
 * 		return err
 * 	} else if len(pIDs) == 1 {
 * 		person.PersonID, _ = strconv.ParseInt(pIDs[0].B.(string), 10, 64)
 * 	}
 * 	person.LastEditTs, person.LastEditUserID = auth.TsAndUserID()
 * 	return validation.Struct(person)
 * }
 *
 * func (*Person) Validation(v *validator.Validate, sl *validator.StructLevel) { //make final struct level validation before trying to persist person to db
 *
 * 	person := sl.CurrentStruct.Interface().(Person)
 *
 * 	if person.FirstName == nil && person.LastName == nil {
 * 		sl.ReportError(reflect.ValueOf(person.FirstName), "FirstName", "first_name", "FirstNameOrLastName")
 * 		sl.ReportError(reflect.ValueOf(person.LastName), "LastName", "last_name", "FirstNameOrLastName")
 * 	}
 *
 * }
 *
 */

//Meta exported
type Meta struct {
	Ordered  []string
	Primary  []string
	Serial   []string
	View     []string
	Writable []string
}

//Fields exported
func Fields(m Model) (meta Meta, val map[string]interface{}) {

	var (
		v = reflect.ValueOf(m.Val())
		t = reflect.Indirect(v)
	)

	val = make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		k, ok := t.Type().Field(i).Tag.Lookup("db")
		if ok && k != "-" {
			val[k] = v.Field(i).Interface()
			meta.Ordered = append(meta.Ordered, k)
			//check for primary
			if primary, _ := t.Type().Field(i).Tag.Lookup("primary"); primary == "1" {
				meta.Primary = append(meta.Primary, k)
				//pID = append(pID, lib.Pair{A: k, B: fmt.Sprintf("%v", val[k])})
			}
			//check for serial
			if serial, _ := t.Type().Field(i).Tag.Lookup("serial"); serial == "1" {
				meta.Serial = append(meta.Serial, k)
			}
			//check for view or writable
			if view, ok := t.Type().Field(i).Tag.Lookup("view"); !ok {
				meta.Writable = append(meta.Writable, k)
			} else if view == "1" {
				meta.View = append(meta.View, k)
			}
		}
	}
	return
}

/*
//Cols exported
func Cols(m Model) (cols map[string]interface{}, colsOrdered []string, pk []string, sk []string, vw []string) {

	var (
		v = reflect.ValueOf(m.Val())
		t = reflect.Indirect(v)
	)

	cols = make(map[string]interface{})
	colsOrdered = []string{}
	pk = []string{}
	sk = []string{}
	vw = []string{}

	for i := 0; i < t.NumField(); i++ {
		k, ok := t.Type().Field(i).Tag.Lookup("db")
		if ok && k != "-" {
			cols[k] = v.Field(i).Interface()
			colsOrdered = append(colsOrdered, k)
			p, _ := t.Type().Field(i).Tag.Lookup("primary")
			if p == "1" {
				pk = append(pk, k)
			}
			s, _ := t.Type().Field(i).Tag.Lookup("serial")
			if s == "1" {
				sk = append(sk, k)
			}
			w, _ := t.Type().Field(i).Tag.Lookup("view")
			if w == "1" {
				vw = append(vw, k)
			}
		}
	}

	return
}
*/

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

//PID exported
func PID(m Model, f []string) (pID []lib.Pair) {

	var (
		v = reflect.ValueOf(m.Val())
		t = reflect.Indirect(v)
	)

	for i := 0; i < t.NumField(); i++ {
		k, ok := t.Type().Field(i).Tag.Lookup("primary")
		if ok && k == "1" {
			c, _ := t.Type().Field(i).Tag.Lookup("db")
			pID = append(pID, lib.Pair{A: c, B: fmt.Sprintf("%v", v.Field(i).Interface())})
		}
	}
	return
}

//MID exported
func MID(k string, v int64) []lib.Pair {
	return []lib.Pair{{A: k, B: strconv.FormatInt(v, 10)}}
}

//TAG exported
func TAG(m Model, tag string) (out map[string]string) {

	var (
		v = reflect.ValueOf(m.Val())
		t = reflect.Indirect(v)
	)

	out = make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		if k, ok := t.Type().Field(i).Tag.Lookup(tag); ok {
			out[t.Type().Field(i).Name] = k
		}
	}
	return
}
