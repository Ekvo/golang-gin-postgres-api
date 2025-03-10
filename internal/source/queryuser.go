package source

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"time"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
)

// UserSave - добавление пользователя в базу данных
// в таблицы: 'users', 'users_access', 'users_biography', 'followers' - (с подпиской на себя)
// с полученим 'id'
func (s SQLSource) SaveOneUser(ctx context.Context, user models.UserModel) (uint, error) {
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
INSERT INTO users_access(id_user,access)
       VALUES ((SELECT id FROM to_users),$8)
), to_biography AS (
INSERT INTO users_biography (id_user,image,biography)
       VALUES((SELECT id FROM to_users),$9,$10)   
), to_followers AS (
INSERT INTO followers (id_user,id_follower)
       VALUES((SELECT id FROM to_users),(SELECT id FROM to_users))
)
SELECT id 
FROM to_users;`,
		user.Login,     //1
		user.Password,  //2
		user.FirstName, //3
		common.WhenEmptyStringThenNULL(user.LastName), //4
		common.WhenEmptyStringThenNULL(user.Phone),    //5
		user.Email,     //6
		user.CreatedAt, //7
		user.Access,    //8
		common.WhenEmptyStringThenNULL(user.Image), //9
		common.WhenEmptyStringThenNULL(user.Bio),   //10
	)
	err := row.Scan(&user.ID)

	return user.ID, err
}

// LoginUser - метод для поиска пооьзователя по параметру и флагу,
// а также обновление поля 'users.last_connection'
func (s SQLSource) LoginUserWithUpdateTime(ctx context.Context, user models.UserModel, flag int) (models.UserModel, error) {
	byField, param, err := models.FindByFieldWithKey(user, flag)
	if err != nil {
		return models.UserModel{}, err
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

func scanUserModel(row *sql.Row) (models.UserModel, error) {
	var (
		uModel                            models.UserModel
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

// FindOneUser - поиск пользователя по ID
func (s SQLSource) FindOneUserByField(ctx context.Context, user models.UserModel, flag int) (models.UserModel, error) {
	byField, param, err := models.FindByFieldWithKey(user, flag)
	if err != nil {
		return models.UserModel{}, err
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
func (s SQLSource) NewDataUser(ctx context.Context, user models.UserModel) error {
	_, err := s.sourceDB.ExecContext(ctx, `
WITH to_user AS (
  UPDATE users
  SET login = $2,
      hash_password = (
          CASE
          WHEN $3 != '' THEN $3
          ELSE (SELECT hash_password
                FROM users
                WHERE id = $1)
          END),
      first_name = $4,
      last_name = $5,
      phone = $6,
      email = $7,
      updated_at = $8
  WHERE id = $1   
), to_access AS (
  UPDATE users_access
  SET access = $9
  WHERE id_user = $1  
)
UPDATE users_biography
SET image = $10,
    biography = $11
WHERE id_user = $1;`,
		user.ID,        //1
		user.Login,     //2
		user.Password,  //3
		user.FirstName, //4
		common.WhenEmptyStringThenNULL(user.LastName), //5
		common.WhenEmptyStringThenNULL(user.Phone),    //6
		user.Email,     //7
		user.UpdatedAT, //8
		user.Access,    //9
		common.WhenEmptyStringThenNULL(user.Image), //10
		common.WhenEmptyStringThenNULL(user.Bio),   //11
	)
	return err
}

// IsRelationship - создает уникальную связку ключей (id_user,id_folower)
func (s SQLSource) NewRelationship(ctx context.Context, userFollower, userSpeaker models.UserModel) error {
	_, err := s.sourceDB.ExecContext(ctx, `
INSERT INTO followers (id_user,id_follower)
VALUES($1,$2);`, userSpeaker.ID, userFollower.ID)
	return err
}

// IsRelationship - проверяет наличие уникальной связки ключей (id_user,id_folower)
func (s SQLSource) IsRelationship(ctx context.Context, userFollower, userSpeaker models.UserModel) (bool, error) {
	follow := false
	err := s.sourceDB.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT *
    FROM followers
    WHERE id_user = $1 AND id_follower = $2);`, userSpeaker.ID, userFollower.ID).Scan(&follow)

	return follow, err
}

// EndRelationship - удаляет уникальную связку ключей (id_user,id_folower)
func (s SQLSource) EndRelationship(ctx context.Context, userFollower, userSpeaker models.UserModel) error {
	_, err := s.sourceDB.ExecContext(ctx, `
DELETE FROM followers
WHERE id_user = $1 AND id_follower = $2;`, userSpeaker.ID, userFollower.ID)
	return err
}
