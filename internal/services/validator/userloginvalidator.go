package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type UserConnectLoginValidator struct {
	UserLogin struct {
		Login    string `form:"login" json:"login" binding:"required,alphanum,min=2,max=128"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	} `json:"user_connect_with_login"`
	uModel models.UserModel `json:"-"`
}

func NewUserConnectLoginValidator() UserConnectLoginValidator {
	return UserConnectLoginValidator{}
}

func (ulv *UserConnectLoginValidator) Model() models.UserModel {
	return ulv.uModel
}

func (ulv *UserConnectLoginValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, ulv); err != nil {
		return err
	}
	ulv.uModel.Login = ulv.UserLogin.Login
	ulv.uModel.Password = ulv.UserLogin.Password
	timeConnect := time.Now()
	ulv.uModel.LastConnection = &timeConnect
	return nil
}
