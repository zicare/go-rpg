package acl

import (
	"errors"
	"reflect"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/lib"
)

var acl map[Grant]lib.TimeRange

//Grant exported
type Grant struct {
	RoleID int64
	Route  string
	Method string
}

//Init exported
func Init(m db.Model) (err error) {

	var (
		f [5]string
		t = reflect.Indirect(reflect.ValueOf(m))

		g Grant

		role   int64
		route  string
		method string
		from   time.Time
		to     time.Time

		now = time.Now()
		sb  = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	for i := 0; i < t.NumField(); i++ {
		if tag, ok := t.Type().Field(i).Tag.Lookup("acl"); ok {
			if col, ok := t.Type().Field(i).Tag.Lookup("db"); ok {
				switch tag {
				case "role":
					f[0] = col
				case "route":
					f[1] = col
				case "method":
					f[2] = col
				case "from":
					f[3] = col
				case "to":
					f[4] = col
				}
			}
		}
	}

	for _, col := range f {
		if col == "" {
			return errors.New(("ACL tags are not properly set"))
		}
	}

	sb.From(m.Table())
	sb.Select(f[0], f[1], f[2], f[3], f[4])
	sb.Where(
		sb.LessThan(f[3], now),
		sb.GreaterThan(f[4], now),
	)

	sql, args := sb.Build()
	//log.Println(sql, args)
	rows, err := db.Db().Query(sql, args...)
	defer rows.Close()

	//scan rows
	acl = make(map[Grant]lib.TimeRange)
	for rows.Next() {
		err := rows.Scan(&role, &route, &method, &from, &to)
		if err != nil {
			return err
		}
		g = Grant{RoleID: role, Route: route, Method: method}
		acl[g] = lib.TimeRange{From: from, To: to}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	//log.Println(acl)
	return nil
}

//ACL exported
func ACL() map[Grant]lib.TimeRange {

	return acl
}