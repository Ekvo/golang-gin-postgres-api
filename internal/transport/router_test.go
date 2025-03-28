package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mod "github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/users"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type StoreMock struct {
	userID uint
	//user key - login, value UserModel
	store map[string]*mod.UserModel

	// user exist by ID
	storeID map[uint]*mod.UserModel

	// key - speaker, value - foolowers
	foolowers map[uint][]uint
}

func NewStoreMock() *StoreMock {
	store := make(map[string]*mod.UserModel)
	storeID := make(map[uint]*mod.UserModel)
	foolowers := make(map[uint][]uint)
	return &StoreMock{
		userID:    0,
		store:     store,
		storeID:   storeID,
		foolowers: foolowers,
	}
}

func (s *StoreMock) SaveOneUser(ctx context.Context, data any) (id uint, err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	user := data.(mod.UserModel)
	if _, ex := s.store[user.Login]; ex {
		return id, source.ErrSourceAlreadyExists
	}
	s.userID++
	user.ID = s.userID
	s.store[user.Login] = &user
	s.storeID[user.ID] = &user
	s.foolowers[user.ID] = append(s.foolowers[user.ID], user.ID)
	return user.ID, nil
}

func (s *StoreMock) LoginUserWithUpdateTime(ctx context.Context, data any) (usr mod.UserModel, err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	userLogin := data.(mod.UserModel)
	if user, ex := s.store[userLogin.Login]; ex {
		if !user.CheckPassword(userLogin.Password) {
			err = mod.ErrModelsPassword
			return
		}
		user.LastConnection = userLogin.LastConnection
		return *user, nil
	}
	err = source.ErrSourceNotFound
	return
}

func (s *StoreMock) FindOneUserByField(ctx context.Context, data any) (usr mod.UserModel, err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	userLogin := data.(mod.UserModel)
	flag := ctx.Value(mod.KeyFlagFiled).(int)
	var user *mod.UserModel = nil
	ex := false
	if flag == mod.FlagLogin {
		user, ex = s.store[userLogin.Login]
	} else if flag == mod.FlagID {
		user, ex = s.storeID[userLogin.ID]
	} else {
		err = errors.New("mosk flag find user")
	}
	if ex {
		user.NumberOfFollowers = uint(len(s.foolowers[user.ID]))
		return *user, nil
	}
	err = source.ErrSourceNotFound
	return
}

func (s *StoreMock) FindUserList(ctx context.Context, data any) (usrl []mod.UserModel, err error) {
	defer func() {
		err = ctx.Err()
	}()
	for _, user := range s.store {
		usrl = append(usrl, *user)
	}
	return
}

func (s *StoreMock) NewDataUser(ctx context.Context, data any) (err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	newUser := data.(mod.UserModel)
	if oldUser, ex := s.storeID[newUser.ID]; ex {
		delete(s.store, oldUser.Login)
		newUser.LastConnection = oldUser.LastConnection
		s.storeID[oldUser.ID] = &newUser
		s.store[newUser.Login] = &newUser
		return
	}
	return source.ErrSourceNotFound
}

func (s *StoreMock) NewRelationship(ctx context.Context, data any) (err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return source.ErrSourceRelationship
	}
	fol, speak := followerSpeaker[0], followerSpeaker[1]
	_, exF := s.storeID[fol]
	_, exS := s.storeID[speak]
	if !exF || !exS {
		return source.ErrSourceNotFound
	}
	for _, follower := range s.foolowers[speak] {
		if follower == fol {
			return source.ErrSourceAlreadyExists
		}
	}
	s.foolowers[speak] = append(s.foolowers[speak], fol)
	return
}

func (s *StoreMock) IsRelationship(ctx context.Context, data any) (ans bool, err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return false, source.ErrSourceRelationship
	}
	fol, speak := followerSpeaker[0], followerSpeaker[1]
	_, exF := s.storeID[fol]
	_, exS := s.storeID[speak]
	if !exF || !exS {
		return false, source.ErrSourceNotFound
	}
	for _, follower := range s.foolowers[speak] {
		if follower == fol {
			return true, nil
		}
	}
	return false, nil
}

func (s *StoreMock) IsRelationshipList(ctx context.Context, data any) (speaker map[uint]bool, err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	lineSpeakerID := data.(string)
	followerID := ctx.Value(mod.KeyUserID).(uint)
	arrSpeakerID := getArrUint(lineSpeakerID)
link:
	for _, sp := range arrSpeakerID {
		for _, f := range s.foolowers[sp] {
			if f == followerID {
				speaker[sp] = true
				continue link
			}
		}
	}
	return
}

