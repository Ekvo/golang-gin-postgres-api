package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type TagPropertyValidator struct {
	Property struct {
		AutorName string    `form:"autor_name" json:"autor_name" binding:"omitempty,alphanum,min=1,max=128"`
		StartDate time.Time `from:"start_date" json:"start_date"`
		EndDate   time.Time `form:"end_date" json:"end_date"`
		Limit     uint      `form:"limit" json:"limit" binding:"omitempty,numeric"`
		Offset    uint      `form:"offset" json:"offset" binding:"omitempty,numeric"`
	} `json:"tag_property"`
	tProperty models.TagPropery `json:"-"`
}

func NewTagPropertyValidator() TagPropertyValidator {
	return TagPropertyValidator{}
}

func (tpv *TagPropertyValidator) Model() models.TagPropery {
	return tpv.tProperty
}

func (tpv *TagPropertyValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, tpv); err != nil {
		return err
	}
	tpv.tProperty.AutorName = tpv.Property.AutorName
	if tpv.Property.Limit == 0 {
		tpv.Property.Limit = 20
	}
	tpv.tProperty.Limit = tpv.Property.Limit
	tpv.tProperty.Offset = tpv.Property.Offset
	if start := tpv.Property.StartDate; !start.IsZero() {
		tpv.tProperty.StartDate = start
	}
	if end := tpv.Property.EndDate; !end.IsZero() {
		tpv.tProperty.EndDate = end
	} else {
		tpv.tProperty.EndDate = time.Now()
	}
	return nil
}
