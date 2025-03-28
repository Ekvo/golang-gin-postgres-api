package serializer

import (
	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	vr "github.com/Ekvo/golang-gin-postgres-api/internal/variables"
)

type UserSerializer struct {
	C *gin.Context
}

// UserResponse - для пердачи персональных данных в личный кабинет пользователя
type UserResponse struct {
	ID        uint    `json:"-"`
	Login     string  `json:"login"`
	Password  string  `json:"-"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     string  `json:"email"`
	Access    string  `json:"access"`
	Image     *string `json:"image,omitempty"`
	Bio       *string `json:"biography,omitempty"`

	CreatedAt      string `json:"created_at"`
	UpdatedAT      string `json:"updated_at,omitempty"`
	LastConnection string `json:"last_connect,omitempty"`
}

func (us *UserSerializer) Response() (UserResponse, error) {
	uModel := us.C.MustGet(flag.KeyUserModel).(models.UserModel)
	userRespnse := UserResponse{
		Login:     uModel.Login,
		FirstName: uModel.FirstName,
		LastName:  uModel.LastName,
		Email:     uModel.Email,
		Phone:     uModel.Phone,
		Access:    uModel.Access,
		Image:     uModel.Image,
		Bio:       uModel.Bio,
		CreatedAt: uModel.CreatedAt.UTC().Format(vr.RFC3339Milli),
	}
	if uModel.UpdatedAt != nil {
		userRespnse.UpdatedAT = uModel.UpdatedAt.UTC().Format(vr.RFC3339Milli)
	}
	if uModel.LastConnection != nil {
		userRespnse.LastConnection = uModel.LastConnection.UTC().Format(vr.RFC3339Milli)
	}
	return userRespnse, nil
}
