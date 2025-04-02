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

// ErrSourceBigSlug - controls length of 'articleModel.Slug','articleModel.Title' when creating, updating an article
var ErrSourceBigSlug = errors.New("slug or title is oversized")

// Article - query

// SaveOneArticle - create new article -> use pgx.Tx with func Transaction
//
// add: autor to favorite new article, tagsID if tag with name exist
// return articleID, error
func (s SQLSource) SaveOneArticle(ctx context.Context, data any) (uint, error) {
	newArticle := data.(models.ArticleModel)
	if len(newArticle.Slug) > models.MaxLenSlug {
		return 0, ErrSourceBigSlug
	}
	insertArticle := func(ctx context.Context) error {
		lineTagsName, _ := common.ArrayToLineForQuery(newArticle.Tags)
		err := s.pTx.Tx.QueryRow(ctx, fmt.Sprintf(`
WITH to_articles AS (
  INSERT INTO articles (slug,
                        title,
                        id_autor,
                        description,
                        body,
                        created_at)
      VALUES ($1, $2, $3, $4, $5, $6)
      RETURNING id),
   to_article_favorite AS (
       INSERT INTO articles_favorite (id_article, id_user)
           VALUES ((SELECT id FROM to_articles), $3)),
   tags_id AS (SELECT id
               FROM tags
               WHERE tag_name IN (%s)),
   to_articles_tags AS (
       INSERT INTO articles_tags (id_article, id_tag)
           SELECT to_articles.id, tags_id.id
           FROM to_articles,
                tags_id
           RETURNING *)

SELECT id
FROM to_articles;`, lineTagsName),
			newArticle.Slug,        //1
			newArticle.Title,       //2
			newArticle.AutorID,     //3
			newArticle.CreatedAt,   //4
			newArticle.Description, //5
			newArticle.Body,        //6
		).Scan(&newArticle.ID)
		if err != nil {
			return ErrSourceAlreadyExists
		}
		return nil
	}
	return newArticle.ID, s.pTx.Transaction(ctx, insertArticle)
}

// NewDataArticle - update article by ID(with ID can change slug) -> use pgx.Tx with func Transaction
//
// removes unique links between article and tag , then add tagsID if tag with name exist from newArticle.Tags
//
// in last check articleID == newArticle.ID -> (check article exists)
func (s SQLSource) NewDataArticle(ctx context.Context, data any) error {
	updateArticle := func(ctx context.Context) error {
		newArticle := data.(models.ArticleModel)
		_, err := s.pTx.Tx.Exec(ctx, `
DELETE
FROM articles_tags
WHERE id_article = $1;`, newArticle.ID)
		if err != nil {
			return err
		}
		lineTagsName, _ := common.ArrayToLineForQuery(newArticle.Tags)
		articleID := uint(0) // check id
		err = s.pTx.Tx.QueryRow(ctx, fmt.Sprintf(`
WITH to_articles AS (
  UPDATE articles
      SET slug = $2,
          title = $3,
          description = $4,
          body = $5,
          updated_at = $6
      WHERE id = $1
      RETURNING id),
   to_tags AS (SELECT id
               FROM tags
               WHERE tag_name IN (%s)),
   to_articles_tags AS (
       INSERT INTO articles_tags (id_article, id_tag)
           SELECT $1, to_tags.id
           FROM to_tags)
SELECT id
FROM to_articles;`, lineTagsName),
			newArticle.ID,          //1
			newArticle.Slug,        //2
			newArticle.Title,       //3
			newArticle.Description, //4
			newArticle.Body,        //5
			newArticle.UpdatedAt,   //6
		).Scan(&articleID)
		if err != nil {
			return ErrSourceNotFound
		}
		return nil
	}
	return s.pTx.Transaction(ctx, updateArticle)
}

