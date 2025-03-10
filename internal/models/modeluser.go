package models

import (
	"context"
	"errors"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"strconv"
	"time"
)

type UserModel struct {
	ID        uint
	Login     string
	Password  string
	FirstName string
	LastName  *string
	Phone     *string
	Email     string

	CreatedAt      time.Time
	UpdatedAT      *time.Time
	LastConnection *time.Time

	// users_acces table (0 1 2 3 4 5 6 7)
	// 1 - чтение, 2 - запись, 4 администратор без
	// остальные - производное cуммы
	// (4 + 1 + 2 = 7) - администратор с правами чтения и записи
	Access string
	// users_biography table
	Image *string
	Bio   *string //biography
}

// UserConnect - описывает регистрацию и подключение пользователя
type UserConnect interface {
	// SaveOneUser - запись данных пользователя и получение ногового, уникального ID
	SaveOneUser(ctx context.Context, user UserModel) (uint, error)

	// LoginUserWithUpdateTime - поиск по определенному полю,
	// определяемому с помощью 'flag' см. modeluser.go в текущем пакете 'source'
	// так же должна обновлять поле LastConnection *time.Time
	LoginUserWithUpdateTime(ctx context.Context, user UserModel, flag int) (UserModel, error)
}

// UserApprove - получение, обновление данных пользователя
type UserApprove interface {
	// FindOneUserByField - поиск пользователя по полую, определяемому с помощью 'flag' - см топ. данного фафла
	FindOneUserByField(ctx context.Context, user UserModel, falg int) (UserModel, error)

	NewDataUser(ctx context.Context, user UserModel) error
}

// UserFollowing - обрабатывает отношение пользователей
type UserFollowing interface {
	// NewRelationship(new following) - создает статус подписки для выбранных пользователей
	NewRelationship(ctx context.Context, userFollower, userSpeaker UserModel) error

	// IsRelationship(is following) - возращает наличие подписки 'userFolower' на 'userSpeaker'
	IsRelationship(ctx context.Context, userFollower, userSpeaker UserModel) (bool, error)

	// EndRelationship(delete following) - удаляет статус подписки для выбранных пользователей
	EndRelationship(ctx context.Context, userFollower, userSpeaker UserModel) error
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

// 'u' - получен из базы данных и имеет захешированный пароль
// checkPassword - сравнивает полученный пароль 'password' и пароль из 'UserMOdel'
func (u *UserModel) CheckPassword(password string) bool {
	hashPassword := common.HashData(password)
	return u.Password == hashPassword
}

// ErrSourceFlag - некорректный флага
var ErrSourceFlag = errors.New("unexpected flag")

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
