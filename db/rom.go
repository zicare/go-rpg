package db

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/huandu/go-sqlbuilder"

	"github.com/zicare/go-rpg/lib"
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
	return errors.New("Read only model")
}

//Validation exported
func (ReadOnlyModel) Validation(v *validator.Validate, sl *validator.StructLevel) {}

//Delete exported
func (ReadOnlyModel) Delete(c *gin.Context, pIDs []lib.Pair) error {
	return errors.New("Read only model")
}

//Scope exported
func (ReadOnlyModel) Scope(b sqlbuilder.Builder, c *gin.Context) {
	return
}