// EndArticleLife - DELETE article by ID(need check autorID befor remove) -> use pgx.Tx with func Transaction
//
// first removes from articles and (check article exists),
// if delID == 0 -> Rollback
// else removes from articles_tags, articles_favorite, comments without check, exept error
func (s SQLSource) EndArticleLife(ctx context.Context, data any) error {
	deleteArticle := func(ctx context.Context) error {
		articleID := data.(uint)
		delID := uint(0)
		err := s.pTx.Tx.QueryRow(ctx, `
DELETE
FROM articles
WHERE id = $1
RETURNING id;`, articleID).Scan(&delID)
		if err != nil {
			return ErrSourceNotFound
		}
		_, err = s.pTx.Tx.Exec(ctx, `
DELETE
FROM articles_tags
WHERE id_article = $1;`, articleID)
		if err != nil {
			return err
		}
		_, err = s.pTx.Tx.Exec(ctx, `
DELETE
FROM articles_favorite
WHERE id_article = $1;`, articleID)
		if err != nil {
			return err
		}
		_, err = s.pTx.Tx.Exec(ctx, `
DELETE
FROM comments
WHERE id_article = $1;`, articleID)
		if err != nil {
			return err
		}
		return nil
	}
	return s.pTx.Transaction(ctx, deleteArticle)
}

// FindOneArticle - find article by slug
func (s SQLSource) FindOneArticle(ctx context.Context, data any) (models.ArticleModel, error) {
	slug := data.(string)
	row := s.pTx.Pool.QueryRow(ctx, `
SELECT id,
     slug,
     title,
     id_autor,
     description,
     body,
     created_at,
     updated_at
     ARRAY(SELECT tag_name
           FROM tags
           WHERE tags.id IN (SELECT id_tag
                             FROM articles_tags
                             WHERE id_article = articles.id)) AS article_tag_list,
     (SELECT COUNT(id_user)
      FROM articles_favorite
      WHERE id_article = articles.id)                         AS article_favorite
FROM articles
WHERE slug = $1;`, slug)
	return scanArcticleModel[pgx.Row](row)
}

func scanArcticleModel[T SQLScan](rows T) (models.ArticleModel, error) {
	var (
		article   models.ArticleModel
		updatedAt sql.NullTime
	)
	if err := rows.Scan(
		&article.ID,
		&article.Slug,
		&article.Title,
		&article.AutorID,
		&article.Description,
		&article.Body,
		&article.CreatedAt,
		&updatedAt,
		&article.Tags,
		&article.NumberOfLikes,
	); err != nil {
		return article, ErrSourceNotFound
	}
	if updatedAt.Valid {
		article.UpdatedAt = &updatedAt.Time
	}
	return article, nil
}

// FindOneArticleList - find article list by 'models.ArticleProperty'
//
// property is exist -> add: command to query, arg to 'array' of 'args'
func (s SQLSource) FindArticleList(ctx context.Context, data any) ([]models.ArticleModel, error) {
	property := data.(models.ArticleProperty)
	tagsLine, _ := common.ArrayToLineForQuery(property.Tags)
	query := strings.Builder{}
	query.WriteString(`
SELECT id,
     slug,
     title,
     id_autor,
     description,
     body,
     created_at,
     updated_at
     ARRAY(SELECT tag_name
           FROM tags
           WHERE tags.id IN (SELECT id_tag
                             FROM articles_tags
                             WHERE id_article = articles.id)) AS article_tag_list,
     (SELECT COUNT(id_user)
      FROM articles_favorite
      WHERE id_article = articles.id)                         AS article_favorite
FROM articles`)
	args := []any{}
	numberOfArgs := 1
	where := false
	if login := property.AutorName; login != "" {
		where = whereANDOR(&query, where, "")
		query.WriteString(fmt.Sprintf(`
id_autor = (SELECT id
          FROM users
          WHERE login = $%d
          LIMIT 1)`, numberOfArgs))
		numberOfArgs++
		args = append(args, property, login)
	}
	if len(tagsLine) != 0 {
		where = whereANDOR(&query, where, "AND")
		query.WriteString(fmt.Sprintf(`
id IN (SELECT DISTINCT id_article
     FROM articles_tags
     WHERE id_tag IN (SELECT id
                      FROM tags
                      WHERE tag_name IN (%s))) --tagList`, tagsLine))
	}
	if property.Favorited {
		where = whereANDOR(&query, where, "AND")
		query.WriteString(fmt.Sprintf(`
id IN (SELECT id_article
     FROM articles_favorite
     WHERE id_user = $%d) -- current user`, numberOfArgs))
		numberOfArgs++
		userID := ctx.Value(flag.KeyUserID).(uint)
		args = append(args, userID)
	}
	if !property.IsRangeZero() {
		where = whereANDOR(&query, where, "AND")
		query.WriteString(fmt.Sprintf(`created_at BETWEEN $%d AND $%d`, numberOfArgs, numberOfArgs+1))
		numberOfArgs += 2
		args = append(args, property.StartDate, property.EndDate)
	}
	limit := false
	if property.IsLimit() {
		limit = true
		query.WriteString(fmt.Sprintf("\nLIMIT $%d", numberOfArgs))
		numberOfArgs++
		args = append(args, property.Limit)
	}
	if property.IsOffset() {
		if limit {
			query.WriteString(fmt.Sprintf(" OFFSET $%d", numberOfArgs))
		} else {
			query.WriteString(fmt.Sprintf("\nOFFSET $%d", numberOfArgs))
		}
		args = append(args, property.Offset)
	}
	query.WriteString(`;`)
	rows, err := s.pTx.Pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArcticles(rows)
}

