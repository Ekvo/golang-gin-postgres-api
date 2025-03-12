package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// UserSave - добавление пользователя в базу данных
// в таблицы: 'users', 'users_access', 'users_biography', 'followers' - (с подпиской на себя)
// с полученим 'id'
func (s *SQLSource) SaveOneUser(ctx context.Context, data any) (uint, error) {
	userModel := data.(models.UserModel)
	insertUser := func(ctx context.Context) error {
		return s.sourceDBTX.Tx.QueryRowContext(ctx, `
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
SELECT id FROM to_users;`,
			userModel.Login,     //1
			userModel.Password,  //2
			userModel.FirstName, //3
			common.WhenEmptyStringThenNULL(userModel.LastName), //4
			common.WhenEmptyStringThenNULL(userModel.Phone),    //5
			userModel.Email,     //6
			userModel.CreatedAt, //7
			userModel.Access,    //8
			common.WhenEmptyStringThenNULL(userModel.Image), //9
			common.WhenEmptyStringThenNULL(userModel.Bio),   //10
		).Scan(&userModel.ID)
	}
	return userModel.ID, s.sourceDBTX.Transaction(ctx, insertUser)
}

// LoginUser - метод для поиска пооьзователя по параметру и флагу,
// а также обновление поля 'users.last_connection'
// 'flag_field' из 'context.Value'
func (s *SQLSource) LoginUserWithUpdateTime(ctx context.Context, data any) (models.UserModel, error) {
	var user models.UserModel
	loginWithUpdateUser := func(ctx context.Context) error {
		flag := ctx.Value(models.KeyFlagFiled).(int)
		userModel := data.(models.UserModel)
		byField, param, err := models.FindByFieldWithKey(userModel, flag)
		if err != nil {
			return err
		}
		row := s.sourceDBTX.Tx.QueryRowContext(ctx, fmt.Sprintf(`
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
        ON up.id = a.id_user;`, byField),
			param, userModel.LastConnection)
		user, err = scanUserModel[*sql.Row](row)
		return err
	}
	return user, s.sourceDBTX.Transaction(ctx, loginWithUpdateUser)
}

func scanUserModel[T SQLRowsRowScan](row T) (models.UserModel, error) {
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

// FindOneUser - поиск пользователя по полю определнноу с помощью флага
// 'flag_field' из 'context.Value'
func (s *SQLSource) FindOneUserByField(ctx context.Context, data any) (models.UserModel, error) {
	userModel := data.(models.UserModel)
	flag := ctx.Value(models.KeyFlagFiled).(int)
	byField, param, err := models.FindByFieldWithKey(userModel, flag)
	if err != nil {
		return userModel, err
	}
	row := s.sourceDBTX.DB.QueryRowContext(ctx, fmt.Sprintf(`
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
    LEFT JOIN users_access a ON u.id = a.id_user
    LEFT JOIN users_biography b ON u.id = b.id_user
WHERE u.%s = $1;`, byField), param)
	return scanUserModel[*sql.Row](row)
}

func (s *SQLSource) FindUserList(ctx context.Context, data any) ([]models.UserModel, error) {
	userProperty := data.(models.UserProperty)
	rows, err := s.sourceDBTX.DB.QueryContext(ctx, `
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
    JOIN users_access a ON u.id = a.id_user 
    JOIN users_biography b ON u.id = b.id_user
WHERE u.created_at BETWEEN $3 AND $4
  AND (
      CASE 
          WHEN length($1) > 0 THEN u.first_name = $1
          ELSE TRUE = TRUE --all first_names
      END
  ) AND (
      CASE 
          WHEN length($2) > 0 THEN u.last_name = $2
          ELSE TRUE = TRUE --all last_names
      END)
LIMIT $5 OFFSET $6;`,
		userProperty.FirstName,
		userProperty.LastName,
		userProperty.StartDate,
		userProperty.EndDate,
		userProperty.Limit,
		userProperty.Offset,
	)
	if err != nil {
		return nil, err
	}
	return scanUsers(rows)
}

func scanUsers(rows *sql.Rows) ([]models.UserModel, error) {
	var users []models.UserModel
	for rows.Next() {
		user, err := scanUserModel[*sql.Rows](rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// NewDataUser - записывае обновления в поля таблиц: 'users','users_access' & 'users_biography'
// поля 'users.created_ad' & 'users.lasct_connection' - не обновляются,
// а также 'users.id', 'users_access.id_user' & 'users_biography.id_user' - не обновляются
func (s *SQLSource) NewDataUser(ctx context.Context, data any) error {
	updateUser := func(ctx context.Context) error {
		userModel := data.(models.UserModel)
		_, err := s.sourceDBTX.DB.ExecContext(ctx, `
WITH to_user AS (
  UPDATE users
  SET login = $2,
      hash_password = (
          CASE
          WHEN length($3) > 0 THEN $3
          ELSE (SELECT hash_password
                FROM users
                WHERE id = $1
                LIMIT 1)
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
			userModel.ID,        //1
			userModel.Login,     //2
			userModel.Password,  //3
			userModel.FirstName, //4
			common.WhenEmptyStringThenNULL(userModel.LastName), //5
			common.WhenEmptyStringThenNULL(userModel.Phone),    //6
			userModel.Email,     //7
			userModel.UpdatedAT, //8
			userModel.Access,    //9
			common.WhenEmptyStringThenNULL(userModel.Image), //10
			common.WhenEmptyStringThenNULL(userModel.Bio),   //11
		)
		return err
	}
	return s.sourceDBTX.Transaction(ctx, updateUser)
}

// ErrSourceRelationship - переданны некорректные данные для реализации отношений пользователей
var ErrSourceRelationship = errors.New("is impossible - update Relationship with current data")

// IsRelationship - создает уникальную связку ключей (id_user,id_folower)
// data - []models.UserModel длины = 2
// data[0] - follower, data[1] - speaker
func (s *SQLSource) NewRelationship(ctx context.Context, data any) error {
	userModels := data.([]models.UserModel)
	if len(userModels) != 2 {
		return ErrSourceRelationship
	}
	updateRelationship := func(ctx context.Context) error {
		_, err := s.sourceDBTX.Tx.ExecContext(ctx, `
INSERT INTO followers (id_user,id_follower)
VALUES($1,$2);`, userModels[1].ID, userModels[0].ID)
		return err
	}
	return s.sourceDBTX.Transaction(ctx, updateRelationship)

}

// IsRelationship - проверяет наличие уникальной связки ключей (id_user,id_folower)
func (s *SQLSource) IsRelationship(ctx context.Context, data any) (bool, error) {
	userModels := data.([]models.UserModel)
	if len(userModels) != 2 {
		return false, ErrSourceRelationship
	}
	follow := false
	err := s.sourceDBTX.DB.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT *
    FROM followers
    WHERE id_user = $1 AND id_follower = $2);`, userModels[1].ID, userModels[0].ID).Scan(&follow)
	return follow, err
}

// EndRelationship - удаляет уникальную связку ключей (id_user,id_folower)
func (s *SQLSource) EndRelationship(ctx context.Context, data any) error {
	userModels := data.([]models.UserModel)
	if len(userModels) != 2 {
		return ErrSourceRelationship
	}
	deleteRelationship := func(ctx context.Context) error {
		_, err := s.sourceDBTX.Tx.ExecContext(ctx, `
DELETE FROM followers
WHERE id_user = $1 AND id_follower = $2 LIMIT 1;`, userModels[1].ID, userModels[0].ID)
		return err
	}

	return s.sourceDBTX.Transaction(ctx, deleteRelationship)
}
