package users

import (
	"errors"
	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	"net/http"
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

// ключи для записи в 'gin.Context.Keys'
const (
	userID     = "user_id"
	userAccess = "user_access"
	userModel  = "user_model"
)

// SetFlagsContext - запись в 'gin.Context.Keys' по ключам 'userID','userAccess' и 'userModel'
func SetFlagsContext(c *gin.Context, uModel source.UserModel) {
	c.Set(userID, uModel.ID)
	c.Set(userAccess, uModel.Access)
	c.Set(userModel, uModel)
}

// DataForContextUserModel -  если 'id_user != 0' получение данных пользователя
// и записи в 'gin.Context.Keys'
// с возможностью выбирать базу данных
func DataForContextUserModel(c *gin.Context, db source.UserApprove, id_user uint) error {
	var uModel source.UserModel
	if id_user != 0 {
		var err error = nil
		uModel, err = db.FindOneUserByField(c.Request.Context(), source.UserModel{ID: id_user}, source.FlagID)
		if err != nil {
			return err
		}
	}
	SetFlagsContext(c, uModel)
	return nil
}

// Autorization - авторизация, получение данныx для последующих gin.HandlerFunc в gin.Group
// с аозможностью выбирать базу данных
func Autorization(db source.UserApprove) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = DataForContextUserModel(c, nil, 0)
		token, errToken := request.ParseFromRequest(c.Request, AuthExtractor, func(token *jwt.Token) (interface{}, error) {
			key := []byte(common.SecretKey)
			return key, nil
		})
		if errToken != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, common.NewError("autorization", errToken))
			return
		}
		user_id, err := common.Indeficator(token, "id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, common.NewError("autorization", err))
			return
		}
		if err := DataForContextUserModel(c, db, uint(user_id)); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
	}
}
