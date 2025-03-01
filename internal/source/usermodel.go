package source

import (
	"errors"
	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"time"
)

var incorrectAccess = errors.New("incorrect access data")

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

	// users_acces table
	Access string //0 1 2 3 4 5 6 7
	// users_biography table
	Image *string
	Bio   *string //biography
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
