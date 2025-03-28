package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type TagCreateValidator struct {
	Tag struct {
		Name string `from:"tag_name" json:"tag_name" binding:"required,startswith=#,min=2,max=26"`
	} `json:"tag_update"`
	tModel models.TagModel `json:"-"`
}

func NewTagCreateValidator() TagCreateValidator {
	return TagCreateValidator{}
}

func (tcv *TagCreateValidator) Model() models.TagModel {
	return tcv.tModel
}

func (tcv *TagCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, tcv); err != nil {
		return err
	}
	tcv.tModel.Name = tcv.Tag.Name
	tcv.tModel.AutorID = c.MustGet(flag.KeyUserID).(uint)
	tcv.tModel.CreatedAt = time.Now()
	return nil
}
