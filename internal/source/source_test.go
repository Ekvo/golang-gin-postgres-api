package source

import (
	"context"
	"fmt"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"log"
	"net"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
)

var (
	timeCreate         = time.Now().UTC()
	timeUpdate         = timeCreate.Add(2 * time.Minute)
	timeLastConnection = timeCreate.Add(time.Minute)

	lastName  = "Create"
	phone     = "+79123456789"
	imageURl  = "https://s3.stroi-news.ru/img/klassnie-kartinki-dlya-devochek-4.jpg"
	biography = "Hi, I'm Alex Test and I love cats.)."
)

func NewUser() models.UserModel {
	return models.UserModel{
		Login:     "alex",
		Password:  "qwer1234",
		Access:    "0",
		FirstName: "Alex",
		LastName:  &lastName,
		Phone:     &phone,
		Email:     "alex@mail.com",
		Image:     &imageURl,
		Bio:       &biography,
		CreatedAt: timeCreate,
	}
}

func UpdateUser() models.UserModel {
	lastName = "Update"
	imageURl = "https://s3.stroi-news.ru/img/klassnie-kartinki-dlya-devochek-4.jpg"
	biography = `Hi, I'm Alex and I love cats, dogs.
I have a cat Jerry, he's a fighter and you can't pet him, but as soon as he gets hungry,
he becomes affectionate and his velvety fur immediately runs to my feet.`
	return models.UserModel{
		ID:        1,
		Login:     "alexMiu",
		Password:  "qwer1234",
		Access:    "4",
		FirstName: "Alex",
		LastName:  &lastName,
		Email:     "alex@mail.com",
		Image:     &imageURl,
		Bio:       &biography,
		UpdatedAt: &timeUpdate,
	}
}

func initDBForTest() (*pgxpool.Pool, error) {
	if err := godotenv.Load("../../.env"); err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_TEST_USER"),
		os.Getenv("DB_TEST_PASSWORD"),
		os.Getenv("HOST_TEST"),
		os.Getenv("DB_TEST_PORT"),
		os.Getenv("DB_TEST_NAME"),
		os.Getenv("DB_TEST_SSLMODE"),
	)
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	config.MaxConns = int32(5)
	config.MinConns = int32(1)
	config.MaxConnLifetime = 2 * time.Hour
	config.MaxConnIdleTime = 10 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.ConnConfig.ConnectTimeout = 1 * time.Second
	config.ConnConfig.DialFunc = (&net.Dialer{
		KeepAlive: config.HealthCheckPeriod,
		Timeout:   config.ConnConfig.ConnectTimeout,
	}).DialContext

	return pgxpool.NewWithConfig(context.Background(), config)
}

