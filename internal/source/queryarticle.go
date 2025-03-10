package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services"
	"github.com/lib/pq"
	"log"
	"time"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// ErrSourceTagsCount - защищает от писка незарегестрированных тегов при создании, обновлении статьи
var ErrSourceTagsCount = errors.New("count tags not equal")

// ErrSourceBigSlug - контралирует длину 'articleModel.Slug','articleModel.Title' при создании, обновлении статьи
var ErrSourceBigSlug = errors.New("slug or title is oversized")

func (s SQLSource) SaveOneArticle(ctx context.Context, data any) (uint, error) {
	articleModel := data.(models.ArticleModel)
	if len(articleModel.Slug) > models.MaxLenSlug {
		return 0, ErrSourceBigSlug

	}
	lineTagsName, countTags := common.ArrayToLineForQuery(articleModel.Tags)
	newRowsInArticleTags := 0
	tx, err := s.sourceDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	err = s.sourceDB.QueryRowContext(ctx, fmt.Sprintf(`
WITH to_articles AS(
    INSERT INTO articles (slug,
                          title,
                          id_autor,
                          created_at)
    VALUES($1,$2,$3,now())
    RETURNING id
), to_articles_body AS(
    INSERT INTO article_body(id_article,
                             discription,
                             body)
        VALUES((SELECT id FROM to_articles),$4,$5)
), tags_id AS(
    SELECT id FROM tags WHERE tag_name IN (%s)
), to_articles_tags AS(
    INSERT INTO articles_tags(id_article,id_tag)
        SELECT to_articles.id,tags_id.id
        FROM to_articles,tags_id
        RETURNING *
)
SELECT id,(SELECT COUNT(*) FROM to_articles_tags)
FROM to_articles;`, lineTagsName),
		articleModel.Slug,        //1
		articleModel.Title,       //2
		articleModel.AutorID,     //3
		articleModel.Description, //4
		articleModel.Body,        //5
	).Scan(&articleModel.ID, &newRowsInArticleTags)

	if newRowsInArticleTags != countTags {
		return 0, SQLRollback(tx, ErrSourceTagsCount)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return articleModel.ID, nil
}

func (s SQLSource) NewDataArticle(ctx context.Context, data any) error {
	articleModel := data.(models.ArticleModel)
	lineTagsName, countTags := common.ArrayToLineForQuery(articleModel.Tags)
	newRowsInArticleTags := 0

	tx, err := s.sourceDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// удаляем старую связь тегов и статьи	в последующем создадим новую
	_, err = s.sourceDB.ExecContext(ctx, `
	DELETE FROM articles_tags
	WHERE id_article = $1;`, articleModel.ID)
	if err != nil {
		return SQLRollback(tx, err)
	}
	err = s.sourceDB.QueryRowContext(ctx, fmt.Sprintf(`
WITH to_articles AS(
    UPDATE articles
        SET slug = $2,
            title = $3,
            updated_at = $4
        WHERE id = $1
), to_articles_body AS(
    UPDATE articles_body
        SET description = $5,
            body = $6'
        WHERE id_article = $1
), to_tags AS(
SELECT id FROM tags WHERE tag_name IN (%s)
),to_articles_tags AS(
    INSERT INTO articles_tags (id_article,id_tag)
        SELECT $1,to_tags.id
        FROM to_tags
        RETURNING *
)
SELECT COUNT(*) FROM to_articles_tags;`, lineTagsName),
		articleModel.ID,          //1
		articleModel.Slug,        //2
		articleModel.Title,       //3
		articleModel.UpdatedAt,   //4
		articleModel.Description, //5
		articleModel.Body,        //6
	).Scan(&newRowsInArticleTags)

	if err != nil {
		return SQLRollback(tx, err)
	}
	if newRowsInArticleTags != countTags {
		return SQLRollback(tx, ErrSourceTagsCount)
	}
	return tx.Commit()
}

func (s SQLSource) EndArticleLife(ctx context.Context, data any) error {
	articleModel := data.(models.ArticleModel)

	tx, err := s.sourceDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = s.sourceDB.ExecContext(ctx, `
DELETE FROM comments_body
       WHERE id_comment IN (SELECT id FROM comments WHERE id_article = $1);`, articleModel.ID)
	if err != nil {
		return SQLRollback(tx, err)
	}
	_, err = s.sourceDB.ExecContext(ctx, `
DELETE FROM comments WHERE id_article = $1;`, articleModel.ID)
	if err != nil {
		return SQLRollback(tx, err)
	}
	_, err = s.sourceDB.ExecContext(ctx, `
DELETE FROM articles_tags WHERE id_article = $1;`, articleModel.ID)
	if err != nil {
		return SQLRollback(tx, err)
	}
	_, err = s.sourceDB.ExecContext(ctx, `
DELETE FROM articles_favorite WHERE id_article = $1;`, articleModel.ID)
	if err != nil {
		return SQLRollback(tx, err)
	}
	_, err = s.sourceDB.ExecContext(ctx, `
DELETE FROM articles_body WHERE id_article = $1;`, articleModel.ID)
	if err != nil {
		return SQLRollback(tx, err)
	}
	_, err = s.sourceDB.ExecContext(ctx, `
DELETE FROM articles WHERE id = $1;`, articleModel.ID)
	if err != nil {
		return SQLRollback(tx, err)
	}
	return tx.Commit()
}

func (s SQLSource) FindOneArticle(ctx context.Context, data any) (models.ArticleModel, error) {
	slug := data.(string)
	row := s.sourceDB.QueryRowContext(ctx, `
SELECT a.id,
       a.slug,
       a.title,
       a.id_autor,
       u.login,
       a.created_at,
       a.updated_at,
       ab.description,
       ab.body,
       ARRAY(SELECT tag_name
             FROM tags
             WHERE tags.id IN (SELECT id_tag
                               FROM articles_tags
                               WHERE id_article = a.id)) AS tag_list
FROM articles a
    JOIN articles_body ab ON a.id = ab.id_article
    JOIN users u ON a.id_autor = u.id
WHERE a.slug = $1;`, slug)
	arcticle, err := scanArcticleModel[*sql.Row](row)
	if err != nil {
		return arcticle, err
	}
	return arcticle, err
}

func scanArcticleModel[T SQLRowsRowScan](rows T) (models.ArticleModel, error) {
	var (
		article   models.ArticleModel
		updatedAt sql.NullTime
		tags      pq.StringArray
	)
	if err := rows.Scan(
		&article.ID,
		&article.Slug,
		&article.Title,
		&article.AutorID,
		&article.AutorName,
		&article.CreatedAt,
		&updatedAt,
		&article.Description,
		&article.Body,
		&tags,
	); err != nil {
		return article, err
	}
	if updatedAt.Valid {
		article.UpdatedAt = &updatedAt.Time
	}
	if len(tags) != 0 {
		article.Tags = models.ArcticleTags(tags)
	}
	return article, nil
}

func (s SQLSource) FindArticleList(ctx context.Context, data any) ([]models.ArticleModel, error) {
	arcticleProperty := data.(models.ArticleProperty)
	userID := ctx.Value(services.UserID).(uint)
	tagsLine, _ := common.ArrayToLineForQuery(arcticleProperty.Tags)
	queryBody := fmt.Sprintf(`
SELECT a.id,
       a.slug,
       a.title,
       a.id_autor,
       u.login,
       a.created_at,
       a.updated_at,
       ab.description,
       ab.body,
       ARRAY(SELECT tag_name
             FROM tags
             WHERE tags.id IN (SELECT id_tag
                               FROM articles_tags
                               WHERE id_article = a.id)) AS tag_list
FROM articles a
    JOIN articles_body ab ON a.id = ab.id_article
    JOIN users u ON a.id_autor = u.id
WHERE  (
    CASE --1  
        WHEN %d > 0 --autor name not empty
            THEN a.id_autor = (SELECT id
                               FROM users
                               WHERE login = '%s' --autor name
                               LIMIT 1)
        ELSE TRUE = TRUE --all autors
        END
    ) AND (
        CASE --2
            WHEN %d > 0 --tagList not empty
                THEN a.id IN (SELECT DISTINCT id_article
                              FROM articles_tags
                              WHERE id_tag IN (SELECT id
                                               FROM tags
                                               WHERE tag_name IN (%s))) --tagList
            ELSE TRUE = TRUE --all tags
            END
    ) AND (
        CASE --3
            WHEN TRUE = %t --if true - find in favorited artciles
                THEN a.id IN (SELECT id_article
                              FROM articles_favorite
                              WHERE id_user = %d) -- current user
            ELSE a.id NOT IN (SELECT id_article
                              FROM articles_favorite
                              WHERE id_user = 1) --find in non-featured articles
        END
    )
LIMIT %d OFFSET %d;`,
		len(arcticleProperty.AutorName), //first case
		arcticleProperty.AutorName,      //first case
		len(tagsLine),                   //second case
		tagsLine,                        //second case
		arcticleProperty.Favorited,      //third case
		userID,                          //third case
		arcticleProperty.Limit,
		arcticleProperty.Offset,
	)
	rows, err := s.sourceDB.QueryContext(ctx, queryBody)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("sql.Rows error - %w", err)
		}
	}()
	return scanArcticles(rows)
}

