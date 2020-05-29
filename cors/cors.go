package cors

import (
	"reflect"

	"github.com/huandu/go-sqlbuilder"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/msg"
)

var cors map[string]string //map[app_key]app_origin

//Init exported
func Init(m db.Model) (err error) {

	var (
		f [2]string
		t = reflect.Indirect(reflect.ValueOf(m))

		key    string
		origin string

		sb = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	for i := 0; i < t.NumField(); i++ {
		if tag, ok := t.Type().Field(i).Tag.Lookup("cors"); ok {
			if col, ok := t.Type().Field(i).Tag.Lookup("db"); ok {
				switch tag {
				case "key":
					f[0] = col
				case "origin":
					f[1] = col
				}
			}
		}
	}

	for _, col := range f {
		if col == "" {
			//CORS tags are not properly set
			return msg.Get("29").M2E()
		}
	}

	sb.From(m.View())
	sb.Select(f[0], f[1])
	//sb.Where(
	//	sb.LessThan(f[3], now),
	//	sb.GreaterThan(f[4], now),
	//)

	sql, args := sb.Build()
	//log.Println(sql, args)
	rows, err := db.Db().Query(sql, args...)
	defer rows.Close()

	//scan rows
	cors = make(map[string]string)
	for rows.Next() {
		err := rows.Scan(&key, &origin)
		if err != nil {
			//Server error: %s
			return msg.Get("25").SetArgs(err).M2E()
		}
		cors[key] = origin
	}
	err = rows.Err()
	if err != nil {
		//Server error: %s
		return msg.Get("25").SetArgs(err).M2E()
	}

	//log.Println(cors)
	return nil
}

//CORS exported
func CORS() map[string]string {

	return cors
}
