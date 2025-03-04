package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

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
func findByFieldWithKey(user UserModel, flag int) (string, string, error) {
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

// UserConnect - описывает регистрацию и подключение пользователя
type UserConnect interface {
	// SaveOneUser - запись данных пользователя и получение ногового, уникального ID
	SaveOneUser(ctx context.Context, user UserModel) (uint, error)

	// LoginUserWithUpdateTime - поиск по определенному полю,
	// определяемому с помощью 'flag' см. usermodel.go в текущем пакете 'source'
	// так же должна обновлять поле LastConnection *time.Time
	LoginUserWithUpdateTime(ctx context.Context, user UserModel, flag int) (UserModel, error)
}

// UserSave - добавление пользователя в базу данных
func (s SQLSource) SaveOneUser(ctx context.Context, user UserModel) (uint, error) {
	row := s.sourceDB.QueryRowContext(ctx, `
WITH to_users AS(
INSERT INTO users (login,
                   hash_password,
                   first_name,
                   last_name,
                   phone,
                   email,
                   created_at)
       VALUES ($1,$2,$3,$4,$5,$6,$7)
       RETURNING id    
),to_access AS (
INSERT INTO users_access(id_user,
                         access)
       VALUES ((SELECT id FROM to_users),$8)
), to_biography AS (
INSERT INTO users_biography (id_user,
                             image,
                             biography)
       VALUES((SELECT id FROM to_users),$9,$10)   
)
SELECT id 
FROM to_users;`,
		user.Login,                             //1
		user.Password,                          //2
		user.FirstName,                         //3
		whenEmptyStringThenNULL(user.LastName), //4
		whenEmptyStringThenNULL(user.Phone),    //5
		user.Email,                             //6
		user.CreatedAt,                         //7
		user.Access,                            //8
		whenEmptyStringThenNULL(user.Image),    //9
		whenEmptyStringThenNULL(user.Bio),      //10
	)
	err := row.Scan(&user.ID)

	return user.ID, err
}

func whenEmptyStringThenNULL(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{"", false}
	}
	return sql.NullString{*s, len(*s) != 0}
}

// LoginUser - метод для поиска пооьзователя по параметру и флагу,
// а также обновление поля 'users.last_connection'
func (s SQLSource) LoginUserWithUpdateTime(ctx context.Context, user UserModel, flag int) (UserModel, error) {
	byField, param, err := findByFieldWithKey(user, flag)
	if err != nil {
		return UserModel{}, err
	}
	row := s.sourceDB.QueryRowContext(ctx, fmt.Sprintf(`
WITH up_last_con AS (
    UPDATE users
        SET last_connection = $2
        WHERE %s = $1
        RETURNING id,last_connection
)
SELECT u.id,
       u.login,
       u.hash_password,
       u.first_name,
       u.last_name,
       u.phone,
       u.email,
       u.created_at,
       u.updated_at,
       up.last_connection,
       a.access,
       b.image,
       b.biography
FROM up_last_con up
    LEFT JOIN users u
        ON up.id = u.id
    LEFT JOIN users_biography b
        ON up.id = b.id_user
    LEFT JOIN users_access a
        ON up.id = a.id_user;`, byField), param, time.Now())

	return scanUserModel(row)
}

func scanUserModel(row *sql.Row) (UserModel, error) {
	var (
		uModel                            UserModel
		lastName, phone, image, biography sql.NullString
		lastConnection, updateAt          sql.NullTime
	)
	err := row.Scan(
		&uModel.ID,
		&uModel.Login,
		&uModel.Password,
		&uModel.FirstName,
		&lastName,
		&phone,
		&uModel.Email,
		&uModel.CreatedAt,
		&updateAt,
		&lastConnection,
		&uModel.Access,
		&image,
		&biography,
	)
	if lastName.Valid {
		uModel.LastName = &lastName.String
	}
	if phone.Valid {
		uModel.Phone = &phone.String
	}
	if updateAt.Valid {
		uModel.UpdatedAT = &updateAt.Time
	}
	if lastConnection.Valid {
		uModel.LastConnection = &lastConnection.Time
	}
	if image.Valid {
		uModel.Image = &image.String
	}
	if biography.Valid {
		uModel.Bio = &biography.String
	}
	return uModel, err
}

