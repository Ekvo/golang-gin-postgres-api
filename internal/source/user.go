package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	"strings"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// ErrSourceAlreadyExists - во время регистрации(signup)
var ErrSourceAlreadyExists = errors.New("resource already exists")

// ErrSourceNoUpdate - для маркировки отрицательного обновления, создания записей
var ErrSourceNoUpdate = errors.New("no update or delete")

var ErrSourceNotFound = errors.New("resource not found")

// UserSave - добавление пользователя в базу данных
// в таблицы: 'users', 'users_access', 'users_biography', 'followers' - (с подпиской на себя)
// с полученим 'id'
func (s *SQLSource) SaveOneUser(ctx context.Context, data any) (uint, error) {
	userModel := data.(models.UserModel)
	insertUser := func(ctx context.Context) error {
		return s.sourceDBTX.Tx.QueryRowContext(ctx, `
WITH to_users AS (
    INSERT INTO users (login,
                       hash_password,
                       first_name,
                       last_name,
                       phone,
                       email,
                       created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id),
     to_access AS (
         INSERT INTO users_access (id_user, access)
             VALUES ((SELECT id FROM to_users), $8)),
     to_biography AS (
         INSERT INTO users_biography (id_user, image, biography)
             VALUES ((SELECT id FROM to_users), $9, $10)),
     to_followers AS (
         INSERT INTO followers (id_user, id_follower)
             VALUES ((SELECT id FROM to_users), (SELECT id FROM to_users)))
SELECT id
FROM to_users`,
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
	fl := ctx.Value(flag.KeyFlagFiled).(int)
	loginWithUpdateUser := func(ctx context.Context) error {
		userModel := data.(models.UserModel)
		byField, param, err := flag.FindByFieldWithKey(userModel, fl)
		if err != nil {
			return err
		}
		row := s.sourceDBTX.Tx.QueryRowContext(ctx, fmt.Sprintf(`
WITH up_last_con AS (
    UPDATE users
        SET last_connection = $2
        WHERE %s = $1
        RETURNING id,last_connection)
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
       b.biography,
       (SELECT COUNT(id_follower)
        FROM followers
        WHERE id_user = u.id) AS unique_followers
FROM up_last_con up
         JOIN users u ON up.id = u.id
         JOIN users_biography b ON up.id = b.id_user
         JOIN users_access a ON up.id = a.id_user;`, byField),
			param, userModel.LastConnection)
		user, err = scanUserModel[*sql.Row](row)
		return err
	}
	return user, s.sourceDBTX.Transaction(ctx, loginWithUpdateUser)
}

// UserPtrField - сканирование полей 'UserModel' с типом 'ptr'
type UserPtrField struct {
	LastName, Phone, Image, Biography sql.NullString
	LastConnection, UpdateAt          sql.NullTime
}

// ToUserModel - передает указатели на поля 'UserModel'
// в паре с 'func scanUserModel[T SQLRowsRowScan](row T) (models.UserModel, error)'
func (uPtrF *UserPtrField) toUserModel(uModel *models.UserModel) {
	if uPtrF.LastName.Valid {
		uModel.LastName = &uPtrF.LastName.String
	}
	if uPtrF.Phone.Valid {
		uModel.Phone = &uPtrF.Phone.String
	}
	if uPtrF.UpdateAt.Valid {
		uModel.UpdatedAt = &uPtrF.UpdateAt.Time
	}
	if uPtrF.LastConnection.Valid {
		uModel.LastConnection = &uPtrF.LastConnection.Time
	}
	if uPtrF.Image.Valid {
		uModel.Image = &uPtrF.Image.String
	}
	if uPtrF.Biography.Valid {
		uModel.Bio = &uPtrF.Biography.String
	}
}

func scanUserModel[T SQLRowsRowScan](row T) (models.UserModel, error) {
	var (
		uModel    models.UserModel
		ptrFields UserPtrField
	)
	err := row.Scan(
		&uModel.ID,
		&uModel.Login,
		&uModel.Password,
		&uModel.FirstName,
		&ptrFields.LastName,
		&ptrFields.Phone,
		&uModel.Email,
		&uModel.CreatedAt,
		&ptrFields.UpdateAt,
		&ptrFields.LastConnection,
		&uModel.Access,
		&ptrFields.Image,
		&ptrFields.Biography,
		&uModel.NumberOfFollowers,
	)
	ptrFields.toUserModel(&uModel)
	return uModel, err
}

// FindOneUser - поиск пользователя по полю определнноу с помощью флага
// 'flag_field' из 'context.Value'
func (s *SQLSource) FindOneUserByField(ctx context.Context, data any) (models.UserModel, error) {
	userModel := data.(models.UserModel)
	fl := ctx.Value(flag.KeyFlagFiled).(int)
	byField, param, err := flag.FindByFieldWithKey(userModel, fl)
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
       b.biography,
       (SELECT COUNT(id_follower) 
        FROM followers 
        WHERE id_user = u.id) AS unique_followers
FROM users u
         JOIN users_access a ON u.id = a.id_user
         JOIN users_biography b ON u.id = b.id_user         
WHERE u.%s = $1;`, byField), param)
	return scanUserModel[*sql.Row](row)
}

func (s *SQLSource) FindUserList(ctx context.Context, data any) ([]models.UserModel, error) {
	property := data.(models.UserProperty)
	query := strings.Builder{}
	query.WriteString(`
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
       b.biography,
       (SELECT COUNT(id_follower) 
        FROM followers 
        WHERE id_user = u.id) AS unique_followers
FROM users u
         JOIN users_access a ON u.id = a.id_user
         JOIN users_biography b ON u.id = b.id_user`)
	args := []any{}
	if property.NoEmpty() {
		numberOfArg := 1
		and := false
		query.WriteString("\nWHERE ")
		if len(property.FirstName) > 0 {
			query.WriteString(fmt.Sprintf("u.first_name = $%d", numberOfArg))
			numberOfArg++
			args = append(args, property.FirstName)
			and = true
		}
		if len(property.LastName) > 0 {
			if and {
				query.WriteString("\nAND ")
			}
			query.WriteString(fmt.Sprintf("u.last_name = $%d", numberOfArg))
			numberOfArg++
			args = append(args, property.LastName)
			and = true
		}
		if !property.StartDate.IsZero() {
			if and {
				query.WriteString("\nAND ")
			}
			query.WriteString(fmt.Sprintf("u.created_at BETWEEN $%d AND $%d", numberOfArg, numberOfArg+1))
			numberOfArg += 2
			args = append(args, property.StartDate)
			args = append(args, property.EndDate)
		}
		if property.Limit != 0 {
			query.WriteString(fmt.Sprintf("\nLIMIT $%d", numberOfArg))
			numberOfArg++
			args = append(args, property.Limit)
		}
		if property.Offset != 0 {
			query.WriteString(fmt.Sprintf(" OFFSET $%d", numberOfArg))
			args = append(args, property.Offset)
		}
	}
	query.WriteByte(';')
	rows, err := s.sourceDBTX.DB.QueryContext(ctx, query.String(), args...)
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
            hash_password = $3,
            first_name = $4,
            last_name = $5,
            phone = $6,
            email = $7,
            updated_at = $8
        WHERE id = $1),
     to_access AS (
         UPDATE users_access
             SET access = $9
             WHERE id_user = $1)
UPDATE users_biography
SET image     = $10,
    biography = $11
WHERE id_user = $1;`,
			userModel.ID,        //1
			userModel.Login,     //2
			userModel.Password,  //3
			userModel.FirstName, //4
			common.WhenEmptyStringThenNULL(userModel.LastName), //5
			common.WhenEmptyStringThenNULL(userModel.Phone),    //6
			userModel.Email,     //7
			userModel.UpdatedAt, //8
			userModel.Access,    //9
			common.WhenEmptyStringThenNULL(userModel.Image), //10
			common.WhenEmptyStringThenNULL(userModel.Bio),   //11
		)
		return err
	}
	return s.sourceDBTX.Transaction(ctx, updateUser)
}

