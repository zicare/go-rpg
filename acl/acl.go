package acl

import (
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
func Init() (err error) {

	var (
		g Grant

		role   int64
		route  string
		method string
		from   time.Time
		to     time.Time

		now = time.Now()
		sb  = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	sb.From("view_lnk_routes_roles")
	sb.Select("role_id", "route", "method", "system_access_from", "system_access_to")
	sb.Where(
		sb.LessThan("system_access_from", now),
		sb.GreaterThan("system_access_to", now),
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
