package transport

import (
	"context"
	"errors"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	mod "github.com/Ekvo/golang-gin-postgres-api/internal/models"
	usr "github.com/Ekvo/golang-gin-postgres-api/internal/services/users"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
)

// ctxTimeRequest - для инициализации ctx в запросах
// context.WithTimeout(c.Request.Context(),ctxTimeRequest)
const CTXUsersTimeRequest = 1000500 * time.Millisecond

func UserBeforeRegister(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.POST("/signup", UserCreate(storeDB, usr.NewUserCreateValidator()))
	router.POST("/login", UserLogin[usr.UCLV](storeDB, usr.NewUserConnectLoginValidator(), mod.FlagLogin))
	router.POST("/phone", UserLogin[usr.UCPV](storeDB, usr.NewUserConnectPhoneValidator(), mod.FlagPhone))
	router.POST("/email", UserLogin[usr.UCEV](storeDB, usr.NewUserConnectEmailValidator(), mod.FlagEmail))
}

func UserAfterRegister(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.GET("/", UserRetrieve())
	router.POST("/", UsersRetrieve(storeDB))
	router.PUT("/", UserUpdate(storeDB))

}

func SpeakerFolower(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.GET("/:nickname", ProfileRetrieve(storeDB))
	router.PUT("/:nickname/follow", ProfileFollow(storeDB))
	router.DELETE("/:nickname/follow", ProfileUnFollow(storeDB))
}

func UserLogin[V any](db mod.UserConnect, modelValidator mod.ValidatorModel[V, mod.UM], flag int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, flag)
		if ctx.Err() != nil {
			return
		}
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		uModelFromValidator := modelValidator.Model()
		uModel, errDB := db.LoginUserWithUpdateTime(ctx, uModelFromValidator)
		if errDB != nil {
			if errDB != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("data_base", errDB))
			}
			return
		}
		if !uModel.CheckPassword(uModelFromValidator.Password) {
			c.JSON(http.StatusForbidden, common.NewError("login", usr.ErrUsersValidatorPassword))
			return
		}
		usr.SetFlagsContext(c, uModel)
		serialize := usr.UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userResponse})
	}
}

func UserCreate(db mod.UserConnect, modelValidator mod.ValidatorModel[usr.UCV, mod.UserModel]) gin.HandlerFunc {
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
		usr.SetFlagsContext(c, uModelFromValidator)

		serialize := usr.UserSerializer{c}
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
		serialize := usr.UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userResponse})
	}
}

func UsersRetrieve(db mod.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		modelValidator := usr.NewUserProperyValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		usersList, err := db.FindUserList(ctx, modelValidator.Model())
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			}
			return
		}
		if len(usersList) == 0 {
			c.JSON(http.StatusNoContent, gin.H{"empty": ""})
			return
		}
		serialize := usr.ProfileListSerializer{c, usersList}
		userresponse, err := serialize.Response(db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": userresponse})
	}
}

func UserUpdate(db mod.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		modelValidator := usr.NewUserCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		uModel := modelValidator.Model()
		uModel.ID = c.MustGet(mod.KeyUserID).(uint)
		uModel.UpdatedAt = &uModel.CreatedAt
		if err := db.NewDataUser(ctx, uModel); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", err))
			}
			return
		}
		if err := usr.DataForContextUserModel(c, db, uModel.ID); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			}
			return
		}
		serialize := usr.UserSerializer{c}
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
func ProfileRetrieve(db mod.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, mod.FlagLogin)
		if ctx.Err() != nil {
			return
		}
		userLogin, ok := common.IsValidParam(c, "nickname")
		if !ok {
			c.JSON(http.StatusBadRequest, common.NewError("profile", ErrUsersInvalidProfile))
			return
		}
		uModelSpeaker, err := db.FindOneUserByField(ctx, mod.UserModel{Login: userLogin})
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			}
			return
		}
		serialize := usr.ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}

func ProfileFollow(db mod.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, mod.FlagLogin)
		if ctx.Err() != nil {
			return
		}
		userLogin := c.Param("nickname")
		uModelFollower := c.MustGet(mod.KeyUserModel).(mod.UserModel)
		uModelSpeaker, err := db.FindOneUserByField(ctx, mod.UserModel{Login: userLogin})
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			}
			return
		}
		if err := db.NewRelationship(ctx, []uint{uModelFollower.ID, uModelSpeaker.ID}); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("profile", err))
			}
			return
		}
		serialize := usr.ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			}
			return
		}
		c.JSON(http.StatusCreated, gin.H{"profile": profileResponse})
	}
}

func ProfileUnFollow(db mod.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, mod.FlagLogin)
		if ctx.Err() != nil {
			return
		}
		userLogin := c.Param("nickname")
		uModelSpeaker, err := db.FindOneUserByField(ctx, mod.UserModel{Login: userLogin})
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("data_base", err))
			}
			return
		}
		uModelFollower := c.MustGet(mod.KeyUserModel).(mod.UserModel)
		if err := db.EndRelationship(c.Request.Context(), []uint{uModelFollower.ID, uModelSpeaker.ID}); err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("data_base", err))
			}
			return
		}
		serialize := usr.ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}