func scanArcticles(rows pgx.Rows) ([]models.ArticleModel, error) {
	var arcticles []models.ArticleModel
	for rows.Next() {
		oneArcticle, err := scanArcticleModel[pgx.Rows](rows)
		if err != nil {
			return nil, err
		}
		arcticles = append(arcticles, oneArcticle)
	}
	return arcticles, rows.Err()
}

// ArticleToFavorite - add article to user favorites -> use pgx.Tx with func Transaction
func (s SQLSource) ArticleToFavorite(ctx context.Context, data any) error {
	insertArticleFavorite := func(ctx context.Context) error {
		slug := data.(string)
		uID := ctx.Value(flag.KeyUserID).(uint)
		articleID, userID := uint(0), uint(0) // check RETURNING
		err := s.pTx.Tx.QueryRow(ctx, `
INSERT INTO articles_favorite (id_article, id_user)
SELECT (SELECT id
       FROM articles
       WHERE slug = $1
       LIMIT 1),
      $2
WHERE EXISTS(SELECT id FROM users WHERE id = $2)
RETURNING *;`, slug, uID).Scan(&articleID, &userID)
		return sourceError(err)
	}
	return s.pTx.Transaction(ctx, insertArticleFavorite)
}

// ArticleUnFovarite - delete article from user favorites -> use pgx.Tx with func Transaction
func (s SQLSource) ArticleUnFovarite(ctx context.Context, data any) error {
	deleteArticleFavorite := func(ctx context.Context) error {
		slug := data.(string)
		uID := ctx.Value(flag.KeyUserID).(uint)
		articleID, userID := uint(0), uint(0) // check RETURNING
		err := s.pTx.Tx.QueryRow(ctx, `
DELETE
FROM articles_favorite
WHERE id_article = (SELECT id
                   FROM articles
                   WHERE slug = $1
   LINIT 1)
 AND id_user = $2;`, slug, uID).Scan(&articleID, &userID)
		return sourceError(err)
	}
	return s.pTx.Transaction(ctx, deleteArticleFavorite)
}

// IsArticleFavorite - check if the article is in user favorites
func (s SQLSource) IsArticleFavorite(ctx context.Context, data any) (bool, error) {
	articleUser := data.([]uint)
	if len(articleUser) != 2 {
		return false, ErrSourceRelationship
	}
	favorite := false
	err := s.pTx.Pool.QueryRow(ctx, `
SELECT EXISTS(SELECT id_article, id_user
           FROM articles_favorite
           WHERE id_article = $1
             AND id_user = $2)`,
		articleUser[0], articleUser[1]).Scan(&favorite)
	return favorite, err
}

// Tag query
//
// describes: interfaces: CommentNew, CommentFindm, CommentChange, CommentRemove
//				  for struct - TagModel
//
// look -> internal/models/tag.go

