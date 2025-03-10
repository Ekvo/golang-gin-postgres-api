package transport

import (
	"context"
	"errors"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/users"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
)

// ctxTimeRequest - для инициализации ctx в запросах
// context.WithTimeout(c.Request.Context(),ctxTimeRequest)
const CTXUsersTimeRequest = 1000500 * time.Millisecond

func UserBeforeRegister(router *gin.RouterGroup, storeDB source.SQLSource) {
	router.POST("/signup", UserCreate(storeDB, users.NewUserCreateValidator()))
	router.POST("/login", UserLogin(storeDB, users.NewUserConnectLoginValidator(), models.FlagLogin))
	router.POST("/phone", UserLogin(storeDB, users.NewUserConnectPhoneValidator(), models.FlagPhone))
	router.POST("/email", UserLogin(storeDB, users.NewUserConnectEmailValidator(), models.FlagEmail))
}

func UserAfterRegister(router *gin.RouterGroup, storeDB source.SQLSource) {
	router.GET("/", UserRetrieve())
	router.PUT("/", UserUpdate(storeDB))
}

func SpeakerFolower(router *gin.RouterGroup, storeDB source.SQLSource) {
	router.GET("/:nickname", ProfileRetrieve(storeDB))
	router.PUT("/:nickname/follow", ProfileFollow(storeDB))
	router.DELETE("/:nickname/follow", ProfileUnFollow(storeDB))
}

func UserLogin(db models.UserConnect, modelValidator models.ValidatorModel[models.UserModel], flag int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		uModelFromValidator := modelValidator.Model()
		uModel, errDB := db.LoginUserWithUpdateTime(ctx, uModelFromValidator, flag)
		if errDB != nil {
			if errDB != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("data_base", errDB))
			}
			return
		}
		if !uModel.CheckPassword(uModelFromValidator.Password) {
			c.JSON(http.StatusForbidden, common.NewError("login", users.ErrUsersValidatorPassword))
			return
		}

		users.SetFlagsContext(c, uModel)
		serialize := users.UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userResponse})
	}
}

func UserCreate(db models.UserConnect, modelValidator models.ValidatorModel[models.UserModel]) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		var err error = nil
		uModelFromValidator := modelValidator.Model()
		uModelFromValidator.ID, err = db.SaveOneUser(ctx, uModelFromValidator)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", err))
			}
			return
		}
		users.SetFlagsContext(c, uModelFromValidator)

		serialize := users.UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": userResponse})
	}
}

func UserRetrieve() gin.HandlerFunc {
	return func(c *gin.Context) {
		serialize := users.UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userResponse})
	}
}

func UserUpdate(db models.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		userID, ok := c.MustGet(services.UserID).(uint)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, common.ErrCommonUnexpectedType)
			return
		}
		modelValidator := users.NewUserCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		uModel := modelValidator.Model()
		uModel.ID = userID
		uModel.UpdatedAT = &uModel.CreatedAt
		if err := db.NewDataUser(ctx, uModel); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", err))
			}
			return
		}
		if err := users.DataForContextUserModel(c, db, userID); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			}
			return
		}
		serialize := users.UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": userResponse})
	}
}

// ErrUsersInvalidProfile - если запрос пришел с пустым параметром 'nickname'
var ErrUsersInvalidProfile = errors.New("invalid login")

// ErrUsersNotFound - пользователь ненайден
var ErrUsersNotFound = errors.New("not faound")

// ProfileRetrieve - проверят имеет ли подписку текуший пользователь из c.Keys[userModel]
//
//	на 'userLogin' полученного в базе при помощи 'login' из 'c.Param'
func ProfileRetrieve(db models.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		userLogin, ok := common.IsValidParam(c, "nickname")
		if !ok {
			c.JSON(http.StatusBadRequest, common.NewError("profile", ErrUsersInvalidProfile))
			return
		}
		uModelSpeaker, err := db.FindOneUserByField(ctx, models.UserModel{Login: userLogin}, models.FlagLogin)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			}
			return
		}
		serialize := users.ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(ctx, db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}

func ProfileFollow(db models.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		userLogin, ok := common.IsValidParam(c, "nickname")
		if !ok {
			c.JSON(http.StatusBadRequest, common.NewError("profile", ErrUsersInvalidProfile))
			return
		}
		uModelFollower, ok := c.MustGet(services.UserModel).(models.UserModel)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, common.ErrCommonUnexpectedType)
			return
		}
		uModelSpeaker, err := db.FindOneUserByField(ctx, models.UserModel{Login: userLogin}, models.FlagLogin)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			}
			return
		}
		if err := db.NewRelationship(ctx, uModelFollower, uModelSpeaker); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("profile", err))
			}
			return
		}
		serialize := users.ProfileSerializer{c, uModelSpeaker}
		profileresponse, err := serialize.Response(ctx, db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			}
			return
		}
		c.JSON(http.StatusCreated, gin.H{"profile": profileresponse})
	}
}

func ProfileUnFollow(db models.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		userLogin, ok := common.IsValidParam(c, "nickname")
		if !ok {
			c.JSON(http.StatusBadRequest, common.NewError("profile", ErrUsersNotFound))
			return
		}
		uModelFollowe, ok := c.MustGet(services.UserModel).(models.UserModel)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, common.ErrCommonUnexpectedType)
			return
		}
		uModelSpeaker, err := db.FindOneUserByField(ctx, models.UserModel{Login: userLogin}, models.FlagLogin)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("data_base", err))
			}
			return
		}
		if err := db.EndRelationship(c.Request.Context(), uModelFollowe, uModelSpeaker); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("data_base", err))
			}
			return
		}
		serialize := users.ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(ctx, db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}
