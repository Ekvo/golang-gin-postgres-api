package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type UserProperyValidator struct {
	Property struct {
		FirstName string    `form:"first_name" json:"first_name" binding:"omitempty,alpha,min=1,max=128"`
		LastName  string    `form:"last_name" json:"last_name" binding:"omitempty,alpha,min=1,max=128"`
		StartDate time.Time `from:"start_date" json:"start_date"`
		EndDate   time.Time `form:"end_date" json:"end_date"`
		Limit     uint      `form:"limit" json:"limit" binding:"omitempty,numeric"`
		Offset    uint      `form:"offset" json:"offset" binding:"omitempty,numeric"`
	} `json:"user_property"`
	uProperty models.UserProperty `json:"-"`
}

func NewUserProperyValidator() UserProperyValidator {
	return UserProperyValidator{}
}

func (upv *UserProperyValidator) Model() models.UserProperty {
	return upv.uProperty
}

func (upv *UserProperyValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, upv); err != nil {
		return err
	}
	upv.uProperty.FirstName = upv.Property.FirstName
	upv.uProperty.LastName = upv.Property.LastName
	upv.uProperty.StartDate = upv.Property.StartDate
	upv.uProperty.EndDate = upv.Property.EndDate
	upv.uProperty.Limit = upv.Property.Limit
	upv.uProperty.Offset = upv.Property.Offset
	return nil
}
