// common - пакет для общего испоьзования другими пакетами приложения
package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const SecretKey = "qwert12345"

// генерирует токен для использования в Request header
func GenToken(id int) (string, error) {
	claims := jwt.MapClaims{
		"id":          id,
		"exploretion": time.Now().Add(60 * time.Minute).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(SecretKey))
}

// CommonError - тип для более подробного описания ошибок
type CommonError struct {
	DataError map[string]any `json:"errors"`
}

const unknown = "unknown"

// обработка ошибок полученных во время выполнения 'context.Bind' из 'gin' framework
func NewDataErrorValidator(err error) CommonError {
	storeErros := make(map[string]any)
	dataErr := err.(validator.ValidationErrors)

	for _, fe := range dataErr {
		param := fe.Param()
		if len(param) == 0 {
			param = unknown
		}
		storeErros[fe.Field()] = fmt.Sprintf("{%v:%v}", fe.Tag(), param)
	}
	return CommonError{storeErros}
}

// для более подробной обработки ошибок(error)
// c.MustBindWith() ->  c.ShouldBindWith()
func Bind(c *gin.Context, obj any) error {
	bind := binding.Default(c.Request.Method, c.ContentType())
	return c.ShouldBindWith(obj, bind)
}

func HashData(line string) string {
	hashLine := sha256.Sum256([]byte(line))
	return hex.EncodeToString(hashLine[:])
}