func getArrUint(line string) []uint {
	n := len(line)
	arr := []uint{}
	for i := 0; i < n; i++ {
		if '0' <= line[i] && line[i] <= '9' {
			start := i
			for ; i <= n; i++ {
				if i == n || '0' > line[i] || line[i] > '9' {
					num, _ := strconv.Atoi(line[start:i])
					arr = append(arr, uint(num))
					break
				}
			}
		}
	}
	return arr
}

func (s *StoreMock) EndRelationship(ctx context.Context, data any) (err error) {
	defer func() {
		if err == nil {
			err = ctx.Err()
		}
	}()
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return source.ErrSourceRelationship
	}
	fol, speak := followerSpeaker[0], followerSpeaker[1]
	_, exF := s.storeID[fol]
	_, exS := s.storeID[speak]
	if !exF || !exS {
		return source.ErrSourceNotFound
	}
	for i, follower := range s.foolowers[speak] {
		if follower == fol {
			s.foolowers[speak] = append(s.foolowers[speak][:i], s.foolowers[speak][i+1:]...)
			return nil
		}
	}
	return source.ErrSourceNotFound
}

const (
	regexpJWT = `[a-zA-Z0-9-_.]{125}`
	//2025-10-22T12:00:01.001Z or 2025-10-22T12:00:01.001+00:00
	regexpDateTimeMilli = `\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))`
)

var dataForTests = []struct {
	url               string
	method            string
	body              string
	withAuthorization bool
	expectedCode      int
	expectedBody      string
	msg               string
}{
	{
		url:    "/zephyr/connect/signup",
		method: http.MethodPost,
		body: `{"user_update":{
"login":"ekv",
"password":"qwert12345",
"first_name":"Alexander",
"last_name":"",
"access":"4",
"phone":"+79012345678",
"email":"alexander@mail.com",
"biography":"Hello!"}}`,
		expectedCode: http.StatusCreated,
		expectedBody: `{"approve":{"token":"[a-zA-Z0-9-_.]{125}"}}`,
		msg:          "status = 201 and body with token",
	},
	{
		url:    "/zephyr/connect/signup",
		method: http.MethodPost,
		body: `{"user_update":{
"login":"","password":"",
"first_name":"",
"last_name":"123",
"access":"5",
"phone":"ghf",
"email":"iii",
"image":"1234567",
"biography":""}}`,
		expectedCode: http.StatusUnprocessableEntity,
		expectedBody: `{"errors":{
"Access":"{excludesall:56789}",
"Email":"{email:Email}",
"FirstName":"{required:FirstName}",
"Image":"{url:Image}",
"LastName":"{alpha:LastName}",
"Login":"{required:Login}",
"Password":"{required:Password}",
"Phone":"{e164:Phone}"}}`,
		msg: "status = 422 with list of errors - 8 lines message + 1 line \"errors\" ",
	},
	{
		url:          "/zephyr/connect/login",
		method:       http.MethodPost,
		body:         `{"user_connect_with_login":{"login":"ekv","password":"qwert12345"}}`,
		expectedCode: http.StatusOK,
		expectedBody: `{"approve":{"token":"[a-zA-Z0-9-_.]{125}"}}`,
		msg:          "status = 200 with token",
	},
	{
		url:          "/zephyr/connect/login",
		method:       http.MethodPost,
		body:         `{"user_connect_with_login":{"login":"alien","password":"11001110101102"}}`, //1100111010110 - ай
		expectedCode: http.StatusNotFound,
		expectedBody: `{"login":"data with current param not found"}`,
		msg:          "status = 404 user - not found",
	},
	{
		url:               "/zephyr/user/",
		method:            http.MethodGet,
		withAuthorization: true,
		expectedCode:      http.StatusOK,
		expectedBody: `{"user":{
"login":"ekv",
"first_name":"Alexander",
"phone":"\+79012345678",
"email":"alexander@mail.com",
"access":"4",
"biography":"Hello!",
"created_at":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"last_connect":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))"}}`,
		msg: "status = 200 with all user fileds witout password",
	},
	{
		url:    "/zephyr/user/",
		method: http.MethodPut,
		body: `{"user_update":{
"login":"ekvo",
"password":"12345qwert",
"first_name":"Alexander",
"last_name":"Burlin",
"Image":"https://avatars.fastly.steamstatic.com/3c52f5b49925abe74105c49e0f6eb368b5149e7c_full.jpg", 
"access":"4",
"phone":"+79876543210",
"email":"alexan@mail.com",
"biography":"Hello, world!!!"}}`,
		withAuthorization: true,
		expectedCode:      http.StatusOK,
		expectedBody: `{"user":{
"login":"ekvo",
"first_name":"Alexander",
"last_name":"Burlin",
"phone":"\+79876543210",
"email":"alexan@mail.com",
"access":"4",
"image":"https:\/\/avatars.fastly.steamstatic.com\/3c52f5b49925abe74105c49e0f6eb368b5149e7c_full.jpg",
"biography":"Hello, world!!!",
"created_at":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"updated_at":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"last_connect":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))"}}`,
		msg: "status = 200 and update all fileds exept \"created_at\"",
	},
	{
		url:               "/zephyr/profile/ekvo",
		method:            http.MethodGet,
		withAuthorization: true,
		expectedCode:      http.StatusOK,
		expectedBody: `{"profile":{
"login":"ekvo",
"first_name":"Alexander",
"image":"https:\/\/avatars.fastly.steamstatic.com\/3c52f5b49925abe74105c49e0f6eb368b5149e7c_full.jpg",
"biography":"Hello, world!!!",
"followers":1,
"created_at":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"last_connect":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"following":true}}`,
		msg: "status = 200, get profile of user",
	},
	{
		url:               "/zephyr/profile/alien",
		method:            http.MethodGet,
		withAuthorization: true,
		expectedCode:      http.StatusNotFound,
		expectedBody:      `{"errors":{"profile":"not faound"}}`,
		msg:               "status = 404 with message \"not faound\"",
	},
	{
		url:               "/zephyr/profile/ekvo/follow",
		method:            http.MethodDelete,
		withAuthorization: true,
		expectedCode:      http.StatusOK,
		expectedBody: `"login":"ekvo",
"first_name":"Alexander",
"image":"https:\/\/avatars.fastly.steamstatic.com\/3c52f5b49925abe74105c49e0f6eb368b5149e7c_full.jpg",
"biography":"Hello, world!!!",
"followers":0,
"created_at":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"last_connect":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"following":false}}`,
		msg: "status = 200 and followers = 0 with following - false",
	},
	{
		url:               "/zephyr/profile/ekvo/follow",
		method:            http.MethodPut,
		withAuthorization: true,
		expectedCode:      http.StatusCreated,
		expectedBody: `{"profile":{
"login":"ekvo",
"first_name":"Alexander",
"image":"https:\/\/avatars.fastly.steamstatic.com\/3c52f5b49925abe74105c49e0f6eb368b5149e7c_full.jpg",
"biography":"Hello, world!!!",
"followers":1,
"created_at":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"last_connect":"\d\d\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[0-1])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]).\d\d\d(Z|([+-]([01][0-9]|2[0-3]):(([0-5][0-9]))))",
"following":true}}`,
		msg: "status 200 now followers is 1 and have following - true",
	},
}

