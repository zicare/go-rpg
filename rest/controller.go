package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zicare/go-rpg/db"
	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/msg"
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

	if meta, data, err := db.FetchAll(c, m); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": err},
		)
	} else if len(data) <= 0 {
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": msg.Get("18")}, //Not found!
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

	if meta, data, err := db.FetchAll(c, m); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": err},
		)
	} else if len(data) <= 0 {
		c.JSON(
			http.StatusNotFound,
			gin.H{"message": msg.Get("18")}, //Not found!
		)
	} else {
		c.Header("X-Range", meta.Range)
		c.Header("X-Checksum", meta.Checksum)
		c.JSON(http.StatusOK, gin.H{})
	}

}

//Get exported
func (ctrl Controller) Get(c *gin.Context, m db.Model) {

	if err := db.Find(c, m); err != nil {
		switch e := err.(type) {
		case *db.NotFoundError:
			c.JSON(
				http.StatusNotFound,
				gin.H{"message": e}, //Not found!
			)
		case *db.ParamError:
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": e},
			)
		default:
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"message": e},
			)
		}
	} else {
		c.JSON(http.StatusOK, m.Xfrm())
	}
}

//Post exported
func (ctrl *Controller) Post(c *gin.Context, m db.Model) {

	if ctrl.err = db.Insert(c, m); ctrl.err != nil {
		switch ctrl.err.(type) {
		case *db.NotFoundError:
			//Resource created but out of the read scope
			//so response is 204
			c.AbortWithStatus(http.StatusNoContent)
		case validator.ValidationErrors, *time.ParseError, *json.UnmarshalTypeError:
			//Resource not created
			//payload isn't correct
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": msg.Get("19"), //There are validation errors
					"errors":  validation.GetMessages(ctrl.err, m),
				},
			)
		default:
			//Resource not created
			//something went wrong but we don't know what
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"message": ctrl.err},
			)
		}
	} else {
		c.JSON(http.StatusCreated, m.Xfrm())
	}
}

//Put exported
func (ctrl Controller) Put(c *gin.Context, m db.Model) {

	if ctrl.err = db.Update(c, m); ctrl.err != nil {
		switch e := ctrl.err.(type) {
		case *db.ParamError:
			//composite key missuse
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": e},
			)
		case *db.NotFoundError:
			//not found or out of scope
			c.JSON(
				http.StatusNotFound,
				gin.H{"message": e},
			)
		case validator.ValidationErrors, *time.ParseError, *json.UnmarshalTypeError:
			//payload issues
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": msg.Get("19"), //There are validation errors
					"errors":  validation.GetMessages(e, m),
				},
			)
		default:
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"message": e},
			)
		}
	} else {
		c.JSON(http.StatusOK, m.Xfrm())
	}
}

//Delete exported
func (ctrl Controller) Delete(c *gin.Context, m db.Model) {

	var pIDs []lib.Pair

	if pIDs, ctrl.err = db.ParamIDs(c, m); ctrl.err != nil {
		//composite key missuse
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": ctrl.err},
		)
	} else if ctrl.err = m.Delete(c, pIDs); ctrl.err != nil {
		switch e := ctrl.err.(type) {
		case *db.NotAllowedError:
			c.JSON(
				http.StatusBadRequest,
				gin.H{"message": e},
			)
		case *db.NotFoundError:
			c.JSON(
				http.StatusNotFound,
				gin.H{"message": e},
			)
		case *db.ConflictError:
			c.JSON(
				http.StatusConflict,
				gin.H{"message": e},
			)
		default:
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"message": e},
			)
		}
	} else {
		//deleted
		c.AbortWithStatus(http.StatusNoContent)
	}
}