func scanArcticles(rows *sql.Rows) ([]models.ArticleModel, error) {
	var arcticles []models.ArticleModel
	for rows.Next() {
		oneArcticle, err := scanArcticleModel[*sql.Rows](rows)
		if err != nil {
			return nil, err
		}
		arcticles = append(arcticles, oneArcticle)
	}
	return arcticles, rows.Err()
}

func (s SQLSource) ArticleToFavorite(ctx context.Context, data any) error {
	userID := ctx.Value(services.UserID).(uint)
	slug := data.(string)
	_, err := s.sourceDB.ExecContext(ctx, `
INSERT INTO articles_favorite (id_article,id_user)
VALUES((SELECT id FROM articles WHERE slug = $1 LIMIT 1),$2);`, slug, userID)
	return err
}

func (s SQLSource) ArticleUnFovarite(ctx context.Context, data any) error {
	userID := ctx.Value(services.UserID).(uint)
	slug := data.(string)
	_, err := s.sourceDB.ExecContext(ctx, `
DELETE FROM articles_favorite
       WHERE id_article = (SELECT id FROM articles WHERE slug = $1 LINIT 1)
         AND id_user = $2;`, slug, userID)
	return err
}

func (s SQLSource) SaveOneTag(ctx context.Context, data any) (uint, error) {
	tagModel := data.(models.TagModel)
	err := s.sourceDB.QueryRowContext(ctx, `
INSERT INTO tags(tag_name,id_tag_maker,created_at)
VALUES($1,$2,$3)
RETURNING id;`, tagModel.Name, tagModel.AutorID, tagModel.CreatedAt).Scan(tagModel.ID)
	return tagModel.ID, err
}

