package rest

import (
	"strconv"
	"strings"

	"github.com/zicare/go-rpg/config"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/lib"

	"github.com/gin-gonic/gin"
)

//ParamIDs exported
func ParamIDs(c *gin.Context, m db.Model) (mIDs []lib.Pair) {

	var (
		pkParam      = strings.Split(c.Param("id"), ",")
		_, _, pkCols = db.Cols(m)
	)

	//mIDs
	mIDs = []lib.Pair{}
	for i, k := range pkCols {
		if len(pkParam) > i {
			mIDs = append(mIDs, lib.Pair{A: k, B: pkParam[i]})
		}
	}
	return
}

func params(c *gin.Context, m db.Model) (opts db.SelectOpt) {

	var (
		cf                   = [5]string{"eq", "gt", "st", "gteq", "steq"}
		cols, colsOrdered, _ = db.Cols(m)
		param                = c.Request.URL.Query()
	)

	//filter
	opts.Filter = make(map[string][]lib.Pair)
	for _, fi := range cf {
		opts.Filter[fi] = []lib.Pair{}
		i, ok := param[fi]
		if ok {
			j := strings.Split(i[0], ";")
			for _, k := range j {
				j := strings.Split(k, "|")
				_, ok := cols[j[0]]
				if ok {
					opts.Filter[fi] = append(opts.Filter[fi], lib.Pair{A: j[0], B: j[1]})
				}
			}
		}
	}

	//null
	opts.Null = []string{}
	if i, ok := param["null"]; ok {
		colsAux := make(map[string]string)
		for _, v := range strings.Split(i[0], ",") {
			colsAux[v] = v
		}
		for _, k := range colsOrdered {
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
		for _, k := range colsOrdered {
			if _, ok := colsAux[k]; ok {
				opts.NotNull = append(opts.NotNull, k)
			}
		}
	}

	//column
	opts.Column = []string{}
	i, ok := param["cols"]
	if ok {
		colsAux := make(map[string]string)
		j := strings.Split(i[0], ",")
		for _, v := range j {
			colsAux[v] = v
		}
		for _, k := range colsOrdered {
			_, ok := colsAux[k]
			if ok {
				opts.Column = append(opts.Column, k)
			}
		}
	} else {
		opts.Column = colsOrdered
	}

	//order
	opts.Order = []string{}
	i, ok = param["order"]
	if ok {
		j := strings.Split(i[0], ";")
		for _, k := range j {
			j := strings.Split(k, "|")
			_, ok := cols[j[0]]
			if !ok {
				continue
			}
			if len(j) == 1 {
				opts.Order = append(opts.Order, j[0]+" ASC")
			} else if strings.ToUpper(j[1]) == "ASC" || strings.ToUpper(j[1]) == "DESC" {
				opts.Order = append(opts.Order, j[0]+" "+strings.ToUpper(j[1]))
			}
		}
	}

	//offset and limit
	opts.Offset = 0
	opts.Limit, _ = strconv.Atoi(config.Config().GetString("param.icpp"))
	i, ok = param["limit"]
	if ok {
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
