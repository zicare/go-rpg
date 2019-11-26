package rest

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/validation"
	"gopkg.in/go-playground/validator.v8"
)

//Controller exported
type Controller struct {
	err error
}

//Error exported
func (ctrl Controller) Error() error {
	return ctrl.err
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

//Get exported
func (ctrl Controller) Get(c *gin.Context, m db.Model) {

	if pIDs, err := ParamIDs(c, m); err != nil {
		/*
		 * composite key missuse
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": err.Error()},
		)
	} else if err = db.Find(m, pIDs); err == sql.ErrNoRows {
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "Not found!"},
		)
	} else if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": err.Error()},
		)
	} else if sm, ok := m.(db.ScopedModel); ok == true && sm.ScopeOk(c) == false {
		/*
		 * requested resource is scoped and it doesn't belong
		 * maybe track unscoped requests and take action on abuse?
		 */
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "Not found!"},
		)
	} else {
		c.JSON(http.StatusOK, m.Xfrm())
	}
}

//Post exported
func (ctrl *Controller) Post(c *gin.Context, m db.Model) {

	if ctrl.err = m.Bind(c, []lib.Pair{}); ctrl.err != nil {
		/*
		 * payload isn't correct
		 */
		switch ctrl.err.(type) {
		case validator.ValidationErrors:
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": "There are validation errors",
					"errors": validation.GetMessages(ctrl.err, m)},
			)
		default:
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": ctrl.err.Error()},
			)
		}
	} else if ctrl.err = db.Insert(m); ctrl.err != nil {
		/*
		 * maybe something went wrong connecting to the db
		 * or some constrain was not verified and violated, etc
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": ctrl.err.Error()},
		)
	} else {
		/*
		 * everything went okay
		 */
		c.JSON(http.StatusCreated, m.Xfrm())
	}
}

//Put exported
func (ctrl Controller) Put(c *gin.Context, m db.Model) {

	var pIDs []lib.Pair
	var aux = m.New()

	if pIDs, ctrl.err = ParamIDs(c, m); ctrl.err != nil {
		/*
		 * pIDs count didn't match table's primary key count
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": ctrl.err.Error()},
		)
	} else if ctrl.err = db.Find(aux, pIDs); ctrl.err != nil {
		/*
		 * requested resource doesn't exist
		 */
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "Not found!"},
		)
	} else if sm, ok := aux.(db.ScopedModel); ok == true && sm.ScopeOk(c) == false {
		/*
		 * requested resource is scoped and it doesn't belong
		 * maybe track unscoped requests and take action on abuse?
		 */
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": "Not found!"},
		)
	} else if ctrl.err = m.Bind(c, pIDs); ctrl.err != nil {
		/*
		 * payload isn't correct
		 */
		switch ctrl.err.(type) {
		case validator.ValidationErrors:
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": "There are validation errors",
					"errors": validation.GetMessages(ctrl.err, m)},
			)
		default:
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": ctrl.err.Error()},
			)
		}
	} else if ctrl.err = db.Update(m, pIDs); ctrl.err == sql.ErrNoRows {
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
	} else if ctrl.err != nil {
		/*
		 * maybe something went wrong connecting to the db
		 * or some constrain was not verified and violated, etc
		 */
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": ctrl.err.Error()},
		)
	} else {
		/*
		 * everything went okay
		 */
		c.JSON(http.StatusCreated, m.Xfrm())
	}
}

//Delete exported
func (ctrl Controller) Delete(c *gin.Context, m db.Model) {

	var pIDs []lib.Pair

	if pIDs, ctrl.err = ParamIDs(c, m); ctrl.err != nil {
		/*
		 * composite key missuse
		 */
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": ctrl.err.Error()},
		)
	} else if ctrl.err = m.Delete(c, pIDs); ctrl.err != nil {
		switch e := ctrl.err.(type) {
		case *db.NotAllowedError:
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": e.Error()},
			)
		case *db.NotFoundError:
			c.JSON(
				http.StatusNotFound,
				gin.H{"message": e.Error()},
			)
		default:
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"message": ctrl.err.Error()},
			)
		}
	} else {
		c.JSON(
			http.StatusOK,
			gin.H{"message": "Deleted"},
		)
	}
}
