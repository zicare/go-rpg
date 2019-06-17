package db

import (
	"database/sql"
	"fmt"
	"reflect"

	//required
	_ "github.com/lib/pq"
	"github.com/zicare/go-rpg/config"
	"github.com/zicare/go-rpg/lib"
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
}

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

	return
}

//Db returns the db handler
func Db() *sql.DB {

	return db
}