var qq = []struct {
	description  string
	init         func(ctx context.Context, pool SQLSource, data any) (any, error)
	ctxKeyVal    []any //[0]-key; [1]-value
	ctxTimeOut   time.Duration
	startData    any
	expectedData any
	expectedErr  error
	msg          string
}{
	{ //1
		description: `Create tables 'users' & 'followers' - valid `,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.NewTable(ctx, tableUsers, tableFollowers)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    nil,
		expectedData: nil,
		expectedErr:  nil,
		msg:          "valid query create tables",
	},
	{ //2
		description: `Save one user - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return pool.SaveOneUser(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    NewUser(),
		expectedData: uint(1),
		expectedErr:  nil,
		msg:          "valid query create new user with ID = 1",
	},
	{ //3
		description: `Login user - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return pool.LoginUserWithUpdateTime(ctx, data)
		},
		ctxKeyVal:  nil,
		ctxTimeOut: 100 * time.Second,
		startData: models.UserModel{
			Login:          "alex",
			LastConnection: &timeLastConnection,
		},
		expectedData: models.UserModel{
			ID:       1,
			Password: "qwer1234",
		},
		expectedErr: nil,
		msg:         "valid query login and return ID,Password",
	},
	{ //4
		description: `Login user - invalid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return pool.LoginUserWithUpdateTime(ctx, data)
		},
		ctxKeyVal:  nil,
		ctxTimeOut: 100 * time.Second,
		startData: models.UserModel{
			Login:          "alien",
			LastConnection: &timeLastConnection,
		},
		expectedData: models.UserModel{},
		expectedErr:  ErrSourceTransaction,
		msg:          "invalid login, err - \"resource not found\"",
	},
	{ //5
		description: `Update user - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.NewDataUser(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    UpdateUser(),
		expectedData: nil,
		expectedErr:  nil,
		msg:          "valid update user error is nil",
	},
	{ //6
		description: `Find user - invalid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return pool.FindOneUserByField(ctx, data)
		},
		ctxKeyVal:    []any{flag.KeyFlagFiled, flag.FlagLogin},
		ctxTimeOut:   100 * time.Second,
		startData:    models.UserModel{Login: "alien"},
		expectedData: models.UserModel{},
		expectedErr:  pgx.ErrNoRows,
		msg:          "invalid find user, return err - \"resource not found\"",
	},
	{ //7
		description: `is follow - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return pool.IsRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1, 1},
		expectedData: true,
		expectedErr:  nil,
		msg:          "valid relationship, return 'true' alexMiu subscribe on alexMiu",
	},
	{ //8
		description: `is follow - invalid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return pool.IsRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{2, 1},
		expectedData: false,
		expectedErr:  nil,
		msg:          "invalid - check follow, return err - \"resource not found\"",
	},
	{ //9
		description: `is follow - invalid (startData)`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return pool.IsRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1},
		expectedData: false,
		expectedErr:  ErrSourceRelationship,
		msg:          "invalid - array, err - \"is impossible - update or get Relationship with current data\"",
	},
	{ //10
		description: `remove follow - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.EndRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1, 1},
		expectedData: nil,
		expectedErr:  nil,
		msg:          "valid  unsubscribe, no error",
	},
	{ //11
		description: `remove follow - invalid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.EndRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1, 1},
		expectedData: nil,
		expectedErr:  ErrSourceTransaction,
		msg:          "invalid  unsubscribe, err - \"resource not found\"",
	},
	{ //12
		description: `remove follow - invalid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.EndRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1},
		expectedData: nil,
		expectedErr:  ErrSourceRelationship,
		msg:          "invalid array, err - \"is impossible - update or get Relationship with current data\"",
	},
	{ //13
		description: `new subscribe - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.NewRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1, 1},
		expectedData: nil,
		expectedErr:  nil,
		msg:          "valid - subscribe, alexMiu is now subscribe on alexMiu",
	},
	{ //14
		description: `new subscribe - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.NewRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1, 1},
		expectedData: nil,
		expectedErr:  ErrSourceTransaction,
		msg:          "invalid  subscribe, alexMiu already subscribe on alexMiu",
	},
	{ //15
		description: `new subscribe - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.NewRelationship(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    []uint{1},
		expectedData: nil,
		expectedErr:  ErrSourceRelationship, //!!!!!
		msg:          "invalid array, err - \"is impossible - update or get Relationship with current data\"",
	},
	{ //16
		description: `Delete user - valid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.RemoveUser(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    uint(1),
		expectedData: nil,
		expectedErr:  nil,
		msg:          "valid  deleted, alexMiu now not on 'users'",
	},
	{ //17
		description: `Delete user - invalid`,
		init: func(ctx context.Context, pool SQLSource, data any) (any, error) {
			return nil, pool.RemoveUser(ctx, data)
		},
		ctxKeyVal:    nil,
		ctxTimeOut:   100 * time.Second,
		startData:    uint(1),
		expectedData: nil,
		expectedErr:  ErrSourceTransaction,
		msg:          "valid  deleted, alexMiu - \"resource not found\" ",
	},
}

func TestUserSource(t *testing.T) {
	pool, err := initDBForTest()
	if err != nil {
		log.Fatalf("source_test: pgxpool error - %v", err)
	}
	defer pool.Close()
	base := NewSQLSource(pool)
	ctx := context.Background()
	// delete tables
	base.pTx.Pool.Exec(ctx, `
DROP TABLE IF EXISTS users,followers,articles,articles_tags,article_favorite,comments,tags;`)

	asserts := assert.New(t)

	for i, query := range qq {
		log.Printf("\t%d query: %s", i+1, query.description)

		ctx, cancel := context.WithTimeout(ctx, query.ctxTimeOut)
		defer cancel()
		if query.ctxKeyVal != nil {
			ctx = context.WithValue(ctx, query.ctxKeyVal[0], query.ctxKeyVal[1])
		}

		res, err := query.init(ctx, base, query.startData)

		asserts.ErrorIs(err, query.expectedErr, "Errors don't match "+query.msg)
		asserts.EqualValues(query.expectedData, res, "Result don't match"+query.msg)
	}
}

func FindUserData() models.UserModel {
	return models.UserModel{
		ID:                1,
		Login:             "alexMiu",
		Password:          "qwer1234",
		Access:            "4",
		FirstName:         "Alex",
		LastName:          &lastName,
		Email:             "alex@mail.com",
		Image:             &imageURl,
		Bio:               &biography,
		CreatedAt:         timeCreate,
		UpdatedAt:         &timeUpdate,
		LastConnection:    nil,
		NumberOfFollowers: 1,
	}
}

