package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// ErrSourceRelationship - переданны некорректные данные для реализации отношений пользователей
var ErrSourceRelationship = errors.New("is impossible - update or get Relationship with current data")

// UserSave - добавление пользователя в базу данных
// в таблицы: 'users', 'users_access', 'users_biography', 'followers' - (с подпиской на себя)
// с полученим 'id'
func (s SQLSource) SaveOneUser(ctx context.Context, data any) (uint, error) {
	userModel := data.(models.UserModel)
	insertUser := func(ctx context.Context) error {
		err := s.pTx.Tx.QueryRow(ctx, `
WITH to_users AS (
    INSERT INTO users (login,
                       hash_password,
                       access,
                       first_name,
                       last_name,
                       phone,
                       email,
                       image,
                       biography,
                       created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id),
     to_followers AS (
         INSERT INTO followers (id_speaker, id_follower)
             VALUES ((SELECT id FROM to_users), (SELECT id FROM to_users)))
SELECT id
FROM to_users;`,
			userModel.Login,     //1
			userModel.Password,  //2
			userModel.Access,    //3
			userModel.FirstName, //4
			common.WhenEmptyStringThenNULL(userModel.LastName), //5
			common.WhenEmptyStringThenNULL(userModel.Phone),    //6
			userModel.Email, //7
			common.WhenEmptyStringThenNULL(userModel.Image), //8
			common.WhenEmptyStringThenNULL(userModel.Bio),   //9
			userModel.CreatedAt,                             //10
		).Scan(&userModel.ID)
		if err != nil {
			return ErrSourceAlreadyExists
		}
		return nil
	}
	return userModel.ID, s.pTx.Transaction(ctx, insertUser)
}

// LoginUser - метод для поиска пооьзователя по параметру и флагу,
// а также обновление поля 'users.last_connection'
// 'flag_field' из 'context.Value'
func (s SQLSource) LoginUserWithUpdateTime(ctx context.Context, data any) (models.UserModel, error) {
	var user models.UserModel
	userModel := data.(models.UserModel)
	loginWithUpdateUser := func(ctx context.Context) error {
		err := s.pTx.Tx.QueryRow(ctx, `
UPDATE users
SET last_connection = $2
WHERE login = $1
RETURNING id,hash_password;`,
			userModel.Login, userModel.LastConnection,
		).Scan(&user.ID, &user.Password)
		if err != nil {
			return ErrSourceNotFound
		}
		return nil
	}
	return user, s.pTx.Transaction(ctx, loginWithUpdateUser)
}

