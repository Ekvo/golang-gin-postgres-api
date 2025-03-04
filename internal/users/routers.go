package users

import (
	"errors"
	"fmt"
	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UserBeforeRegister(router *gin.RouterGroup, storeDB source.SQLSource) {
	router.POST("/signup", UserCreate(storeDB, NewUserCreateValidator()))
	router.POST("/login", UserLogin(storeDB, NewUserConnectLoginValidator(), source.FlagLogin))
	router.POST("/phone", UserLogin(storeDB, NewUserConnectPhoneValidator(), source.FlagPhone))
	router.POST("/email", UserLogin(storeDB, NewUserConnectEmailValidator(), source.FlagEmail))
}

func UserAfterRegister(router *gin.RouterGroup, storeDB source.SQLSource) {
	router.GET("/", UserRetrieve(storeDB))
	router.PUT("/", UserUpdate(storeDB))
}

func SpeakerFolower(router *gin.RouterGroup, storeDB source.SQLSource) {
	router.GET("/:nickname", ProfileRetrieve(storeDB))
	router.PUT("/:nickname/follow", ProfileFollow(storeDB))
	router.DELETE("/:nickname/follow", ProfileUnFollow(storeDB))
}

func UserLogin(db source.UserConnect, modelValidator UserValidator, flag int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		uModelFromValidator := modelValidator.Model()
		uModel, errDB := db.LoginUserWithUpdateTime(c.Request.Context(), uModelFromValidator, flag)
		if errDB != nil || !uModel.CheckPassword(uModelFromValidator.Password) {
			flagName := source.FlagName(flag)
			err := common.NewError("login", fmt.Errorf("not registred %s or %w", flagName, ErrUsersValidatorPassword))
			c.JSON(http.StatusNotFound, err)
			return
		}
		SetFlagsContext(c, uModel)

		serialize := UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userResponse})
	}
}

func UserCreate(db source.UserConnect, modelValidator UserValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		var err error = nil
		uModelFromValidator := modelValidator.Model()
		uModelFromValidator.ID, err = db.SaveOneUser(c.Request.Context(), uModelFromValidator)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", err))
			return
		}
		SetFlagsContext(c, uModelFromValidator)

		serialize := UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": userResponse})
	}
}

func UserRetrieve(db source.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {
		serialize := UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userResponse})
	}
}

func UserUpdate(db source.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {
		oldUserModel, ok := c.MustGet(userModel).(source.UserModel)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, common.ErrCommonUnexpectedType)
			return
		}
		modelValidator := NewUserCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		modelValidator.uModel.ID = oldUserModel.ID
		if err := db.NewDataUser(c.Request.Context(), modelValidator.uModel); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", err))
			return
		}
		if err := DataForContextUserModel(c, db, oldUserModel.ID); err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serialize := UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": userResponse})
	}
}

var ErrUsersInvalidProfile = errors.New("invalid login")

var ErrUsersNotFound = errors.New("not faound")

// ProfileRetrieve - проверят имеет ли подписку текуший пользователь из c.Keys[userModel]
//
//	на 'userLogin' полученного в базе при помощи 'login' из 'c.Param'
func ProfileRetrieve(db source.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, ok := common.IsValidParam(c, "nickname")
		if !ok {
			c.JSON(http.StatusBadRequest, common.NewError("profile", ErrUsersInvalidProfile))
			return
		}
		uModelSpeaker, err := db.FindOneUserByField(c.Request.Context(), source.UserModel{Login: userLogin}, source.FlagLogin)
		if err != nil {
			c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			return
		}
		serialize := ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}

func ProfileFollow(db source.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, ok := common.IsValidParam(c, "nickname")
		if !ok {
			c.JSON(http.StatusBadRequest, common.NewError("profile", ErrUsersInvalidProfile))
			return
		}
		uModelFollower, ok := c.MustGet(userModel).(source.UserModel)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, common.ErrCommonUnexpectedType)
			return
		}
		uModelSpeaker, err := db.FindOneUserByField(c.Request.Context(), source.UserModel{Login: userLogin}, source.FlagLogin)
		if err != nil {
			c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			return
		}
		if err := db.NewRelationship(c.Request.Context(), uModelFollower, uModelSpeaker); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("profile", err))
			return
		}
		serialize := ProfileSerializer{c, uModelSpeaker}
		profileresponse, err := serialize.Response(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"profile": profileresponse})
	}
}

func ProfileUnFollow(db source.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, ok := common.IsValidParam(c, "nickname")
		if !ok {
			c.JSON(http.StatusBadRequest, common.NewError("profile", ErrUsersNotFound))
			return
		}
		uModelFollowe, ok := c.MustGet(userModel).(source.UserModel)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, common.ErrCommonUnexpectedType)
			return
		}
		uModelSpeaker, err := db.FindOneUserByField(c.Request.Context(), source.UserModel{Login: userLogin}, source.FlagLogin)
		if err != nil {
			c.JSON(http.StatusNotFound, common.NewError("data_base", err))
			return
		}
		if err := db.EndRelationship(c.Request.Context(), uModelFollowe, uModelSpeaker); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("profile", err))
			return
		}
		serialize := ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}