var userFindTeasData = []struct {
	description  string
	ctxKeyVal    []any //[0]-key; [1]-value
	ctxTimeOut   time.Duration
	startData    models.UserModel
	expectedData models.UserModel
	expectedErr  error
	msg          string
}{
	{
		description:  "find user by ID",
		ctxKeyVal:    []any{flag.KeyFlagFiled, flag.FlagID},
		ctxTimeOut:   100 * time.Second,
		startData:    models.UserModel{ID: 1},
		expectedData: FindUserData(),
		expectedErr:  nil,
		msg:          "valid find - all is empty, no error",
	},
	{
		description:  "wrong find user",
		ctxKeyVal:    []any{flag.KeyFlagFiled, flag.FlagLogin},
		ctxTimeOut:   100 * time.Second,
		startData:    models.UserModel{Login: "predator"},
		expectedData: models.UserModel{},
		expectedErr:  pgx.ErrNoRows,
		msg:          "invalid find - ans empty with error - no eows",
	},
}

// FindOneUserByField - there are pointers and time.Time
//
// need unique test or table test will become more complicated
func TestSQLSource_FindOneUserByField(t *testing.T) {
	pool, err := initDBForTest()
	if err != nil {
		log.Fatalf("source_test: pgxpool error - %v", err)
	}
	defer pool.Close()
	base := NewSQLSource(pool)
	ctx := context.Background()

	// delete tables
	base.pTx.Pool.Exec(ctx, `
DROP TABLE IF EXISTS users,followers,articles,articles_tags,article_favorite,comments,tags;`)

	asserts := assert.New(t)
	requires := require.New(t)

	err = base.NewTable(ctx, tableUsers, tableFollowers)
	requires.NoError(err, "New Table err - should be nil")

	id, err := base.SaveOneUser(ctx, NewUser())
	requires.Equal(uint(1), id, "ID should be equal 1")
	requires.NoError(err, "Create user err - should be nil")

	err = base.NewDataUser(ctx, UpdateUser())
	requires.NoError(err, "Update user err - should be nil")

	for i, query := range userFindTeasData {
		log.Printf("\t%d query: %s", i+1, query.description)

		ctx, cancel := context.WithTimeout(ctx, query.ctxTimeOut)
		defer cancel()
		ctx = context.WithValue(ctx, query.ctxKeyVal[0], query.ctxKeyVal[1])

		user, err := base.FindOneUserByField(ctx, query.startData)
		asserts.ErrorIs(err, query.expectedErr, "Errors don't match "+query.msg)

		expectedUser := query.expectedData
		asserts.Equal(expectedUser.ID, user.ID, "ID")
		asserts.Equal(expectedUser.Login, user.Login, "login")
		asserts.Equal(expectedUser.Password, user.Password, "password")
		asserts.Equal(expectedUser.Access, user.Access, "access")
		asserts.Equal(expectedUser.FirstName, user.FirstName, "firstName")

		if expectedUser.LastName != nil {
			requires.NotNil(user.LastName, "expected lastName - is nil")
			asserts.Equal(*expectedUser.LastName, *user.LastName, "lastName")
		} else {
			asserts.Nil(user.LastName, "res lastName - not nil")
		}
		if expectedUser.Phone != nil {
			requires.NotNil(user.Phone, "expected Phone - is nil")
			asserts.Equal(*expectedUser.Phone, *user.Phone, "Phone")
		} else {
			asserts.Nil(user.Phone, "res Phone - not nil")
		}
		asserts.Equal(expectedUser.Email, user.Email, "email")
		if expectedUser.Image != nil {
			requires.NotNil(user.Image, "expected Image - is nil")
			asserts.Equal(*expectedUser.Image, *user.Image, "Image")
		} else {
			asserts.Nil(user.Image, "res Image - not nil")
		}
		if expectedUser.Bio != nil {
			requires.NotNil(user.Bio, "expected Bio - is nil")
			asserts.Equal(*expectedUser.Bio, *user.Bio, "Bio")
		} else {
			asserts.Nil(user.Bio, "res Bio - not nil")
		}
		//postgres safe time .10^6 - time.StampMicro -> 1000
		asserts.WithinDuration(expectedUser.CreatedAt.UTC(), user.CreatedAt.UTC(), 1000)
		if expectedUser.UpdatedAt != nil {
			requires.NotNil(user.UpdatedAt, "expected UpdatedAt - is nil")
			asserts.WithinDuration(expectedUser.UpdatedAt.UTC(), user.UpdatedAt.UTC(), 1000)
		} else {
			asserts.Nil(user.UpdatedAt, "res UpdatedAt - not nil")
		}
		if expectedUser.LastConnection != nil {
			requires.NotNil(user.LastConnection, "expected LastConnection - is nil")
			asserts.WithinDuration(expectedUser.LastConnection.UTC(), user.LastConnection.UTC(), 1000)
		} else {
			asserts.Nil(user.LastConnection, "res LastConnection - not nil")
		}

		asserts.Equal(expectedUser.NumberOfFollowers, user.NumberOfFollowers, "followers")
	}
}

