// common - пакет для общего испоьзования другими пакетами приложения
package common

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

const (
	SecretKey     = "qwert12345"
	TrickPassword = "not a password"
)

var ErrCommonTokenIncorrect = errors.New("token not valid")

// GenToken - генерирует токен для использования в Request header
func GenToken(id uint) (string, error) {
	claims := jwt.MapClaims{
		"id":          id,
		"exploretion": time.Now().Add(60 * time.Minute).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(SecretKey))
}

// IDFromToken - проверяет время действия токена и возвращает 'ID' пользователя
func IDFromToken(token *jwt.Token) (uint, error) {
	exploretion, err := Indeficator[float64](token, "exploretion")
	if err != nil {
		return 0, err
	}
	if int64(exploretion) < time.Now().Unix() {
		return 0, errors.New("token time expired")
	}
	user_id, err := Indeficator[float64](token, "id")
	if err != nil {
		return 0, err
	}
	return uint(user_id), nil
}

// Indeficator - для получения 'id' из jwt.MapClaims (для удобства)
func Indeficator[T any](token *jwt.Token, key string) (T, error) {
	var value T
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return value, ErrCommonTokenIncorrect
	}
	value, ok = claims[key].(T)
	if !ok {
		return value, ErrCommonTokenIncorrect
	}
	return value, nil
}

// ContextMiddleware - стартовая функция для создания context.WithTimeout
// и передачи ctx через 'c.Request'
func ContextMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// JSONWithContext - with pair middleware 'ContextMiddleware' -> (look up)
//
// context.DeadlineExceeded -> don't write object to Response, set code in 'ContextMiddleware'
func JSONWithContext(c *gin.Context, httpStatus int, obj any) {
	if c.Request.Context().Err() == context.DeadlineExceeded {
		c.AbortWithStatus(http.StatusRequestTimeout)
		return
	}
	c.JSON(httpStatus, obj)
}

// ошибка для защиты от некоректного использования:
// getFieldName(structNamespace string) (string, error)
var ErrCommonFiledValidatorIncorrect = errors.New("incorrect struct namespace")

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
		return "", ErrCommonFiledValidatorIncorrect
	}
	for i := n - 1; i > -1; i-- {
		if structNamespace[i] == '.' {
			return structNamespace[i+1 : n], nil
		}
	}
	return "", ErrCommonFiledValidatorIncorrect
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
	dataErr, ok := err.(validator.ValidationErrors)
	if !ok {
		storeErros["validator"] = ErrCommonUnexpectedType.Error()
		return CommonError{storeErros}
	}
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

// WhenEmptyStringThenNULL - для записи в базу данных значения - 'NULL' по заданным условиям
func WhenEmptyStringThenNULL(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{"", false}
	}
	return sql.NullString{*s, len(*s) != 0}
}

// strings.Trim(strings.Replace(fmt.Sprint([]int{1, 2, 3, 4}), " ", ",", -1), "[]")!!!!!!!
// ArrayToLineForQuery - формализует массив в строку для запросов типа 'IN (line)'
// возвращает строку для запроса и количесво записанных элеменов
// []string{"abc","def"} -> "'abc','def'"
func ArrayToLineForQuery(data []string) (string, int) {
	var buff bytes.Buffer
	n := len(data)
	// защита от пустого запроса в sql.DB
	// ... WHERE some IN('')
	if n == 0 {
		buff.Write([]byte{'\'', '\''})
		return buff.String(), n
	}
	for i := 0; i < n; i++ {
		buff.WriteByte('\'')
		buff.WriteString(data[i])
		buff.Write([]byte{'\'', ','})
	}
	if lenArr := buff.Len(); lenArr > 0 {
		buff.Truncate(lenArr - 1)
	}
	return buff.String(), n
}

// TimeRange - диапозон времени
type TimeRange struct {
	StartDate time.Time
	EndDate   time.Time
}

func (tr *TimeRange) IsRangeZero() bool {
	return tr.StartDate.IsZero() || tr.EndDate.IsZero()
}

// LimitOffset - характеристики для SQL query
// команды LIMIT number OFFSET number;
type LimitOffset struct {
	Limit  uint
	Offset uint
}