// FindOneUser - поиск пользователя по полю определнноу с помощью флага
// 'flag_field' из 'context.Value'
func (s SQLSource) FindOneUserByField(ctx context.Context, data any) (models.UserModel, error) {
	userModel := data.(models.UserModel)
	fl := ctx.Value(flag.KeyFlagFiled).(int)
	byField, param, err := flag.FindByFieldWithKey(userModel, fl)
	if err != nil {
		return userModel, err
	}
	row := s.pTx.Pool.QueryRow(ctx, fmt.Sprintf(`
SELECT id,
       login,
       hash_password,
       access,
       first_name,
       last_name,
       phone,
       email,
       image,
       biography,
       created_at,
       updated_at,
       last_connection,
       (SELECT COUNT(id_follower)
        FROM followers
        WHERE id_speaker = users.id) AS unique_followers
FROM users
WHERE %s = $1;`, byField), param)
	return scanUserModel[pgx.Row](row)
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

func scanUserModel[T SQLScan](row T) (models.UserModel, error) {
	var (
		uModel    models.UserModel
		ptrFields UserPtrField
	)
	if err := row.Scan(
		&uModel.ID,
		&uModel.Login,
		&uModel.Password,
		&uModel.Access,
		&uModel.FirstName,
		&ptrFields.LastName,
		&ptrFields.Phone,
		&uModel.Email,
		&ptrFields.Image,
		&ptrFields.Biography,
		&uModel.CreatedAt,
		&ptrFields.UpdateAt,
		&ptrFields.LastConnection,
		&uModel.NumberOfFollowers,
	); err != nil {
		return uModel, ErrSourceNotFound
	}
	ptrFields.toUserModel(&uModel)
	return uModel, nil
}

func (s SQLSource) FindUserList(ctx context.Context, data any) ([]models.UserModel, error) {
	property := data.(models.UserProperty)
	query := strings.Builder{}
	query.WriteString(`
SELECT id,
       login,
       hash_password,
       access,
       first_name,
       last_name,
       phone,
       email,
       image,
       biography,
       created_at,
       updated_at,
       last_connection,
       (SELECT COUNT(id_follower)
        FROM followers
        WHERE id_speaker = users.id) AS unique_followers
FROM users`)
	args := []any{}
	numberOfArg := 1
	where := false
	if property.IsFirstName() {
		where = whereANDOR(&query, where, "")
		query.WriteString(fmt.Sprintf("first_name = $%d", numberOfArg))
		numberOfArg++
		args = append(args, property.FirstName)
	}
	if property.IsLastName() {
		where = whereANDOR(&query, where, "AND")
		query.WriteString(fmt.Sprintf("last_name = $%d", numberOfArg))
		numberOfArg++
		args = append(args, property.LastName)
	}
	if !property.IsRangeZero() {
		_ = whereANDOR(&query, where, "AND")
		query.WriteString(fmt.Sprintf("created_at BETWEEN $%d AND $%d", numberOfArg, numberOfArg+1))
		numberOfArg += 2
		args = append(args, property.StartDate, property.EndDate)
	}
	limit := false
	if property.IsLimit() {
		limit = true
		query.WriteString(fmt.Sprintf("\nLIMIT $%d", numberOfArg))
		numberOfArg++
		args = append(args, property.Limit)
	}
	if property.IsOffset() {
		if limit {
			query.WriteString(fmt.Sprintf(" OFFSET $%d", numberOfArg))
		} else {
			query.WriteString(fmt.Sprintf("\nOFFSET $%d", numberOfArg))
		}
		args = append(args, property.Offset)
	}
	query.WriteByte(';')
	rows, err := s.pTx.Pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	return scanUsers(rows)
}

// whereANDOR - add to query(ptr) SQLcommand
//
// if it first - field unloading - WHERE else write commandSQL(OR, AND)
func whereANDOR(query *strings.Builder, where bool, commandSQL string) bool {
	if !where {
		query.WriteString("\nWHERE ")
	} else {
		query.WriteString(fmt.Sprintf("\n%s ", commandSQL))
	}
	return true
}

func limitOffsetToQuery(query *strings.Builder, lo common.LimitOffset, numberOfArg int) {

}

func scanUsers(rows pgx.Rows) ([]models.UserModel, error) {
	var users []models.UserModel
	for rows.Next() {
		user, err := scanUserModel[pgx.Rows](rows)
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
func (s SQLSource) NewDataUser(ctx context.Context, data any) error {
	userModel := data.(models.UserModel)
	updateUser := func(ctx context.Context) error {
		userID := uint(0) // check RETURNING
		err := s.pTx.Tx.QueryRow(ctx, `
UPDATE users
SET login        = $2,
    hash_password=$3,
    access=$4,
    first_name=$5,
    last_name=$6,
    phone=$7,
    email=$8,
    image=$9,
    biography=$10,
    updated_at=$11
WHERE id = $1
RETURNING id;`,
			userModel.ID,        //1
			userModel.Login,     //2
			userModel.Password,  //3
			userModel.Access,    //4
			userModel.FirstName, //5
			common.WhenEmptyStringThenNULL(userModel.LastName), //6
			common.WhenEmptyStringThenNULL(userModel.Phone),    //7
			userModel.Email, //8
			common.WhenEmptyStringThenNULL(userModel.Image), //9
			common.WhenEmptyStringThenNULL(userModel.Bio),   //10
			userModel.UpdatedAt,                             //11
		).Scan(&userID)
		if err != nil {
			return ErrSourceNotFound
		}
		return nil
	}
	return s.pTx.Transaction(ctx, updateUser)
}

// RemoveUser - remove user from database
//
// delete from table 'users' one row by 'users.id'
// remove from 'followers' where 'followers.id_speaker = users.id'
func (s SQLSource) RemoveUser(ctx context.Context, data any) error {
	userID := data.(uint)
	deleteUser := func(ctx context.Context) error {
		delID := uint(0) // check RETURNING
		err := s.pTx.Tx.QueryRow(ctx, `
DELETE 
FROM users 
WHERE id = $1 
RETURNING id;`, userID).Scan(&delID)
		if err != nil {
			return ErrSourceNotFound
		}
		_, err = s.pTx.Tx.Exec(ctx, `
DELETE
FROM followers
WHERE id_speaker = $1;`, userID)
		if err != nil {
			return err
		}
		return nil
	}
	return s.pTx.Transaction(ctx, deleteUser)
}

// IsRelationship - creates a unique keychain (id_user,id_folower)
//
// data[0] - follower, data[1] - speaker
func (s SQLSource) NewRelationship(ctx context.Context, data any) error {
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return ErrSourceRelationship
	}
	speakerID, followerID := uint(0), uint(0) // check RETURNING
	updateRelationship := func(ctx context.Context) error {
		err := s.pTx.Tx.QueryRow(ctx, `
INSERT INTO followers (id_speaker, id_follower)
SELECT $1, $2
WHERE EXISTS (SELECT id from users where id = $1)
  AND EXISTS (SELECT id from users where id = $2)
RETURNING *;`, followerSpeaker[1], followerSpeaker[0]).Scan(&speakerID, &followerID)
		return sourceError(err)
	}
	return s.pTx.Transaction(ctx, updateRelationship)
}

// sourceError - if Query return some error -> set some one Error from package source
func sourceError(err error) error {
	if err == pgx.ErrNoRows {
		return ErrSourceNotFound
	}
	if err != nil {
		return ErrSourceAlreadyExists
	}
	return nil
}

// IsRelationship - проверяет наличие уникальной связки ключей (id_user,id_folower)
func (s SQLSource) IsRelationship(ctx context.Context, data any) (bool, error) {
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return false, ErrSourceRelationship
	}
	subscription := false
	err := s.pTx.Pool.QueryRow(ctx, `
SELECT EXISTS(SELECT *
              FROM followers
              WHERE id_speaker = $1
                AND id_follower = $2);`, followerSpeaker[1], followerSpeaker[0]).Scan(&subscription)
	return subscription, err
}

// IsRelationshipList - получение статус подписки пользователя полученного через context.Value()
// на пользователей из списка переаднного череа 'data'
func (s SQLSource) IsRelationshipList(ctx context.Context, data any) (map[uint]bool, error) {
	lineSpeakerID := data.(string)
	followerID := ctx.Value(flag.KeyUserID).(uint)
	rows, err := s.pTx.Pool.Query(ctx, fmt.Sprintf(`
SELECT id_speaker
FROM followers
WHERE id_speaker IN (%s)
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
func (s SQLSource) EndRelationship(ctx context.Context, data any) error {
	followerSpeaker := data.([]uint)
	if len(followerSpeaker) != 2 {
		return ErrSourceRelationship
	}
	speakerID, followerID := uint(0), uint(0)
	deleteRelationship := func(ctx context.Context) error {
		err := s.pTx.Tx.QueryRow(ctx, `
DELETE
FROM followers
WHERE (id_speaker, id_follower) = ($1, $2)
RETURNING *;`, followerSpeaker[1], followerSpeaker[0]).Scan(&speakerID, &followerID)
		if err != nil {
			return ErrSourceNotFound
		}
		return nil
	}
	return s.pTx.Transaction(ctx, deleteRelationship)
}
