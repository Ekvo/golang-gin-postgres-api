package models

import (
	"context"
	"errors"
	"time"

	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// alias - for comfortable
type UM = UserModel

type UserModel struct {
	ID uint

	Login    string
	Password string

	FirstName string
	LastName  *string

	Phone *string
	Email string

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

// UserDelete - удаление пользователя
type UserDelete interface {
	RemoveUser(ctx context.Context, data any) error
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

var ErrModelsUserInvalidPassword = errors.New("invalid password")

func (u *UserModel) HashPassword() {
	u.Password = common.HashData(u.Password)
}

// 'u' - получен из базы данных и имеет захешированный пароль
// checkPassword - сравнивает полученный пароль 'password' и пароль из 'UserMOdel'
func (u *UserModel) CheckPassword(password string) bool {
	hashPassword := common.HashData(password)
	return u.Password == hashPassword
}
