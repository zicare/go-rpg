package db

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"reflect"
	"strconv"
	"strings"

	"database/sql"

	"github.com/gin-gonic/gin"

	"github.com/huandu/go-sqlbuilder"
	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/slice"
)

//FetchAll exported
func FetchAll(c *gin.Context, m Model) (ResultSetMeta, []interface{}, error) {

	var (
		opt     = params(c, m)
		meta    = ResultSetMeta{Range: "*/*", Checksum: "*"}
		total   string
		results []interface{}
		table   = m.View()
		ms      = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
		sb      = ms.SelectFrom(table)

		//where
		fnFst = func(v string) string { p := strings.Split(v, ","); return p[0] }
		cond  = map[string]func(string, string) string{
			"eq": func(k string, v string) string {
				p := strings.Split(v, ",")
				if len(p) > 1 {
					return sb.In(k, sqlbuilder.Flatten(p)...)
				}
				return sb.Equal(k, v)
			},
			"gt":   func(k string, v string) string { return sb.GreaterThan(k, fnFst(v)) },
			"gteq": func(k string, v string) string { return sb.GreaterEqualThan(k, fnFst(v)) },
			"st":   func(k string, v string) string { return sb.LessThan(k, fnFst(v)) },
			"steq": func(k string, v string) string { return sb.LessEqualThan(k, fnFst(v)) },
		}
	)

	//set where scope
	m.Scope(sb, c)

	//set where
	for i, j := range cond {
		op, ok := opt.Filter[i]
		if ok {
			for _, v := range op {
				sb.Where(j(v.A.(string), v.B.(string)))
			}
		}
	}

	//set where null
	for _, j := range opt.Null {
		sb.Where(sb.IsNull(j))
	}

	//set where not null
	for _, j := range opt.NotNull {
		sb.Where(sb.IsNotNull(j))
	}

	//get total count
	sb.Select(sb.As("COUNT(*)", "t"))
	sql, args := sb.Build()
	//fmt.Println(sql, args)
	err := db.QueryRow(sql, args...).Scan(&total)
	if err != nil {
		return meta, results, err
	}

	//total = 0 ? no need to continue
	if total == "0" {
		return meta, results, nil
	}

	//set columns, order by, limit and offset
	sb.Select(opt.Column...)
	sb.OrderBy(opt.Order...)
	sb.Limit(opt.Limit)
	sb.Offset(opt.Offset)

	//buils sql and execute it
	sql, args = sb.Build()
	//fmt.Println(sql, args)
	rows, err := db.Query(sql, args...)
	defer rows.Close()

	//scan rows
	for rows.Next() {
		err := rows.Scan(ms.AddrWithCols(opt.Column, &m)...)
		if err != nil {
			return meta, results, err
		}
		results = append(results, m.Xfrm().Val())
	}
	err = rows.Err()
	if err != nil {
		return meta, results, err
	}

	//meta
	from := strconv.Itoa(opt.Offset)
	to := strconv.Itoa(lib.Max(opt.Offset, len(results)-opt.Offset-1))
	meta.Range = fmt.Sprintf("%s-%s/%s", from, to, total)
	if opt.Checksum == 1 {
		bytes, _ := json.Marshal(results)
		checksum := crc32.ChecksumIEEE([]byte(bytes))
		meta.Checksum = strconv.FormatUint(uint64(checksum), 16)
	}

	return meta, results, nil
}

//Find exported
func Find(c *gin.Context, m Model) error {

	var (
		err error
		id  []lib.Pair
	)

	if id, err = ParamIDs(c, m); err != nil {
		return err
	}

	return find(c, m, id, true)
}

//Insert exported
func Insert(c *gin.Context, m Model) error {

	if err := m.Bind(c, []lib.Pair{}); err != nil {
		return err
	}

	var (
		table       = m.Table()
		fields, val = Fields(m)
		ms          = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
		ib          = sqlbuilder.PostgreSQL.NewInsertBuilder()
	)

	var v []interface{}
	for _, w := range fields.Ordered {
		if slice.Contains(fields.Primary, w) && slice.Contains(fields.Serial, w) {
			v = append(v, sqlbuilder.Raw("DEFAULT"))
		} else if !slice.Contains(fields.View, w) {
			v = append(v, val[w])
		}
	}

	ib.InsertInto(table)
	ib.Values(v...)

	sql, args := ib.Build()
	//fmt.Println(sql, args)
	if err := db.QueryRow(sql+" RETURNING *", args...).
		Scan(ms.AddrWithCols(fields.Writable, &m)...); err != nil {
		return err
	}

	return find(c, m, PID(m, fields.Primary), false)
}

//Update exported
func Update(c *gin.Context, m Model) error {

	var (
		err   error
		id    []lib.Pair
		table = m.Table()
		meta  Meta
		val   map[string]interface{}
		ub    = sqlbuilder.PostgreSQL.NewUpdateBuilder()
	)

	if id, err = ParamIDs(c, m); err != nil {
		//composite key misuse
		return err
	} else if err = m.Bind(c, id); err != nil {
		//payload problem
		return err
	}

	meta, val = Fields(m)
	ub.Update(table)

	m.Scope(ub, c)

	for _, p := range id {
		ub.Where(ub.Equal(p.A.(string), p.B.(string)))
		_, ok := val[p.A.(string)]
		if ok {
			delete(val, p.A.(string))
		}
	}

	var asg []string
	for k, v := range val {
		if !reflect.ValueOf(v).IsNil() && slice.Contains(meta.Writable, k) {
			asg = append(asg, ub.Assign(k, v))
		}
	}
	ub.Set(asg...)

	sql, args := ub.Build()
	//fmt.Println(sql, args)
	if res, err := db.Exec(sql, args...); err != nil {
		return err
	} else if rows, _ := res.RowsAffected(); rows == 0 {
		e := new(NotFoundError)
		e.MSG = "Not found"
		return e
	}

	return find(c, m, id, false)
}

//ByID exported
func ByID(c *gin.Context, m Model, id []lib.Pair) error {
	return find(c, m, id, true)
}

func find(c *gin.Context, m Model, id []lib.Pair, scope bool) error {

	var (
		table = m.View()
		ms    = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
		sb    = ms.SelectFrom(table)
	)

	if scope {
		m.Scope(sb, c)
	}

	for _, p := range id {
		sb.Where(sb.Equal(p.A.(string), p.B.(string)))
	}

	q, args := sb.Build()
	//log.Println(q, args)
	if err := db.QueryRow(q, args...).Scan(ms.Addr(&m)...); err == sql.ErrNoRows {
		e := new(NotFoundError)
		e.MSG = "Not found"
		return e
	} else if err != nil {
		return err
	}
	return nil
}
