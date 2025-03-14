package users

import (
	"github.com/gin-gonic/gin"
	"time"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type UserSerializer struct {
	C *gin.Context
}

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

	NumberOfFollowers uint `json:"followers"`

	// jwt.Token - пустой при поиске списка пользователей
	Token string `jsom:"token,omitempty"`
}

func (us *UserSerializer) Response() (UserResponse, error) {
	uModel := us.C.MustGet(models.KeyUserModel).(models.UserModel)
	token, err := common.GenToken(uModel.ID)
	if err != nil {
		return UserResponse{}, err
	}
	userRespnse := UserResponse{
		Login:             uModel.Login,
		FirstName:         uModel.FirstName,
		LastName:          uModel.LastName,
		Email:             uModel.Email,
		Phone:             uModel.Phone,
		Access:            uModel.Access,
		Image:             uModel.Image,
		Bio:               uModel.Bio,
		CreatedAt:         uModel.CreatedAt.Format(time.RFC3339Nano),
		NumberOfFollowers: uModel.NumberOfFollowers,
		Token:             token,
	}
	if uModel.UpdatedAt != nil {
		userRespnse.UpdatedAT = uModel.UpdatedAt.Format(time.RFC3339Nano)
	}
	if uModel.LastConnection != nil {
		userRespnse.LastConnection = uModel.LastConnection.Format(time.RFC3339Nano)
	}
	return userRespnse, nil
}

type ProfileSerializer struct {
	C *gin.Context
	models.UserModel
}

// Progileresponse - характеристики текущего пользователя
type ProfileResponse struct {
	Login             string  `json:"login"`
	FirstName         string  `json:"first_name"`
	Image             *string `json:"image,omitempty"`
	Bio               *string `json:"biography,omitempty"`
	NumberOfFollowers uint    `json:"followers"`

	CreatedAt      string `json:"created_at"`
	LastConnection string `json:"last_connect,omitempty"`

	// Following - состояние подписки во время сравнения с другим пользователем
	Following bool `json:"following"`
}

func (ps *ProfileSerializer) Response(db models.UserFollowing) (ProfileResponse, error) {
	uModel := ps.C.MustGet(models.KeyUserModel).(models.UserModel)
	following, err := db.IsRelationship(ps.C.Request.Context(), []uint{uModel.ID, ps.UserModel.ID})
	if err != nil {
		return ProfileResponse{}, err
	}
	profileResponse := ProfileResponse{
		Login:             uModel.Login,
		FirstName:         uModel.FirstName,
		Image:             uModel.Image,
		Bio:               uModel.Bio,
		NumberOfFollowers: uModel.NumberOfFollowers,
		CreatedAt:         uModel.CreatedAt.Format(time.RFC3339Nano),
		Following:         following,
	}
	if uModel.LastConnection != nil {
		profileResponse.LastConnection = uModel.LastConnection.Format(time.RFC3339Nano)
	}
	return profileResponse, nil
}

type ProfileListSerializer struct {
	C     *gin.Context
	Users []models.UserModel
}

func (pls *ProfileListSerializer) Response(db models.UserFollowing) ([]ProfileResponse, error) {
	arrProfileResponse := make([]ProfileResponse, 0, len(pls.Users))
	for _, user := range pls.Users {
		serialize := ProfileSerializer{pls.C, user}
		profileResponse, err := serialize.Response(db)
		if err != nil {
			return nil, err
		}
		arrProfileResponse = append(arrProfileResponse, profileResponse)
	}
	return arrProfileResponse, nil
}
