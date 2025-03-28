package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type CommentPropertyValidator struct {
	Property struct {
		ArcticleSlug string    `form:"arcticle_slug" json:"arcticle_slug" binding:"required,aphanum,min=1,max=50"`
		AutorName    string    `form:"autor_name" json:"autor_name" binding:"omitempty,alphanum,min=1,max=128"`
		StartDate    time.Time `from:"start_date" json:"start_date"`
		EndDate      time.Time `form:"end_date" json:"end_date"`
		Limit        uint      `form:"limit" json:"limit" binding:"omitempty,numeric"`
		Offset       uint      `form:"offset" json:"offset" binding:"omitempty,numeric"`
	} `json:"comment_property"`
	cProperty models.CommentProperty `json:"-"`
}

func NewCommentPropertyValidator() CommentPropertyValidator {
	return CommentPropertyValidator{}
}

func (cpv *CommentPropertyValidator) Model() models.CommentProperty {
	return cpv.cProperty
}

func (cpv *CommentPropertyValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, cpv); err != nil {
		return err
	}
	cpv.cProperty.ArcticleSlug = cpv.Property.ArcticleSlug
	cpv.cProperty.AutorName = cpv.Property.AutorName
	if cpv.Property.Limit == 0 {
		cpv.Property.Limit = 20
	}
	cpv.cProperty.Limit = cpv.Property.Limit
	cpv.cProperty.Offset = cpv.Property.Offset
	if start := cpv.Property.StartDate; !start.IsZero() {
		cpv.cProperty.StartDate = start
	}
	if end := cpv.Property.EndDate; !end.IsZero() {
		cpv.cProperty.EndDate = end
	} else {
		cpv.cProperty.EndDate = time.Now()
	}
	return nil
}
