package users

import (
	"errors"
	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"time"
)

type UserModel struct {
	ID             uint
	Login          string
	Password       string
	FirstName      string
	LastName       string
	Phone          string
	Email          string
	Access         string
	Image          string
	Bio            string //biography
	CreatedAt      time.Time
	UpdatedAT      time.Time
	LastConnection time.Time
}

func (u *UserModel) hashPassword() error {
	if err := u.validPassword(); err != nil {
		return err
	}
	u.Password = common.HashData(u.Password)
	return nil
}

const (
	minLenghtPassword = 8
	maxLenghtPassword = 255
)

func (u *UserModel) validPassword() error {
	lenghtPassword := len(u.Password)
	if minLenghtPassword > lenghtPassword ||
		lenghtPassword > maxLenghtPassword {
		return errors.New("incorrect lenght of password")
	}
	return nil
}
