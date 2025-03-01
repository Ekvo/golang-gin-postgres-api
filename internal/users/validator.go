package users

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
)

var ErrUsersValidatorPassword = errors.New("invalid password")

type UserValidator interface {
	Bind(c *gin.Context) error
	Model() source.UserModel
}

type UserCreateValidator struct {
	User struct {
		Login     string `form:"login" json:"login" binding:"required,alphanum,min=1,max=128"`
		Password  string `form:"password" json:"password" binding:"required,min=8,max=255"`
		FirstName string `form:"first_name" json:"first_name" binding:"required,alpha,min=1,max=128"`
		LastName  string `form:"last_name" json:"last_name" binding:"omitempty,alpha,min=1,max=128"`
		Phone     string `form:"phone" json:"phone" binding:"omitempty,e164"`
		Email     string `form:"email" json:"email" binding:"required,email"`
		Access    string `from:"access" json:"access" binding:"required,len=1,numeric,excludesall=89"`
		Image     string `form:"image" json:"image" binding:"omitempty,url"`
		Bio       string `form:"biography" json:"biography" binding:"omitempty,max=512"`
	} `json:"user_update"`
	uModel source.UserModel `json:"-"`
}

// Если необходимо добавить значения по-умолчанию
func NewUserCreateValidator() *UserCreateValidator {
	return &UserCreateValidator{}
}

func (ucv *UserCreateValidator) Model() source.UserModel {
	return ucv.uModel
}

func (ucv *UserCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, ucv); err != nil {
		return err
	}
	ucv.uModel.Login = ucv.User.Login
	ucv.uModel.Password = ucv.User.Password
	_ = ucv.uModel.HashPassword()
	ucv.uModel.FirstName = ucv.User.FirstName
	ucv.uModel.Email = ucv.User.Email
	ucv.uModel.Access = ucv.User.Access
	ucv.uModel.LastName = &ucv.User.LastName
	ucv.uModel.Phone = &ucv.User.Phone
	ucv.uModel.Image = &ucv.User.Image
	ucv.uModel.Bio = &ucv.User.Bio

	createTime := time.Now()
	ucv.uModel.CreatedAt = createTime
	return nil
}

type UserConnectLoginValidator struct {
	UserLogin struct {
		Login    string `form:"login" json:"login" binding:"required,alphanum,min=1,max=128"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	} `json:"user_connect_with_login"`
	uModel source.UserModel `json:"-"`
}

// Если необходимо добавить значения по-умолчанию
func NewUserConnectLoginValidator() *UserConnectLoginValidator {
	return &UserConnectLoginValidator{}
}

func (ulv *UserConnectLoginValidator) Model() source.UserModel {
	return ulv.uModel
}

func (ulv *UserConnectLoginValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, ulv); err != nil {
		return err
	}
	ulv.uModel.Login = ulv.UserLogin.Login
	ulv.uModel.Password = ulv.UserLogin.Password
	return nil
}

type UserConnectPhoneValidator struct {
	UserPhone struct {
		Phone    string `form:"phone" json:"phone" binding:"required,e164"`
		Password string `form:"password" json:"password" binding:"required,min=8max=255"`
	} `json:"user_connect_wiht_phone"`
	uModel source.UserModel `json:"-"`
}

// Если необходимо добавить значения по-умолчанию
func NewUserConnectPhoneValidator() *UserConnectPhoneValidator {
	return &UserConnectPhoneValidator{}
}

func (upv *UserConnectPhoneValidator) Model() source.UserModel {
	return upv.uModel
}

func (upv *UserConnectPhoneValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, upv); err != nil {
		return err
	}
	upv.uModel.Phone = &upv.UserPhone.Phone
	upv.uModel.Password = upv.UserPhone.Password
	return nil
}

type UserConnectEmailValidator struct {
	UserEmail struct {
		Email    string `form:"email" json:"email" binding:"required,email"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	} `json:"user_connect_with_email"`
	uModel source.UserModel `json:"-"`
}

// Если необходимо добавить значения по-умолчанию
func NewUserConnectEmailValidator() *UserConnectEmailValidator {
	return &UserConnectEmailValidator{}
}

func (uev *UserConnectEmailValidator) Model() source.UserModel {
	return uev.uModel
}

func (uev *UserConnectEmailValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, uev); err != nil {
		return err
	}
	uev.uModel.Email = uev.UserEmail.Email
	uev.uModel.Password = uev.UserEmail.Password
	return nil
}
