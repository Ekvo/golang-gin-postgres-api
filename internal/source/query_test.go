package source

import (
	"context"
	"time"

	mod "github.com/Ekvo/golang-gin-postgres-api/internal/models"
)

var dbTest SQLSource

func NewUser() mod.UserModel {
	return mod.UserModel{}
}

var dataForTests = []struct {
	nameWithNumber string
	function       func(data any) (any, error)
	startData      any
	expectedData   any
	expectedErr    error
	msg            string
}{
	{
		nameWithNumber: `first test - 'save user'`,
		function: func(data any) (any, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			return dbTest.SaveOneUser(ctx, data)
		},
		startData: NewUser(),
	},
}