// SaveOneTag - create new article -> use pgx.Tx with func Transaction
func (s SQLSource) SaveOneTag(ctx context.Context, data any) (uint, error) {
	newTag := data.(models.TagModel)
	insertTag := func(ctx context.Context) error {
		err := s.pTx.Tx.QueryRow(ctx, `
INSERT INTO tags(tag_name, id_tag_maker, created_at)
SELECT $1, $2, $3
WHERE EXISTS(SELECT id FROM users WHERE id = $2)
RETURNING id;`, newTag.Name, newTag.AutorID, newTag.CreatedAt).Scan(newTag.ID)
		return sourceError(err)
	}
	return newTag.ID, s.pTx.Transaction(ctx, insertTag)
}

// EndTagLife - DELETE tag by ID(need check autorID befor remove) -> use pgx.Tx with func Transaction
//
// first removes from tags and (check article exists)
//
// removes from articles_tags without check, exept error
func (s SQLSource) EndTagLife(ctx context.Context, data any) error {
	deleteTag := func(ctx context.Context) error {
		tagID := data.(uint)
		delID := uint(0)
		err := s.pTx.Tx.QueryRow(ctx, `
DELETE 
FROM tags
WHERE id = $1
RETURNING id;`, tagID).Scan(&delID)
		if err != nil {
			return ErrSourceNotFound
		}
		_, err = s.pTx.Tx.Exec(ctx, `
DELETE
FROM articles_tags
WHERE id_tag = $1;`, tagID)
		return err
	}
	return s.pTx.Transaction(ctx, deleteTag)
}

// FindOneTag - find tag by name
func (s SQLSource) FindOneTag(ctx context.Context, data any) (models.TagModel, error) {
	tagName := data.(string)
	row := s.pTx.Pool.QueryRow(ctx, `
SELECT id,
       tag_name,
       id_tag_maker,
       created_at
FROM tags
WHERE tag_name = $1
LIMIT 1;`, tagName)
	return scanTagModel[pgx.Row](row)
}

func scanTagModel[T SQLScan](r T) (models.TagModel, error) {
	var tag models.TagModel
	if err := r.Scan(
		&tag.ID,
		&tag.Name,
		&tag.AutorID,
		&tag.CreatedAt,
	); err != nil {
		return tag, ErrSourceNotFound
	}
	return tag, nil
}