// UserApproveAndFollowing - групирует методы связанные с обработкой зарегистрированных пользователей
type UserApproveFollowing interface {
	UserApprove
	UserFollowing
}

// UserApprove - получение, обновление данных пользователя
type UserApprove interface {
	// FindOneUserByField - поиск пользователя по полую, определяемому с помощью 'flag' - см топ. данного фафла
	FindOneUserByField(ctx context.Context, user UserModel, falg int) (UserModel, error)

	NewDataUser(ctx context.Context, user UserModel) error
}

// FindOneUser - поиск пользователя по ID
func (s SQLSource) FindOneUserByField(ctx context.Context, user UserModel, flag int) (UserModel, error) {
	byField, param, err := findByFieldWithKey(user, flag)
	if err != nil {
		return UserModel{}, err
	}
	row := s.sourceDB.QueryRowContext(ctx, fmt.Sprintf(`
SELECT u.id,
       u.login,
       u.hash_password,
       u.first_name,
       u.last_name,
       u.phone,
       u.email,
       u.created_at,
       u.updated_at,
       u.last_connection,
       a.access,
       b.image,
       b.biography
FROM users u 
LEFT JOIN users_access a 
ON u.id = a.id_user
LEFT JOIN users_biography b 
ON u.id = b.id_user
WHERE u.%s = $1;`, byField), param)

	return scanUserModel(row)
}

// NewDataUser - записывае обновления в поля таблиц: 'users','users_access' & 'users_biography'
// поля 'users.created_ad' & 'users.lasct_connection' - не обновляются,
// а также 'users.id', 'users_access.id_user' & 'users_biography.id_user' - не обновляются
func (s SQLSource) NewDataUser(ctx context.Context, user UserModel) error {
	_, err := s.sourceDB.ExecContext(ctx, `
WITH to_user AS (
  UPDATE users
  SET logon = $2,
      hash_password = $3,
      first_name = $4,
      second_name = $5,
      phone = $6,
      email = $7,
      updated_at = $8
  WHERE id = $1   
), to_access AS (
  UPDATE users_access
  SET access = $9
  WHERE id_user = $1  
),
UPDATE users_biography
SET image = $10,
    biography = $11
WHERE id_user = $1;`,
		user.ID,                                //1
		user.Login,                             //2
		user.Password,                          //3
		user.FirstName,                         //4
		whenEmptyStringThenNULL(user.LastName), //5
		whenEmptyStringThenNULL(user.Phone),    //6
		user.Email,                             //7
		user.UpdatedAT,                         //8
		user.Access,                            //9
		whenEmptyStringThenNULL(user.Image),    //10
		whenEmptyStringThenNULL(user.Bio),      //11
	)
	return err
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

// IsRelationship - создает уникальную связку ключей (id_user,id_folower)
func (s SQLSource) NewRelationship(ctx context.Context, userFollower, userSpeaker UserModel) error {
	_, err := s.sourceDB.ExecContext(ctx, `
INSERT INTO followers (id_user,id_follower)
VALUES($1,$2);`, userSpeaker.ID, userFollower.ID)
	return err
}

// IsRelationship - проверяет наличие уникальной связки ключей (id_user,id_folower)
func (s SQLSource) IsRelationship(ctx context.Context, userFollower, userSpeaker UserModel) (bool, error) {
	result := false
	err := s.sourceDB.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT *
    FROM followers
    WHERE id_user = $1 AND id_follower = $2);`, userSpeaker.ID, userFollower.ID).Scan(&result)

	return result, err
}

// EndRelationship - удаляет уникальную связку ключей (id_user,id_folower)
func (s SQLSource) EndRelationship(ctx context.Context, userFollower, userSpeaker UserModel) error {
	_, err := s.sourceDB.ExecContext(ctx, `
DELETE FROM followers
WHERE id_user = $1 AND id_follower = $2;`, userSpeaker.ID, userFollower.ID)
	return err
}
