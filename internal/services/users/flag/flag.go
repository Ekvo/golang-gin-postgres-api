package flag

import (
	"errors"
	"strconv"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
)

// ErrSourceFlag - некорректный флага
var ErrServicesSourceFlag = errors.New("unexpected flag")

// ключи для записи в 'gin.Context.Keys'
const (
	// храним в формате 'uint'
	KeyUserID = "user_id"

	// храним в формате 'string'
	KeyUserAccess = "user_access"

	// храним в формате 'UserModel'
	KeyUserModel = "user_model"

	// храним в формате 'int'
	KeyFlagFiled = "flag_filed"
)

// маркеры для поиска по полю 'UserModel'
const (
	FlagID = iota + 1
	FlagLogin
	FlagPhone
	FlagEmail
)

// хранение имени полей из 'UserMOdel'
var flagsName = []string{"unknown", "id", "login", "phone", "email"}

func FlagName(flag int) string {
	if flag < 1 || flag >= len(flagsName) {
		flag = 0
	}
	return flagsName[flag]
}

// findByField - определяет поле и параметр из 'UserModel' для поиска в базе данных
func FindByFieldWithKey(user models.UserModel, flag int) (string, string, error) {
	field, param := FlagName(flag), ""
	if field == "unknown" {
		return "", "", ErrServicesSourceFlag
	}
	switch flag {
	case FlagID:
		param = strconv.Itoa(int(user.ID))
	case FlagLogin:
		param = user.Login
	case FlagPhone:
		if user.Phone != nil {
			param = *user.Phone
		}
	case FlagEmail:
		param = user.Email
	default:
		return "", "", ErrServicesSourceFlag
	}
	if len(param) < 1 {
		return "", "", errors.New("can't find by empty param")
	}
	return field, param, nil
}