var jwtTOken = ""

func getJWTTokenFromResponse(w *httptest.ResponseRecorder) {
	res := w.Result()
	defer res.Body.Close()
	usr := struct {
		users.TokenResponse `json:"approve"`
	}{}
	_ = json.NewDecoder(res.Body).Decode(&usr)
	jwtTOken = usr.Token
}

func TestRouterUser(t *testing.T) {
	assertT := assert.New(t)

	store := NewStoreMock()
	router := gin.New()

	first := router.Group("/zephyr")
	first.Use(common.ContextMiddleware(CTXUsersTimeRequest))
	UserBeforeRegister(first.Group("/connect"), store)
	first.Use(users.Autorization(store))
	UserAfterRegister(first.Group("/user"), store)
	SpeakerFolower(first.Group("/profile"), store)

	for i, test := range dataForTests {
		test.body = strings.Replace(test.body, "\n", "", -1)

		req, err := http.NewRequest(test.method, test.url, bytes.NewBufferString(test.body))
		require.NoError(t, err)
		if test.withAuthorization {
			req.Header.Set("Authorization", "Bearer "+jwtTOken)
		}
		if test.method == http.MethodPost || test.method == http.MethodPut {
			req.Header.Set("Content-Type", "application/json")
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assertT.Equal(test.expectedCode, w.Code, fmt.Sprintf("Status code invalid - %s", test.msg))

		test.expectedBody = strings.Replace(test.expectedBody, "\n", "", -1)
		assertT.Regexp(test.expectedBody, w.Body.String(), fmt.Sprintf("Response Body invalid - %s", test.msg))

		if i == 0 {
			getJWTTokenFromResponse(w)
		}
	}
}
