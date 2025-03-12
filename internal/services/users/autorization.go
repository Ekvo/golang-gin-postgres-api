package users

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

var ErrUsersautorizationToken = errors.New("token from Authorization - incorrect")

// lenBearer - длина prefix в token
const lenBearer = 7 // "Bearer "

// truncateBearerToken - удаляем prefix из 'token'
// ищем "Bearer " и удаляем или ошибка
func truncateBearerToken(token string) (string, error) {
	if len(token) > lenBearer && token[:lenBearer] == "Bearer " {
		return token[lenBearer:], nil
	}
	return "", ErrUsersautorizationToken
}

// AuthHeaderExtractor - для получения token из header с помощью 'truncateBearerToken'(см. выше)
var AuthHeaderExtractor = &request.PostExtractionFilter{
	Extractor: request.HeaderExtractor{"Authorization"},
	Filter:    truncateBearerToken,
}

// AuthExtractor - набор экстракторов (при необходимости можно расширить пул)
var AuthExtractor = &request.MultiExtractor{
	AuthHeaderExtractor,
}

// SetFlagsContext - запись в 'gin.Context.Keys' по ключам 'userID','userAccess' и 'userModel'
func SetFlagsContext(c *gin.Context, uModel models.UserModel) {
	c.Set(models.KeyUserID, uModel.ID)
	c.Set(models.KeyUserModel, uModel)
	c.Set(models.KeyUserAccess, uModel.Access)
}

// DataForContextUserModel -  если 'id_user != 0' получение данных пользователя
// и записи в 'gin.Context.Keys'
// с возможностью выбирать базу данных
func DataForContextUserModel(c *gin.Context, db models.UserApprove, id_user uint) error {
	ctx := context.WithValue(c.Request.Context(), models.KeyFlagFiled, models.FlagID)
	if err := ctx.Err(); err != nil {
		return err
	}
	var uModel models.UserModel
	if id_user != 0 {
		var err error = nil
		uModel, err = db.FindOneUserByField(ctx, models.UserModel{ID: id_user})
		if err != nil {
			return err
		}
	}
	SetFlagsContext(c, uModel)
	return nil
}

// Autorization - авторизация, получение данныx для последующих gin.HandlerFunc в gin.Group
// с аозможностью выбирать базу данных
func Autorization(db models.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Context().Err() != nil {
			return
		}
		_ = DataForContextUserModel(c, nil, 0)
		token, errToken := request.ParseFromRequest(c.Request, AuthExtractor, func(token *jwt.Token) (interface{}, error) {
			key := []byte(common.SecretKey)
			return key, nil
		})
		if errToken != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, common.NewError("autorization", errToken))
			return
		}
		user_id, err := common.IDFromToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, common.NewError("autorization", err))
			return
		}
		if err := DataForContextUserModel(c, db, uint(user_id)); err != nil {
			if err != context.DeadlineExceeded {
				c.AbortWithStatusJSON(http.StatusNotFound, common.NewError("data_base", err))
			}
			return
		}
	}
}
