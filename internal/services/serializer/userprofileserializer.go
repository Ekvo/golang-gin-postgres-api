package serializer

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	vr "github.com/Ekvo/golang-gin-postgres-api/internal/variables"
)

type ProfileSerializer struct {
	C *gin.Context
	models.UserModel
}

// Progileresponse - открытые характеристики текущего пользователя для других пользователей
type ProfileResponse struct {
	Login             string  `json:"login"`
	FirstName         string  `json:"first_name"`
	Image             *string `json:"image,omitempty"`
	Bio               *string `json:"biography,omitempty"`
	NumberOfFollowers uint    `json:"followers"`

	CreatedAt      string `json:"created_at"`
	LastConnection string `json:"last_connect,omitempty"`

	// Following - состояние подписки во время сравнения с пользователем делающий запрос
	Following bool `json:"following"`
}

func (ps *ProfileSerializer) Response(db models.UserFollowing) (ProfileResponse, error) {
	userID := ps.C.MustGet(flag.KeyUserID).(uint)
	following, err := db.IsRelationship(ps.C.Request.Context(), []uint{userID, ps.UserModel.ID})
	if err != nil {
		return ProfileResponse{}, err
	}
	profileResponse := ProfileResponse{
		Login:             ps.Login,
		FirstName:         ps.FirstName,
		Image:             ps.Image,
		Bio:               ps.Bio,
		NumberOfFollowers: ps.NumberOfFollowers,
		CreatedAt:         ps.CreatedAt.UTC().Format(vr.RFC3339Milli),
		Following:         following,
	}
	if ps.LastConnection != nil {
		profileResponse.LastConnection = ps.LastConnection.UTC().Format(vr.RFC3339Milli)
	}
	return profileResponse, nil
}

type ProfileListSerializer struct {
	C     *gin.Context
	Users []models.UserModel
}

func (pls *ProfileListSerializer) Response(db models.UserFollowing) ([]ProfileResponse, error) {
	// array with users ID to string
	lineSpeakerID := strings.Trim(strings.Replace(fmt.Sprint(pls.userIDList()), " ", ",", -1), "[]")
	if len(lineSpeakerID) == 0 {
		//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! need err
		return nil, nil
	}
	ctx := context.WithValue(pls.C.Request.Context(), flag.KeyUserID, pls.C.MustGet(flag.KeyUserID).(uint))
	speakerFollow, err := db.IsRelationshipList(ctx, lineSpeakerID)
	if err != nil {
		return nil, err
	}
	arrProfileResponse := make([]ProfileResponse, 0, len(pls.Users))
	for _, user := range pls.Users {
		arrProfileResponse = append(arrProfileResponse, ProfileResponse{
			Login:             user.Login,
			FirstName:         user.FirstName,
			Image:             user.Image,
			Bio:               user.Bio,
			NumberOfFollowers: user.NumberOfFollowers,
			CreatedAt:         user.CreatedAt.UTC().Format(vr.RFC3339Milli),
			Following:         speakerFollow[user.ID],
		})
	}
	return arrProfileResponse, nil
}

func (pls *ProfileListSerializer) userIDList() []uint {
	usersAlias := pls.Users
	n := len(usersAlias)
	arrUserID := make([]uint, n)
	for i := 0; i < n; i++ {
		arrUserID[i] = usersAlias[i].ID
	}
	return arrUserID
}
