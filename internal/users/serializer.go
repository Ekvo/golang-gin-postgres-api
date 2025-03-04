package users

import (
	"github.com/gin-gonic/gin"
	"time"

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
	Access    string  `json:"access"`
	Image     *string `json:"image,omitempty"`
	Bio       *string `json:"biography,omitempty"`

	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAT      *time.Time `json:"updated_at,omitempty"`
	LastConnection *time.Time `json:"last_connect,omitempty"`

	// jwt.Token
	Token string `jsom:"token"`
}

func (us *UserSerializer) Response() (UserResponse, error) {
	uModel, ok := us.C.MustGet(userModel).(source.UserModel)
	if !ok {
		return UserResponse{}, common.ErrCommonUnexpectedType
	}
	token, err := common.GenToken(uModel.ID)
	if err != nil {
		return UserResponse{}, err
	}
	uResponse := UserResponse{
		Login:          uModel.Login,
		FirstName:      uModel.FirstName,
		LastName:       uModel.LastName,
		Email:          uModel.Email,
		Phone:          uModel.Phone,
		Access:         uModel.Access,
		Image:          uModel.Image,
		Bio:            uModel.Bio,
		CreatedAt:      uModel.CreatedAt,
		UpdatedAT:      uModel.UpdatedAT,
		LastConnection: uModel.LastConnection,
		Token:          token,
	}
	return uResponse, nil
}

type ProfileSerializer struct {
	c *gin.Context
	source.UserModel
}

// Progileresponse - характеристики текущего пользователя
type ProfileResponse struct {
	Login     string  `json:"login"`
	FirstName string  `json:"first_name"`
	Image     *string `json:"image,omitempty"`
	Bio       *string `json:"biography,omitempty"`

	// Following - состояние подписки во время сравнения с другим пользователем
	Following bool `json:"following"`
}

func (ps *ProfileSerializer) Response(db source.UserFollowing) (ProfileResponse, error) {
	uModel, ok := ps.c.MustGet(userModel).(source.UserModel)
	if !ok {
		return ProfileResponse{}, common.ErrCommonUnexpectedType
	}
	following, err := db.IsRelationship(ps.c.Request.Context(), uModel, ps.UserModel)
	if err != nil {
		return ProfileResponse{}, err
	}
	return ProfileResponse{
		Login:     ps.Login,
		FirstName: ps.FirstName,
		Image:     ps.Image,
		Bio:       ps.Bio,
		Following: following,
	}, nil
}