func (s SQLSource) FindOneTag(ctx context.Context, data any) (models.TagModel, error) {
	tagName := data.(string)
	row := s.sourceDB.QueryRowContext(ctx, `
SELECT t.id,
       t.tag_name,
       t.id_tag_maker,
       u.login,
       t.created_at
FROM tags t 
    LEFT JOIN users u ON t.id_tag_maker = u.id 
WHERE tag_name = $1
LIMIT 1;`, tagName)
	return scanTagModel[*sql.Row](row)
}

func scanTagModel[T SQLRowsRowScan](r T) (models.TagModel, error) {
	var tag models.TagModel
	err := r.Scan(
		&tag.ID,
		&tag.Name,
		&tag.AutorID,
		&tag.AutorName,
		&tag.CreatedAt,
	)
	return tag, err
}

func (s SQLSource) TagsList(ctx context.Context, data any) ([]models.TagModel, error) {
	tagProperty := data.(models.TagPropery)
	rows, err := s.sourceDB.QueryContext(ctx, `
SELECT t.id,
       t.tag_name,
       t.id_tag_maker,
       u.login,
       t.created_at
FROM tags t 
    LEFT JOIN users u ON t.id_tag_maker = u.id 
WHERE t.created_at BETWEEN $2 AND $3 
  AND(
      CASE
          WHEN lenght($1) > 0 THEN id_tag_maker = (SELECT id
                                                   FROM users 
                                                   WHERE login = $1
                                                   LIMIT 1)
          ELSE TRUE = TRUE 
          END
      ) 
LIMIT $4 OFFSET $5;`,
		tagProperty.AutorName, //1
		tagProperty.StartDate, //2
		tagProperty.EndDate,   //3
		tagProperty.Limit,     //4
		tagProperty.Offset,    //5
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("sql.Rows error - %w", err)
		}
	}()
	return scanTags(rows)
}

