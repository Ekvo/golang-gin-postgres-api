package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type ArticlePropertyValidator struct {
	Property struct {
		Tags      []string  `from:"tags" json:"tags"`
		AutorName string    `form:"autor_name" json:"autor_name" binding:"omitempty,alphanum,min=1,max=128"'`
		Favorited bool      `form:"favorited" json:"favorited"`
		StartDate time.Time `from:"start_date" json:"start_date"`
		EndDate   time.Time `form:"end_date" json:"end_date"`
		Limit     uint      `form:"limit" json:"limit" binding:"omitempty,numeric"`
		Offset    uint      `form:"offset" json:"offset" binding:"omitempty,numeric"`
	} `json:"artcicle_property"`
	aProperty models.ArticleProperty `json:"-"`
}

func NewArticlePropertyValidator() ArticlePropertyValidator {
	return ArticlePropertyValidator{}
}

func (apv *ArticlePropertyValidator) Model() models.ArticleProperty {
	return apv.aProperty
}

func (apv *ArticlePropertyValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, apv); err != nil {
		return err
	}
	apv.aProperty.Tags = apv.Property.Tags
	apv.aProperty.AutorName = apv.Property.AutorName
	apv.aProperty.Favorited = apv.Property.Favorited
	if apv.Property.Limit == 0 {
		apv.Property.Limit = 20
	}
	apv.aProperty.Limit = apv.Property.Limit
	apv.aProperty.Offset = apv.Property.Offset

	if start := apv.Property.StartDate; !start.IsZero() {
		apv.aProperty.StartDate = start
	}
	if end := apv.Property.EndDate; !end.IsZero() {
		apv.aProperty.EndDate = end
	} else {
		apv.aProperty.EndDate = time.Now()
	}
	return nil
}
