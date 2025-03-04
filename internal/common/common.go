// common - пакет для общего испоьзования другими пакетами приложения
package common

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

const SecretKey = "qwert12345"

var incorrectToken = errors.New("token not valid")

// GenToken - генерирует токен для использования в Request header
func GenToken(id uint) (string, error) {
	claims := jwt.MapClaims{
		"id":          id,
		"exploretion": time.Now().Add(60 * time.Minute).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(SecretKey))
}

// Indeficator - для получения 'id' из jwt.MapClaims (для удобства)
func Indeficator(token *jwt.Token, key string) (int, error) {
	var value int
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return value, incorrectToken
	}
	value = int(claims[key].(float64))

	return value, nil
}

// ошибка для защиты от некоректного использования:
// getFieldName(structNamespace string) (string, error)
var incorrectStructNamespace = errors.New("incorrect struct namespace")

// FiledValidator - для создания кастомных функций обработки ошибок
// объекта validator.FieldError
type FiledValidator struct {
	validator.FieldError
}

// получении имени поля из validator.FieldError StructNamespace()
func (fv FiledValidator) FieldName() (string, error) {
	structNamespace := fv.StructNamespace()
	n := len(structNamespace)
	if n < 3 ||
		structNamespace[0] == '.' ||
		structNamespace[n-1] == '.' {
		return "", incorrectStructNamespace
	}
	for i := n - 1; i > -1; i-- {
		if structNamespace[i] == '.' {
			return structNamespace[i+1 : n], nil
		}
	}
	return "", incorrectStructNamespace
}

// ErrCommonUnexpectedType - маркировка ошибкок приведения типов
// во время использования 'func (c *Context) MustGet(key string) any'
var ErrCommonUnexpectedType = errors.New("model - unexpected type")

// CommonError - тип для более подробного описания ошибок
type CommonError struct {
	DataError map[string]any `json:"errors"`
}

// формат записи ошибки, обертывая в объект
func NewError(key string, err error) CommonError {
	storeErros := map[string]any{key: err.Error()}
	return CommonError{DataError: storeErros}
}

// обработка ошибок полученных во время выполнения 'context.Bind' из 'gin' framework
func NewDataErrorValidator(err error) CommonError {
	storeErros := make(map[string]any)
	dataErr := err.(validator.ValidationErrors)

	for _, fe := range dataErr {
		info := fe.Param()
		if len(info) == 0 {
			fv := FiledValidator{fe}
			info, _ = fv.FieldName()
		}
		storeErros[fe.Field()] = fmt.Sprintf("{%v:%v}", fe.Tag(), info)
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

// IsValidParam - проверяет наличие значение по ключу в 'gin.Context.Params'
func IsValidParam(c *gin.Context, key string) (string, bool) {
	val := c.Param(key)
	if len(val) == 0 {
		return "", false
	}
	return val, true
}
