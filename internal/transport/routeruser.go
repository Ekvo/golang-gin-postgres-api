package transport

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	mod "github.com/Ekvo/golang-gin-postgres-api/internal/models"
	usr "github.com/Ekvo/golang-gin-postgres-api/internal/services/users"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// ctxTimeRequest - для инициализации ctx в запросах
// context.WithTimeout(c.Request.Context(),ctxTimeRequest)
const CTXUsersTimeRequest = 100500 * time.Millisecond

func UserBeforeRegister(router *gin.RouterGroup, db mod.UserConnect) {
	router.POST("/signup", UserCreate(db))
	router.POST("/login", UserLogin(db))
}

func UserAfterRegister(router *gin.RouterGroup, db mod.UserApproveFollowing) {
	router.GET("/", UserRetrieve())
	router.PUT("/", UserUpdate(db))
}

func SpeakerFolower(router *gin.RouterGroup, db mod.UserApproveFollowing) {
	router.GET("/:nickname", ProfileRetrieve(db))
	router.PUT("/:nickname/follow", ProfileFollow(db))
	router.DELETE("/:nickname/follow", ProfileUnFollow(db))

	router.POST("/", ProfileListRetrieve(db))
}

func UserCreate(db mod.UserConnect) gin.HandlerFunc {
	return func(c *gin.Context) {
		modelValidator := usr.NewUserCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		var err error = nil
		uModelFromValidator := modelValidator.Model()
		uModelFromValidator.ID, err = db.SaveOneUser(c.Request.Context(), uModelFromValidator)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", source.ErrSourceAlreadyExists))
			return
		}
		usr.SetFlagsContext(c, uModelFromValidator)

		serialize := usr.TokenSerializer{c}
		tokenResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"approve": tokenResponse})
	}
}

func UserLogin(db mod.UserConnect) gin.HandlerFunc {
	return func(c *gin.Context) {
		modelValidator := usr.NewUserConnectLoginValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		uModelFromValidator := modelValidator.Model()
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, mod.FlagLogin)
		user, errDB := db.LoginUserWithUpdateTime(ctx, uModelFromValidator)
		if errDB != nil {
			if errDB == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("login", source.ErrSourceNotFound))
			return
		}
		if !user.CheckPassword(uModelFromValidator.Password) {
			c.JSON(http.StatusForbidden, common.NewError("login", mod.ErrModelsPassword))
			return
		}
		usr.SetFlagsContext(c, user)
		serialize := usr.TokenSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"approve": userResponse})
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

func UserUpdate(db mod.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {
		modelValidator := usr.NewUserCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		uModel := modelValidator.Model()
		uModel.ID = c.MustGet(mod.KeyUserID).(uint)
		uModel.UpdatedAt = &uModel.CreatedAt
		if err := db.NewDataUser(c.Request.Context(), uModel); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusUnprocessableEntity, common.NewError("data_base", err))
			return
		}
		if err := usr.DataForContextUserModel(c, db, uModel.ID); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serialize := usr.UserSerializer{c}
		userResponse, err := serialize.Response()
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userResponse})
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
		userLogin := c.Param("nickname")
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, mod.FlagLogin)
		uModelSpeaker, err := db.FindOneUserByField(ctx, mod.UserModel{Login: userLogin})
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			return
		}
		serialize := usr.ProfileSerializer{c, uModelSpeaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}

func ProfileFollow(db mod.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin := c.Param("nickname")
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, mod.FlagLogin)
		speaker, err := db.FindOneUserByField(ctx, mod.UserModel{Login: userLogin})
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("profile", ErrUsersNotFound))
			return
		}
		userID := c.MustGet(mod.KeyUserID).(uint)
		if err := db.NewRelationship(ctx, []uint{userID, speaker.ID}); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusUnprocessableEntity, common.NewError("profile", err))
			return
		}
		speaker.NumberOfFollowers++
		serialize := usr.ProfileSerializer{c, speaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"profile": profileResponse})
	}
}

func ProfileUnFollow(db mod.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin := c.Param("nickname")
		ctx := context.WithValue(c.Request.Context(), mod.KeyFlagFiled, mod.FlagLogin)
		speaker, err := db.FindOneUserByField(ctx, mod.UserModel{Login: userLogin})
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("data_base", err))
			return
		}
		userID := c.MustGet(mod.KeyUserID).(uint)
		if err := db.EndRelationship(c.Request.Context(), []uint{userID, speaker.ID}); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("data_base", err))
			return
		}
		speaker.NumberOfFollowers--
		serialize := usr.ProfileSerializer{c, speaker}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": profileResponse})
	}
}

func ProfileListRetrieve(db mod.UserApproveFollowing) gin.HandlerFunc {
	return func(c *gin.Context) {
		modelValidator := usr.NewUserProperyValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		usersList, err := db.FindUserList(c.Request.Context(), modelValidator.Model())
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		if len(usersList) == 0 {
			c.JSON(http.StatusNoContent, nil)
			return
		}
		serialize := usr.ProfileListSerializer{c, usersList}
		userresponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serialize", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": userresponse})
	}
}
