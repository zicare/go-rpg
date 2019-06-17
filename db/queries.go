package db

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/slice"
)

//FetchAll exported
func FetchAll(m Model, opt SelectOpt) (ResultSetMeta, []interface{}, error) {

	var (
		meta        = ResultSetMeta{Range: "*/*", Checksum: "*"}
		total       string
		results     []interface{}
		table       = m.Table()
		modelStruct = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
		sb          = modelStruct.SelectFrom(table)

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
	log.Println(sql, args)
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
	log.Println(sql, args)
	rows, err := db.Query(sql, args...)
	defer rows.Close()

	//scan rows
	for rows.Next() {
		err := rows.Scan(modelStruct.AddrWithCols(opt.Column, &m)...)
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
func Find(m Model, id []lib.Pair) error {

	var (
		table       = m.Table()
		modelStruct = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
		sb          = modelStruct.SelectFrom(table)
	)

	for _, p := range id {
		sb.Where(sb.Equal(p.A.(string), p.B.(string)))
	}

	sql, args := sb.Build() //;log.Println(sql, args)
	return db.QueryRow(sql, args...).Scan(modelStruct.Addr(&m)...)
}

//Insert exported
func Insert(m Model) error {

	var (
		table                 = m.Table()
		cols, orderedCols, pk = Cols(m)
		ms                    = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
		ib                    = sqlbuilder.PostgreSQL.NewInsertBuilder()
	)

	var v []interface{}
	for _, c := range orderedCols {
		if slice.Contains(pk, c) {
			v = append(v, sqlbuilder.Raw("DEFAULT"))
		} else {
			v = append(v, cols[c])
		}
	}

	ib.InsertInto(table)
	ib.Values(v...)
	sql, args := ib.Build() //;log.Println(sql, args)
	return db.QueryRow(sql+" RETURNING *", args...).Scan(ms.Addr(&m)...)
}

//Update exported
func Update(m Model, id []lib.Pair) error {

	var (
		table      = m.Table()
		cols, _, _ = Cols(m)
		ms         = sqlbuilder.NewStruct(m).For(sqlbuilder.PostgreSQL)
		ub         = sqlbuilder.PostgreSQL.NewUpdateBuilder()
	)

	ub.Update(table)

	for _, p := range id {
		ub.Where(ub.Equal(p.A.(string), p.B.(string)))
		_, ok := cols[p.A.(string)]
		if ok {
			delete(cols, p.A.(string))
		}
	}

	var asg []string
	for k, v := range cols {
		if !reflect.ValueOf(v).IsNil() {
			asg = append(asg, ub.Assign(k, v))
		}
	}
	ub.Set(asg...)

	sql, args := ub.Build() //;log.Println(sql, args)
	return db.QueryRow(sql+" RETURNING *", args...).Scan(ms.Addr(&m)...)
}
