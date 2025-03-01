package users

import (
	"context"
	"fmt"
	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func UserBeforeRegister(router *gin.RouterGroup, storeDB source.SQLSource) {
	router.POST("/signup", UserCreate(storeDB, NewUserCreateValidator()))
	router.POST("/login", UserLogin(storeDB, NewUserConnectLoginValidator(), source.FlagLogin))
	router.POST("/phone", UserLogin(storeDB, NewUserConnectPhoneValidator(), source.FlagPhone))
	router.POST("/email", UserLogin(storeDB, NewUserConnectEmailValidator(), source.FlagEmail))
}

func UserAfterRegister(router *gin.RouterGroup, storeDB source.SQLSource) {
	//router.GET("/", (storeDB))
	//router.PUT("/")
}

func UserLogin(db source.UserConnect, modelValidator UserValidator, flag int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		uModelFromValidator := modelValidator.Model()
		uModel, errDB := db.LoginUserWithUpdateTime(ctx, uModelFromValidator, flag)
		if errDB != nil || !uModel.CheckPassword(uModelFromValidator.Password) {
			flagName := source.FlagName(flag)
			err := common.NewError("login", fmt.Errorf("not registred %s or %w", flagName, ErrUsersValidatorPassword))
			c.JSON(http.StatusNotFound, err)
			return
		}
		c.Set(userModel, uModel)
		serialize := UserSerializer{c}
		c.JSON(http.StatusOK, gin.H{"user": serialize.Responce()})
	}
}

func UserCreate(db source.UserConnect, modelValidator UserValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		uModelFromValidator := modelValidator.Model()
		if err := db.SaveOneUser(ctx, uModelFromValidator); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", err))
			return
		}
		c.Set(userModel, uModelFromValidator)
		serializer := UserSerializer{c}
		c.JSON(http.StatusOK, gin.H{"user": serializer.Responce()})
	}
}

func UserRetrive(db source.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {

		serializer := UserSerializer{c}
		c.JSON(http.StatusOK, gin.H{"user": serializer.Responce()})
	}
}