var (
	loginList = []string{"alex", "alien", "predator", "Wayland"}
	email     = "@gmail.com"
)

func NewUserForList(login, email string, createdAt time.Time) models.UserModel {
	return models.UserModel{
		Login:             login,
		Password:          "qwer1234",
		Access:            "4",
		FirstName:         login,
		Email:             email,
		CreatedAt:         createdAt,
		NumberOfFollowers: 1, // no use on SQLquery -> for compare in tests
	}
}

// in in expectedRes add users lowercase by login - "aa","aaa","aab"
var userPropertyList = []struct {
	description string
	ctxTimeOut  time.Duration
	property    models.UserProperty
	expectedRes []models.UserModel
	msg         string
}{
	{
		description: "find only -> alex",
		ctxTimeOut:  100 * time.Second,
		property:    models.UserProperty{FirstName: "alex"},
		expectedRes: []models.UserModel{
			NewUserForList("alex", "alex"+email, timeCreate),
		},
		msg: "found one User with name 'alex'",
	},
	{
		description: "zero list",
		ctxTimeOut:  100 * time.Second,
		property:    models.UserProperty{FirstName: "someuser"},
		expectedRes: []models.UserModel{},
		msg:         "no result",
	},
	{
		description: "limit, offsett -> AVP",
		ctxTimeOut:  100 * time.Second,
		property: models.UserProperty{
			LimitOffset: common.LimitOffset{
				Limit:  2,
				Offset: 1,
			},
		},
		expectedRes: []models.UserModel{
			NewUserForList("alien", "alien"+email, timeCreate.Add(time.Hour)),
			NewUserForList("predator", "predator"+email, timeCreate.Add(time.Hour*2)),
		},
		msg: "found two Users with name 'alien','predator'",
	},
	{
		description: "time range, limit, offsett -> Wayland",
		ctxTimeOut:  100 * time.Second,
		property: models.UserProperty{
			TimeRange: common.TimeRange{
				StartDate: timeCreate.Add(time.Hour),
				EndDate:   timeCreate.Add(time.Hour * 4),
			},
			LimitOffset: common.LimitOffset{
				Limit:  1,
				Offset: 2,
			},
		},
		expectedRes: []models.UserModel{
			NewUserForList("Wayland", "Wayland"+email, timeCreate.Add(time.Hour*3)),
		},
		msg: "found one User with name 'Wayland'",
	},
}

func TestSQLSource_FindUserList(t *testing.T) {
	pool, err := initDBForTest()
	if err != nil {
		log.Fatalf("source_test: pgxpool error - %v", err)
	}
	defer pool.Close()
	base := NewSQLSource(pool)
	ctx := context.Background()

	// delete tables
	base.pTx.Pool.Exec(ctx, `
DROP TABLE IF EXISTS users,followers,articles,articles_tags,article_favorite,comments,tags;`)

	asserts := assert.New(t)
	requires := require.New(t)

	err = base.NewTable(ctx, tableUsers, tableFollowers)
	requires.NoError(err, "create table error")

	//expectedUserList := make([]models.UserModel, len(loginList))
	for i, login := range loginList {
		//expectedUserList[i] = NewUserForList(login, login+email, timeCreate.Add(time.Hour*time.Duration(i)))
		_, err := base.SaveOneUser(ctx, NewUserForList(login, login+email, timeCreate.Add(time.Hour*time.Duration(i))))
		requires.NoError(err, fmt.Sprintf("create users step %d error", i))
		//expectedUserList[i].ID = id
	}

	for i, property := range userPropertyList {
		log.Printf("\t%d test: %s\n", i+1, property.description)
		ctx, cancel := context.WithTimeout(ctx, property.ctxTimeOut)
		defer cancel()

		resUserList, err := base.FindUserList(ctx, property.property)
		requires.NoError(err, "Error "+property.msg)
		requires.Equal(len(property.expectedRes), len(resUserList), "Dif lenght array "+property.msg)

		sort.Slice(resUserList, func(i, j int) bool { return resUserList[i].Login < resUserList[j].Login })

		for i, expUser := range property.expectedRes {
			asserts.Equal(expUser.Login, resUserList[i].Login, "login")
			asserts.Equal(expUser.Password, resUserList[i].Password, "password")
			asserts.Equal(expUser.Access, resUserList[i].Access, "access")
			asserts.Equal(expUser.Email, resUserList[i].Email, "email")
			asserts.WithinDuration(expUser.CreatedAt, resUserList[i].CreatedAt, 1000)
			asserts.Equal(expUser.NumberOfFollowers, resUserList[i].NumberOfFollowers, "follower")
		}
	}
}
