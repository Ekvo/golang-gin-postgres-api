package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type UserCreateValidator struct {
	User struct {
		Login     string `form:"login" json:"login" binding:"required,alphanum,min=1,max=128"`
		Password  string `form:"password" json:"password" binding:"required,min=8,max=255"`
		FirstName string `form:"first_name" json:"first_name" binding:"required,alpha,min=1,max=128"`
		LastName  string `form:"last_name" json:"last_name" binding:"omitempty,alpha,min=1,max=128"`
		Phone     string `form:"phone" json:"phone" binding:"omitempty,e164"`
		Email     string `form:"email" json:"email" binding:"required,email"`
		Access    string `from:"access" json:"access" binding:"required,len=1,numeric,excludesall=56789"`
		Image     string `form:"image" json:"image" binding:"omitempty,url"`
		Bio       string `form:"biography" json:"biography" binding:"omitempty,max=512"`
	} `json:"user_update"`
	uModel models.UserModel `json:"-"`
}

func NewUserCreateValidator() UserCreateValidator {
	return UserCreateValidator{}
}

func (ucv *UserCreateValidator) Model() models.UserModel {
	return ucv.uModel
}

func (ucv *UserCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, ucv); err != nil {
		return err
	}
	ucv.uModel.Login = ucv.User.Login
	if ucv.User.Password == common.TrickPassword {
		user := c.MustGet(flag.KeyUserModel).(models.UserModel)
		ucv.uModel.Password = user.Password
	} else {
		ucv.uModel.Password = ucv.User.Password
		ucv.uModel.HashPassword()
	}
	ucv.uModel.FirstName = ucv.User.FirstName
	if len(ucv.User.LastName) > 0 {
		ucv.uModel.LastName = &ucv.User.LastName
	}
	if len(ucv.User.Phone) > 0 {
		ucv.uModel.Phone = &ucv.User.Phone
	}
	ucv.uModel.Email = ucv.User.Email
	ucv.uModel.Access = ucv.User.Access
	if len(ucv.User.Image) > 0 {
		ucv.uModel.Image = &ucv.User.Image
	}
	if len(ucv.User.Bio) > 0 {
		ucv.uModel.Bio = &ucv.User.Bio
	}
	ucv.uModel.CreatedAt = time.Now()
	return nil
}
