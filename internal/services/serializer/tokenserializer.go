package serializer

import (
	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type TokenSerializer struct {
	C *gin.Context
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (ts *TokenSerializer) Response() (TokenResponse, error) {
	token, err := common.GenToken(ts.C.MustGet(flag.KeyUserID).(uint))
	return TokenResponse{Token: token}, err
}
