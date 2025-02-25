package users

import (
	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/gin-gonic/gin"
)

type UserCreateValidator struct {
	User struct {
		Login     string `form:"login" json:"login" binding:"required,alphanum,min=1,max=128"`
		Password  string `form:"password" json:"password" binding:"required,min=8,max=255"`
		FirstName string `form:"first_name" json:"first_name" binding:"required,alphanum,min=1,max=128"`
		LastName  string `form:"last_name" json:"last_name" binding:"omitempty,alphanum,min=1,max=128"`
		Phone     string `form:"phone" json:"phone" binding:"omitempty,e164"`
		Email     string `form:"email" json:"email" binding:"required,email"`
		Access    string `from:"access" json:"access" binding:"omitempty,min=1,max=3"`
		Image     string `form:"image" json:"image" binding:"omitempty,url"`
		Bio       string `form:"biography" json:"biography" binding:"omitempty,max=512"`
	} `json:"user_update"`
	uModel UserModel `json:"-"`
}

// удалить !!!
func (ucv *UserCreateValidator) D() UserModel {
	return ucv.uModel
}

// Если необходимы значения по-умолчанию.
func NewUserCreateValidator() UserCreateValidator {
	return UserCreateValidator{}
}

func (ucv *UserCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, ucv); err != nil {
		return err
	}
	ucv.uModel.Login = ucv.User.Login
	ucv.uModel.FirstName = ucv.User.FirstName
	ucv.uModel.Email = ucv.User.Email
	ucv.uModel.Password = ucv.User.Password
	if err := ucv.uModel.hashPassword(); err != nil {
		return err
	}

	if ucv.User.LastName != "" {
		ucv.uModel.LastName = ucv.User.LastName
	}
	if ucv.User.Phone != "" {
		ucv.uModel.Phone = ucv.User.Phone
	}
	if ucv.User.Image != "" {
		ucv.uModel.Image = ucv.User.Image
	}
	if len(ucv.User.Bio) != 0 {
		ucv.uModel.Bio = ucv.User.Bio
	}

	return nil
}

type UserConnectLoginValidator struct {
	UserLogin struct {
		Login    string `form:"login" json:"login" binding:"required,alphanum,min=1,max=128"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	} `json:"user_connect_with_login"`
	uModel UserModel `json:"-"`
}

// удалить !!!
func (ucv *UserConnectLoginValidator) D() UserModel {
	return ucv.uModel
}

func NewUserConnectLoginValidator() UserConnectLoginValidator {
	return UserConnectLoginValidator{}
}

func (ulv *UserConnectLoginValidator) Bind(c *gin.Context) error {
	err := common.Bind(c, ulv)
	if err != nil {
		return err
	}
	ulv.uModel.Login = ulv.UserLogin.Login
	ulv.uModel.Password = ulv.UserLogin.Password

	return nil
}

type UserConnectEmailValidator struct {
	UserEmail struct {
		Email    string `form:"email" json:"email" binding:"required,email"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	} `json:"user_connect_with_email"`
	uModel UserModel `json:"-"`
}

// удалить !!!
func (ucv *UserConnectEmailValidator) D() UserModel {
	return ucv.uModel
}

func NewUserConnectEmailValidator() UserConnectEmailValidator {
	return UserConnectEmailValidator{}
}

func (uev *UserConnectEmailValidator) Bind(c *gin.Context) error {
	err := common.Bind(c, uev)
	if err != nil {
		return err
	}
	uev.uModel.Email = uev.UserEmail.Email
	uev.uModel.Password = uev.UserEmail.Password

	return nil
}