func scanTags(rows *sql.Rows) ([]models.TagModel, error) {
	var tags []models.TagModel
	for rows.Next() {
		tag, err := scanTagModel[*sql.Rows](rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (s SQLSource) SaveOneComment(ctx context.Context, data any) (uint, error) {

	return 0, nil
}

func (s SQLSource) NewDataComment(ctx context.Context, data any) error {
	return nil
}

func (s SQLSource) EndCommentLife(ctx context.Context, data any) error {
	return nil
}

func (s SQLSource) FindOneComment(ctx context.Context, data any) (models.CommentModel, error) {
	commentID := data.(uint)
	row := s.sourceDB.QueryRowContext(ctx, `
SELECT c.id,
       c.id_article,
       a.slug,
       c.id_autor,
       u.login,
       c.created_at,
       c.updated_at,
       cb.body
FROM comments c
    JOIN comments_body cb ON c.id = cb.id_comment
    LEFT JOIN articles a ON c.id_article = a.id
    LEFT JOIN users u ON c.id_autor = u.id
WHERE c.id = $1
LIMIT 1;`, commentID)
	return scanCommentModel[*sql.Row](row)
}

func scanCommentModel[T SQLRowsRowScan](r T) (models.CommentModel, error) {
	var (
		comment   models.CommentModel
		updatedAt sql.NullTime
	)
	if err := r.Scan(
		&comment.ID,
		&comment.ArticleID,
		&comment.ArticleSlug,
		&comment.AutorID,
		&comment.AutorName,
		&comment.CreatedAt,
		&updatedAt,
		&comment.Body,
	); err != nil {
		return comment, err
	}
	if updatedAt.Valid {
		comment.UpdatedAt = &updatedAt.Time
	}
	return comment, nil
}

func (s SQLSource) FindCommentList(ctx context.Context, data any) ([]models.CommentModel, error) {
	commentProperty := data.(models.CommentProperty)
	rows, err := s.sourceDB.QueryContext(ctx, `
SELECT c.id,
       c.id_article,
       a.slug,
       c.id_autor,
       u.login,
       c.created_at,
       c.updated_at,
       cb.body
FROM comments c
    JOIN comments_body cb ON c.id = cb.id_comment
    LEFT JOIN articles a ON c.id_article = a.id
    LEFT JOIN users u ON c.id_autor = u.id
WHERE a.slug = $1
  AND (c.created_at BETWEEN $3 AND $4)
  AND (
    CASE
        WHEN length($2) > 0 THEN c.id_autor = (SELECT id
                                       FROM users
                                       WHERE login = $2
                                       LIMIT 1)
        ELSE TRUE = TRUE --all users
        END
      )
ORDER BY c.created_at ASC
LIMIT $5 OFFSET $6;`,
		commentProperty.ArcticleSlug,
		commentProperty.AutorName,
		commentProperty.StartDate.Format(time.DateTime),
		commentProperty.EndDate.Format(time.DateTime),
		commentProperty.Limit,
		commentProperty.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("sql.Rows error - %w", err)
		}
	}()
	return scanComments(rows)
}

func scanComments(rows *sql.Rows) ([]models.CommentModel, error) {
	var comments []models.CommentModel
	for rows.Next() {
		comment, err := scanCommentModel[*sql.Rows](rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}