// ErrSourceRelationship - переданны некорректные данные для реализации отношений пользователей
var ErrSourceRelationship = errors.New("is impossible - update or get Relationship with current data")

// IsRelationship - создает уникальную связку ключей (id_user,id_folower)
// data[0] - follower, data[1] - speaker
func (s *SQLSource) NewRelationship(ctx context.Context, data any) error {
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return ErrSourceRelationship
	}
	updateRelationship := func(ctx context.Context) error {
		_, err := s.sourceDBTX.Tx.ExecContext(ctx, `
INSERT INTO followers (id_user, id_follower)
VALUES ($1, $2);`, followerSpeaker[1], followerSpeaker[0])
		return err
	}
	return s.sourceDBTX.Transaction(ctx, updateRelationship)
}

// IsRelationship - проверяет наличие уникальной связки ключей (id_user,id_folower)
func (s *SQLSource) IsRelationship(ctx context.Context, data any) (bool, error) {
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return false, ErrSourceRelationship
	}
	subscription := false
	err := s.sourceDBTX.DB.QueryRowContext(ctx, `
SELECT EXISTS(SELECT *
              FROM followers
              WHERE id_user = $1
                AND id_follower = $2);`, followerSpeaker[1], followerSpeaker[0]).Scan(&subscription)
	return subscription, err
}

