package users

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
)

type UserSerializer struct {
	C *gin.Context
}

type UserResponse struct {
	Login     string  `json:"login"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     string  `json:"email"`
	Image     *string `json:"image,omitempty"`
	Bio       *string `json:"biography,omitempty"`

	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAT      *time.Time `json:"updated_at,omitempty"`
	LastConnection *time.Time `json:"last_connect,omitempty"`

	Access string `json:"access"`
	// jwt.Token
	Token string `jsom:"token"`
	// возможны при обработке 'Responce()'
	Error error `json:"error,omitempty"`
}

func (us *UserSerializer) Responce() UserResponse {
	uModel, ok := us.C.MustGet(userModel).(source.UserModel)
	if !ok {
		return UserResponse{Error: errors.New("model - unexpected type")}
	}
	token, err := common.GenToken(uModel.ID)
	if err != nil {
		return UserResponse{Error: err}
	}
	uResponse := UserResponse{
		Login:          uModel.Login,
		FirstName:      uModel.FirstName,
		LastName:       uModel.LastName,
		Email:          uModel.Email,
		Phone:          uModel.Phone,
		Image:          uModel.Image,
		Bio:            uModel.Bio,
		CreatedAt:      uModel.CreatedAt,
		UpdatedAT:      uModel.UpdatedAT,
		LastConnection: uModel.LastConnection,
		Access:         uModel.Access,
		Token:          token,
		Error:          nil,
	}
	return uResponse
}
