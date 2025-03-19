package models

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// alias - для удобсва
type UM = UserModel

type UserModel struct {
	// users table
	ID        uint
	Login     string
	Password  string
	FirstName string
	LastName  *string
	Phone     *string
	Email     string

	CreatedAt      time.Time
	UpdatedAt      *time.Time
	LastConnection *time.Time

	// users_acces table (0 1 2 3 4 5 6 7)
	// 1 - чтение, 2 - запись, 4 администратор без
	// остальные - производное cуммы
	// (4 + 1 + 2 = 7) - администратор с правами чтения и записи
	Access string

	// users_biography table
	Image *string
	Bio   *string //biography

	// followers table
	// Количесво уникальных подписчиков
	NumberOfFollowers uint
}

// UserConnect - описывает регистрацию и подключение пользователя
type UserConnect interface {
	// SaveOneUser - запись данных пользователя и получение ногового, уникального ID
	SaveOneUser(ctx context.Context, data any) (uint, error)

	// LoginUserWithUpdateTime - поискользователя,
	// так же должна обновлять поле LastConnection *time.Time
	LoginUserWithUpdateTime(ctx context.Context, data any) (UserModel, error)
}

// UserApprove - получение, обновление данных пользователя
type UserApprove interface {
	// FindOneUserByField - получение данных пользователя
	//
	// идея - передать данные с определенным флагом, для поиска по заданному имени столбца в базе данных
	FindOneUserByField(ctx context.Context, data any) (UserModel, error)

	// FindUserList - поиск пользователей по заданным параметрам переданным через 'data'
	FindUserList(ctx context.Context, data any) ([]UserModel, error)

	// NewDataUser - обновление данных пользователя
	NewDataUser(ctx context.Context, data any) error
}

// UserFollowing - обрабатывает отношение пользователей
type UserFollowing interface {
	// NewRelationship(new following) - создает статус подписки для выбранных пользователей
	NewRelationship(ctx context.Context, data any) error

	// IsRelationship(is following) - возращает наличие подписки 'userFolower' на 'userSpeaker'
	IsRelationship(ctx context.Context, data any) (bool, error)

	// IsRelationshipList - возвращает список профилей пользователей со статусом подписки на каждого
	// относительно пользователя сделавшего запрос
	// возвращает map[uint]bool - ключ speakerID, значение - статус наличия подписки
	IsRelationshipList(ctx context.Context, data any) (map[uint]bool, error)

	// EndRelationship(delete following) - удаляет статус подписки для выбранных пользователей
	EndRelationship(ctx context.Context, data any) error
}

// UserApproveAndFollowing - групирует методы связанные с обработкой зарегистрированных пользователей
type UserApproveFollowing interface {
	UserApprove
	UserFollowing
}

func (u *UserModel) HashPassword() error {
	if err := u.ValidPassword(); err != nil {
		return err
	}
	u.Password = common.HashData(u.Password)
	return nil
}

const (
	minLenghtPassword = 8
	maxLenghtPassword = 255
)

func (u *UserModel) ValidPassword() error {
	lenghtPassword := len(u.Password)
	if minLenghtPassword > lenghtPassword ||
		lenghtPassword > maxLenghtPassword {
		return errors.New("incorrect lenght of password")
	}
	return nil
}

var ErrModelsPassword = errors.New("invalid password")

// 'u' - получен из базы данных и имеет захешированный пароль
// checkPassword - сравнивает полученный пароль 'password' и пароль из 'UserMOdel'
func (u *UserModel) CheckPassword(password string) bool {
	hashPassword := common.HashData(password)
	return u.Password == hashPassword
}

// UserProperty - свойсва для поиска списка пользователей
// см.'UserApprove' 'FindUserList(ctx context.Context, data any) ([]UserModel, error)'
type UserProperty struct {
	FirstName string
	LastName  string
	common.TimeRange
	common.LimitOffset
}

func (up *UserProperty) NoEmpty() bool {
	// время invaild
	if up.IsRangeZero() {
		up.StartDate = time.Time{}
	}
	return len(up.FirstName) > 0 ||
		len(up.LastName) > 0 ||
		!up.TimeRange.StartDate.IsZero() ||
		up.Limit != 0 ||
		up.Offset != 0
}

// ErrSourceFlag - некорректный флага
var ErrSourceFlag = errors.New("unexpected flag")

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
func FindByFieldWithKey(user UserModel, flag int) (string, string, error) {
	field, param := FlagName(flag), ""
	if field == "unknown" {
		return "", "", ErrSourceFlag
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
		return "", "", ErrSourceFlag
	}
	if len(param) < 1 {
		return "", "", errors.New("can't find by empty param")
	}
	return field, param, nil
}
