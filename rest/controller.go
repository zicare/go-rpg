package rest

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/validation"
)

//Controller exported
type Controller struct{}

//Get exported
func (ctrl Controller) Get(c *gin.Context, m db.Model) {

	var (
		pIDs = ParamIDs(c, m)
		err  = db.Find(m, pIDs)
	)

	if err == sql.ErrNoRows {
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "Not found!"},
		)
	} else if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": err.Error()},
		)
	} else {
		c.JSON(http.StatusOK, m.Xfrm())
	}
}

//Index exported
func (ctrl Controller) Index(c *gin.Context, m db.Model) {

	var (
		opt             = params(c, m)
		meta, data, err = db.FetchAll(m, opt)
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": err.Error()},
		)
	} else if len(data) <= 0 {
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "No found!"},
		)
	} else {
		c.Header("X-Range", meta.Range)
		c.Header("X-Checksum", meta.Checksum)
		c.JSON(http.StatusOK, func() []interface{} {
			for k, v := range data {
				data[k] = v
			}
			return data
		}())
	}
}

//IndexHead exported
func (ctrl Controller) IndexHead(c *gin.Context, m db.Model) {

	var (
		opt             = params(c, m)
		meta, data, err = db.FetchAll(m, opt)
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": err.Error()},
		)
	} else if len(data) <= 0 {
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "No found!"},
		)
	} else {
		c.Header("X-Range", meta.Range)
		c.Header("X-Checksum", meta.Checksum)
		c.JSON(http.StatusOK, gin.H{})
	}

}

//POST exported
func (ctrl Controller) POST(c *gin.Context, m db.Model) {

	if err := c.ShouldBind(m); err != nil {
		/*
		 * payload isn't correct
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"errors": validation.GetMessages(err)},
		)
	} else if err := m.FilterInput([]lib.Pair{}); err != nil {
		/*
		 * payload isn't correct
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"errors": err},
		)
	} else if err := db.Insert(m); err != nil {
		/*
		 * maybe something went wrong connecting to the db
		 * or some constrain was not verified and violated, etc
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": err.Error()},
		)
	} else {
		/*
		 * everything went okay
		 */
		c.JSON(http.StatusCreated, m.Xfrm())
	}
}

//PUT exported
func (ctrl Controller) PUT(c *gin.Context, m db.Model) {

	pIDs := ParamIDs(c, m)

	if err := db.Find(m.New(), pIDs); err != nil {
		/*
		 * requested resource doesn't exist
		 */
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "Not found!"},
		)
	} else if err := c.ShouldBind(m); err != nil {
		/*
		 * payload isn't correct
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"errors": validation.GetMessages(err)},
		)
	} else if err := m.FilterInput(pIDs); err != nil {
		/*
		 * payload isn't correct
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"errors": err},
		)
	} else if err := db.Update(m, pIDs); err == sql.ErrNoRows {
		/*
		 * this shouldn't happen as suppousedly the resource exists
		 * and user input curated, but you never know
		 * maybe someone else deleted the resource
		 * since it was previously inspected
		 */
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "Not found!"},
		)
	} else if err != nil {
		/*
		 * maybe something went wrong connecting to the db
		 * or some constrain was not verified and violated, etc
		 */
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": err.Error()},
		)
	} else {
		/*
		 * everything went okay
		 */
		c.JSON(http.StatusCreated, m.Xfrm())
	}
}