// IsRelationshipList - получение статус подписки пользователя полученного через context.Value()
// на пользователей из списка переаднного череа 'data'
func (s *SQLSource) IsRelationshipList(ctx context.Context, data any) (map[uint]bool, error) {
	lineSpeakerID := data.([]uint)
	followerID := ctx.Value(flag.KeyUserID).(uint)
	rows, err := s.sourceDBTX.DB.QueryContext(ctx, fmt.Sprintf(`
SELECT id_user
FROM followers
WHERE id_user IN (%s)
  AND id_follower = $1;`, lineSpeakerID), followerID)
	if err != nil {
		return nil, err
	}
	speakerFollow := make(map[uint]bool)
	for rows.Next() {
		var speakerID uint
		if err := rows.Scan(&speakerID); err != nil {
			return nil, err
		}
		speakerFollow[speakerID] = true
	}
	return speakerFollow, rows.Err()
}

// EndRelationship - удаляет уникальную связку ключей (id_user,id_folower)
func (s *SQLSource) EndRelationship(ctx context.Context, data any) error {
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return ErrSourceRelationship
	}
	deleteRows := 0
	deleteRelationship := func(ctx context.Context) error {
		err := s.sourceDBTX.Tx.QueryRowContext(ctx, `
WITH del_relationship AS (
    DELETE
        FROM followers
            WHERE (id_user, id_follower) = ($1, $2)
					RETURNING *)
SELECT COUNT(*)
FROM del_relationship;`, followerSpeaker[1], followerSpeaker[0]).Scan(&deleteRows)
		if err != nil {
			return err
		}
		if deleteRows != 1 {
			return ErrSourceNotFound
		}
		return nil
	}
	return s.sourceDBTX.Transaction(ctx, deleteRelationship)
}

// ErrSourceMainUser - ошибка удаления пользователя с индексом '1'
var ErrSourceMainUser = errors.New("main user cannot be deleted")

// DeleteUser - в струтуре базы user играет ключивую роль (references)
// полное удаление  может с большой вероятностью задеть данные иных пользоватлей
// например, если удаляеться создатель тега
//
// выход: теги становяться частью пользователя с ID =  1 (собсвенностью компании)
// остальное полностью удаляется
//
// 'delRows' - проверяет количесво удаленный пользователей 'delRows!=1 -> Rollback'
func (s *SQLSource) DeleteOneUser(ctx context.Context, data any) error {
	userID := data.(uint)
	if userID == 1 {
		return ErrSourceMainUser
	}
	deleteUser := func(ctx context.Context) error {
		delRows := 0
		err := s.sourceDBTX.Tx.QueryRowContext(ctx, `
WITH tag_update AS (
    UPDATE tags
        SET id_tag_maker = 1
        WHERE id_tag_maker = $1),
     id_comment AS (SELECT DISTINCT id
                    FROM comments
                    WHERE id_autor = $1
                       OR id_article IN (SELECT id
                                         FROM articles
                                         WHERE id_autor = $1)),
     del_body_comment AS (
         DELETE
             FROM comments_body
                 WHERE id_comment IN (SELECT id from id_comment)),
     del_comment AS (
         DELETE
             FROM comments
                 WHERE id IN (SELECT ID from id_comment)),
     del_favorite_article AS (
         DELETE
             FROM articles_favorite
                 WHERE id_user = $1),
     del_article_tag AS (
         DELETE
             FROM articles_tags
                 WHERE id_article IN (SELECT id
                                      FROM articles
                                      WHERE id_autor = $1)),
     del_body_article AS (
         DELETE
             FROM articles_body
                 WHERE id_article IN (SELECT id
                                      FROM articles
                                      WHERE id_autor = $1)),
     del_article AS (
         DELETE
             FROM articles
                 WHERE id_autor = $1),
     del_followers AS (
         DELETE
             FROM followers
                 WHERE id_user = $1),
     del_user_access AS (
         DELETE
             FROM users_access
                 WHERE id_user = $1),
     del_body_user AS (
         DELETE
             FROM users_biography
                 WHERE id_user = $1),
     del_user AS (
         DELETE
             FROM users
                 WHERE id = $1
                 RETURNING *)
SELECT COUNT(*)
FROM del_user;`, userID).Scan(&delRows)
		if err != nil {
			return err
		}
		if delRows != 1 {
			return ErrSourceNotFound
		}
		return nil
	}
	return s.sourceDBTX.Transaction(ctx, deleteUser)
}
