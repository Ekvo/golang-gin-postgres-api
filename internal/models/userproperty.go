package models

import (
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"time"
)

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
