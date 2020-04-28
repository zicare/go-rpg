package db

import (
	"github.com/gin-gonic/gin"
	"github.com/huandu/go-sqlbuilder"

	"github.com/zicare/go-rpg/lib"
	"github.com/zicare/go-rpg/msg"
	"gopkg.in/go-playground/validator.v8"
)

//ReadOnlyModel exported
type ReadOnlyModel struct{}

//Table exported
func (ReadOnlyModel) Table() string {
	return ""
}

//Bind exported
func (ReadOnlyModel) Bind(c *gin.Context, pIDs []lib.Pair) error {
	//Read only model
	return msg.Get("16").M2E()
}

//Validation exported
func (ReadOnlyModel) Validation(v *validator.Validate, sl *validator.StructLevel) {}

//Delete exported
func (ReadOnlyModel) Delete(c *gin.Context, pIDs []lib.Pair) error {
	//Read only model
	return msg.Get("16").M2E()
}

//Scope exported
func (ReadOnlyModel) Scope(b sqlbuilder.Builder, c *gin.Context) {
	return
}