// TagsList - find tag list by 'models.TagPropery'
//
// property -> look up 'FindArticleList'
func (s SQLSource) TagsList(ctx context.Context, data any) ([]models.TagModel, error) {
	property := data.(models.TagPropery)
	query := strings.Builder{}
	query.WriteString(`
SELECT id,
       tag_name,
       id_tag_maker,
       created_at
FROM tags`)
	args := []any{}
	numberOfArg := 1
	where := false
	if login := property.AutorName; login != "" {
		where = whereANDOR(&query, where, "")
		query.WriteString(fmt.Sprintf(`
id_tag_maker = (SELECT id
               FROM users
               WHERE login = $%d
               LIMIT 1)`, numberOfArg))
		numberOfArg++
		args = append(args, login)
	}
	if !property.IsRangeZero() {
		where = whereANDOR(&query, where, "AND")
		query.WriteString(fmt.Sprintf(`created_at BETWEEN $%d AND $%d`, numberOfArg, numberOfArg+1))
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
	defer rows.Close()
	return scanTags(rows)
}

func scanTags(rows pgx.Rows) ([]models.TagModel, error) {
	var tags []models.TagModel
	for rows.Next() {
		tag, err := scanTagModel[pgx.Rows](rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// Comment - query
//
// describe: interfaces: CommentNew, CommentFindm, CommentChange, CommentRemove
//           for struct - CommentModel
//
// look -> internal/models/comment.go

// SaveOneComment - create new comment -> use pgx.Tx with func Transaction
//
// return commentID, error
func (s SQLSource) SaveOneComment(ctx context.Context, data any) (uint, error) {
	newComment := data.(models.CommentModel)
	insertComment := func(ctx context.Context) error {
		err := s.pTx.Pool.QueryRow(ctx, `
INSERT INTO comments (id_article,
                      id_autor,
                      body,
                      created_at)
VALUES ((SELECT id
         FROM articles
         WHERE slug = $1
         LIMIT 1), $2, $3, $4)
RETURNING id;`,
			newComment.ArticleSlug, //1
			newComment.AutorID,     //2
			newComment.Body,        //3
			newComment.CreatedAt,   //4
		).Scan(&newComment.ID)
		if err != nil {
			return ErrSourceAlreadyExists
		}
		return nil
	}
	return newComment.ID, s.pTx.Transaction(ctx, insertComment)
}

// NewDataComment - update comment by ID -> use pgx.Tx with func Transaction
//
// in last check articleID == newArticle.ID -> (check article exists)
func (s SQLSource) NewDataComment(ctx context.Context, data any) error {
	newComment := data.(models.CommentModel)
	updateComment := func(ctx context.Context) error {
		commentID := uint(0)
		err := s.pTx.Tx.QueryRow(ctx, `
UPDATE comments
SET body       = $2,
    updated_at = $3
WHERE id = $1
RETURNING id;`,
			newComment.ID,        //1
			newComment.Body,      //2
			newComment.UpdatedAt, //3
		).Scan(&commentID)
		if err != nil {
			return ErrSourceNotFound
		}
		return nil
	}
	return s.pTx.Transaction(ctx, updateComment)
}

// EndCommentLife - delete comment by ID ->  use pgx.Tx with func Transaction
func (s SQLSource) EndCommentLife(ctx context.Context, data any) error {
	deleteComment := func(ctx context.Context) error {
		commentID := data.(uint)
		delCommentID := uint(0)
		err := s.pTx.Tx.QueryRow(ctx, `
DELETE
FROM comments
WHERE id = $1
RETURNING id;`, commentID).Scan(&delCommentID)
		if err != nil {
			return ErrSourceNotFound
		}
		return nil
	}
	return s.pTx.Transaction(ctx, deleteComment)
}

// FindOneComment - find comment by ID
func (s SQLSource) FindOneComment(ctx context.Context, data any) (models.CommentModel, error) {
	commentID := data.(uint)
	row := s.pTx.Pool.QueryRow(ctx, `
SELECT c.id,
       c.id_article,
       a.slug,
       c.id_autor,
       c.body,
       c.created_at,
       c.updated_at
FROM comments c
         JOIN articles a ON c.id_article = a.id
WHERE c.id = $1
LIMIT 1;`, commentID)
	return scanCommentModel[pgx.Row](row)
}

func scanCommentModel[T SQLScan](r T) (models.CommentModel, error) {
	var (
		comment   models.CommentModel
		updatedAt sql.NullTime
	)
	if err := r.Scan(
		&comment.ID,
		&comment.ArticleID,
		&comment.ArticleSlug,
		&comment.AutorID,
		&comment.Body,
		&comment.CreatedAt,
		&updatedAt,
	); err != nil {
		return comment, ErrSourceNotFound
	}
	if updatedAt.Valid {
		comment.UpdatedAt = &updatedAt.Time
	}
	return comment, nil
}

// FindCommentList - find comment list by 'models.CommentProperty'
//
// property -> look up FindArticleList
func (s SQLSource) FindCommentList(ctx context.Context, data any) ([]models.CommentModel, error) {
	property := data.(models.CommentProperty)
	query := strings.Builder{}
	query.WriteString(`
SELECT c.id,
       c.id_article,
       a.slug,
       c.id_autor,
       c.body,
       c.created_at,
       c.updated_at
FROM comments c
         JOIN articles a ON c.id_article = a.id`)
	args := []any{}
	numberOfArg := 1
	where := false
	if slug := property.ArcticleSlug; slug != "" {
		where = whereANDOR(&query, where, "")
		query.WriteString(fmt.Sprintf(`a.slug = $%d`, numberOfArg))
		numberOfArg++
		args = append(args, slug)
	}
	if login := property.AutorName; login != "" {
		where = whereANDOR(&query, where, "AND")
		query.WriteString(fmt.Sprintf(`
c.id_autor = (SELECT id
              FROM users
              WHERE login = $%d
              LIMIT 1)`, numberOfArg))
		numberOfArg++
		args = append(args, login)
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
	defer rows.Close()
	return scanComments(rows)
}

func scanComments(rows pgx.Rows) ([]models.CommentModel, error) {
	var comments []models.CommentModel
	for rows.Next() {
		comment, err := scanCommentModel[pgx.Rows](rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}
