package models

import (
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// UserProperty - свойсва для поиска списка пользователей
// см.'UserApprove' 'FindUserList(ctx context.Context, data any) ([]UserModel, error)'
type UserProperty struct {
	FirstName string
	LastName  string

	common.TimeRange

	common.LimitOffset
}

func (up UserProperty) IsFirstName() bool {
	return up.FirstName != ""
}

func (up UserProperty) IsLastName() bool {
	return up.LastName != ""
}
