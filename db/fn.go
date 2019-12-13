package db

import (
	"strconv"
	"strings"

	"github.com/zicare/go-rpg/slice"

	"github.com/zicare/go-rpg/config"
	"github.com/zicare/go-rpg/lib"

	"github.com/gin-gonic/gin"
)

//ParamIDs exported
func ParamIDs(c *gin.Context, m Model) ([]lib.Pair, error) {

	var (
		mIDs      = []lib.Pair{}
		pkParam   = strings.Split(c.Param("id"), ",")
		fields, _ = Fields(m)
	)

	if len(pkParam) != len(fields.Primary) {
		e := new(ParamError)
		e.MSG = "Composite key missuse"
		return mIDs, e
	}

	//mIDs
	for i, k := range fields.Primary {
		mIDs = append(mIDs, lib.Pair{A: k, B: pkParam[i]})
	}
	return mIDs, nil
}

func params(c *gin.Context, m Model) (opts SelectOpt) {

	var (
		cf          = [5]string{"eq", "gt", "st", "gteq", "steq"}
		fields, val = Fields(m)
		param       = c.Request.URL.Query()
	)

	//filter
	opts.Filter = make(map[string][]lib.Pair)
	for _, fi := range cf {
		opts.Filter[fi] = []lib.Pair{}
		if i, ok := param[fi]; ok {
			//j := strings.Split(i[0], ";")
			for _, k := range i {
				j := strings.Split(k, "|")
				_, ok := val[j[0]]
				if ok {
					opts.Filter[fi] = append(opts.Filter[fi], lib.Pair{A: j[0], B: j[1]})
				}
			}
		}
	}

	/*
		//scope filter
		if sm, ok := m.(ScopedModel); ok == true {
			for _, sf := range sm.GetScope(c) {
				opts.Filter["eq"] = append(opts.Filter["eq"], sf)
			}
		}
	*/

	//null
	opts.Null = []string{}
	if i, ok := param["null"]; ok {
		colsAux := make(map[string]string)
		for _, v := range strings.Split(i[0], ",") {
			colsAux[v] = v
		}
		for _, k := range fields.Ordered {
			if _, ok := colsAux[k]; ok {
				opts.Null = append(opts.Null, k)
			}
		}
	}

	//not null
	opts.NotNull = []string{}
	if i, ok := param["notnull"]; ok {
		colsAux := make(map[string]string)
		for _, v := range strings.Split(i[0], ",") {
			colsAux[v] = v
		}
		for _, k := range fields.Ordered {
			if _, ok := colsAux[k]; ok {
				opts.NotNull = append(opts.NotNull, k)
			}
		}
	}

	//column
	opts.Column = []string{}
	if i, ok := param["cols"]; ok {
		colsAux := make(map[string]string)
		j := strings.Split(i[0], ",")
		for _, v := range j {
			colsAux[v] = v
		}
		for _, k := range fields.Ordered {
			if _, ok := colsAux[k]; ok {
				opts.Column = append(opts.Column, k)
			}
		}
	} else {
		opts.Column = fields.Ordered
	}

	//xcols
	if i, ok := param["xcols"]; ok {
		opts.Column = slice.Diff(opts.Column, strings.Split(i[0], ","))
	}

	//order
	opts.Order = []string{}
	if i, ok := param["order"]; ok {
		j := strings.Split(i[0], ";")
		for _, k := range j {
			j := strings.Split(k, "|")
			if _, ok := val[j[0]]; !ok {
				continue
			} else if len(j) == 1 {
				opts.Order = append(opts.Order, j[0]+" ASC")
			} else if strings.ToUpper(j[1]) == "ASC" || strings.ToUpper(j[1]) == "DESC" {
				opts.Order = append(opts.Order, j[0]+" "+strings.ToUpper(j[1]))
			}
		}
	}

	//offset and limit
	opts.Offset = 0
	opts.Limit, _ = strconv.Atoi(config.Config().GetString("param.icpp"))
	if i, ok := param["limit"]; ok {
		j := strings.Split(i[0], ",")
		switch len(j) {
		case 1:
			opts.Offset = 0
			opts.Limit, _ = strconv.Atoi(j[0])
		case 2:
			opts.Offset, _ = strconv.Atoi(j[0])
			opts.Limit, _ = strconv.Atoi(j[1])
		}
	}

	//checksum
	if c.Query("checksum") == "1" {
		opts.Checksum = 1
	}

	return
}
