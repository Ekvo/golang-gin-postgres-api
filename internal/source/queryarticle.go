package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// ErrSourceTagsCount - защищает от писка незарегестрированных тегов при создании, обновлении статьи
var ErrSourceTagsCount = errors.New("count tags not equal")

// ErrSourceBigSlug - контралирует длину 'articleModel.Slug','articleModel.Title' при создании, обновлении статьи
var ErrSourceBigSlug = errors.New("slug or title is oversized")

// ErrSourceNoUpdate - для маркировки отрицательного обновления, создания записей
var ErrSourceNoUpdate = errors.New("no update or delete")

func (s *SQLSource) SaveOneArticle(ctx context.Context, data any) (uint, error) {
	articleModel := data.(models.ArticleModel)
	if len(articleModel.Slug) > models.MaxLenSlug {
		return 0, ErrSourceBigSlug
	}
	insertArticle := func(ctx context.Context) error {
		lineTagsName, countTags := common.ArrayToLineForQuery(articleModel.Tags)
		newRowsInArticleTags := 0

		err := s.sourceDBTX.Tx.QueryRowContext(ctx, fmt.Sprintf(`
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
		if err != nil {
			return err
		}
		// проверяем на отсутсвие инородных тегов
		if newRowsInArticleTags != countTags {
			return ErrSourceTagsCount
		}
		return nil
	}
	return articleModel.ID, s.sourceDBTX.Transaction(ctx, insertArticle)
}

func (s *SQLSource) NewDataArticle(ctx context.Context, data any) error {
	updateArticle := func(ctx context.Context) error {
		articleModel := data.(models.ArticleModel)
		lineTagsName, countTags := common.ArrayToLineForQuery(articleModel.Tags)
		newRowsInArticleTags := 0

		// удаляем старую связь тегов и статьи	в последующем создадим новую
		_, err := s.sourceDBTX.Tx.ExecContext(ctx, `
	DELETE FROM articles_tags
	WHERE id_article = $1;`, articleModel.ID)
		if err != nil {
			return err
		}
		err = s.sourceDBTX.Tx.QueryRowContext(ctx, fmt.Sprintf(`
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
			return err
		}
		if newRowsInArticleTags != countTags {
			return ErrSourceTagsCount
		}
		return nil
	}
	return s.sourceDBTX.Transaction(ctx, updateArticle)
}

func (s *SQLSource) EndArticleLife(ctx context.Context, data any) error {
	deleteArticle := func(ctx context.Context) error {
		articleModel := data.(models.ArticleModel)
		_, err := s.sourceDBTX.Tx.ExecContext(ctx, `
DELETE FROM comments_body
       WHERE id_comment IN (SELECT id FROM comments WHERE id_article = $1);`, articleModel.ID)
		if err != nil {
			return err
		}
		_, err = s.sourceDBTX.Tx.ExecContext(ctx, `
DELETE FROM comments WHERE id_article = $1;`, articleModel.ID)
		if err != nil {
			return err
		}
		_, err = s.sourceDBTX.Tx.ExecContext(ctx, `
DELETE FROM articles_tags WHERE id_article = $1;`, articleModel.ID)
		if err != nil {
			return err
		}
		_, err = s.sourceDBTX.Tx.ExecContext(ctx, `
DELETE FROM articles_favorite WHERE id_article = $1;`, articleModel.ID)
		if err != nil {
			return err
		}
		_, err = s.sourceDBTX.Tx.ExecContext(ctx, `
DELETE FROM articles_body WHERE id_article = $1;`, articleModel.ID)
		if err != nil {
			return err
		}
		_, err = s.sourceDBTX.Tx.ExecContext(ctx, `
DELETE FROM articles WHERE id = $1;`, articleModel.ID)
		if err != nil {
			return err
		}
		return nil
	}
	return s.sourceDBTX.Transaction(ctx, deleteArticle)
}

func (s *SQLSource) FindOneArticle(ctx context.Context, data any) (models.ArticleModel, error) {
	slug := data.(string)
	row := s.sourceDBTX.DB.QueryRowContext(ctx, `
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

func (s *SQLSource) FindArticleList(ctx context.Context, data any) ([]models.ArticleModel, error) {
	arcticleProperty := data.(models.ArticleProperty)
	userID := ctx.Value(models.KeyUserID).(uint)
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
WHERE a.created_at BETWEEN '%s' AND '%s'
    (
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
                              WHERE id_user = %d) --find in non-featured articles
        END
    )
LIMIT %d OFFSET %d;`,
		arcticleProperty.StartDate.Format(time.DateTime),
		arcticleProperty.EndDate.Format(time.DateTime),
		len(arcticleProperty.AutorName), //first case
		arcticleProperty.AutorName,      //first case
		len(tagsLine),                   //second case
		tagsLine,                        //second case
		arcticleProperty.Favorited,      //third case
		userID,                          //third case
		userID,                          //third case
		arcticleProperty.Limit,
		arcticleProperty.Offset,
	)
	rows, err := s.sourceDBTX.DB.QueryContext(ctx, queryBody)
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

func (s *SQLSource) ArticleToFavorite(ctx context.Context, data any) error {
	insertArticleFavorite := func(ctx context.Context) error {
		userID := ctx.Value(models.KeyUserID).(uint)
		slug := data.(string)
		_, err := s.sourceDBTX.DB.ExecContext(ctx, `
INSERT INTO articles_favorite (id_article,id_user)
VALUES((SELECT id FROM articles WHERE slug = $1 LIMIT 1),$2);`, slug, userID)
		return err
	}
	return s.sourceDBTX.Transaction(ctx, insertArticleFavorite)
}

func (s *SQLSource) ArticleUnFovarite(ctx context.Context, data any) error {
	deleteArticleFavorite := func(ctx context.Context) error {
		userID := ctx.Value(models.KeyUserID).(uint)
		slug := data.(string)
		_, err := s.sourceDBTX.DB.ExecContext(ctx, `
DELETE FROM articles_favorite
       WHERE id_article = (SELECT id FROM articles WHERE slug = $1 LINIT 1)
         AND id_user = $2;`, slug, userID)
		return err
	}
	return s.sourceDBTX.Transaction(ctx, deleteArticleFavorite)
}

func (s *SQLSource) SaveOneTag(ctx context.Context, data any) (uint, error) {
	tagModel := data.(models.TagModel)
	insertTag := func(ctx context.Context) error {
		return s.sourceDBTX.DB.QueryRowContext(ctx, `
INSERT INTO tags(tag_name,id_tag_maker,created_at)
VALUES($1,$2,$3)
RETURNING id;`, tagModel.Name, tagModel.AutorID, tagModel.CreatedAt).Scan(tagModel.ID)
	}
	return tagModel.ID, s.sourceDBTX.Transaction(ctx, insertTag)
}

func (s *SQLSource) FindOneTag(ctx context.Context, data any) (models.TagModel, error) {
	tagName := data.(string)
	row := s.sourceDBTX.DB.QueryRowContext(ctx, `
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

func (s *SQLSource) TagsList(ctx context.Context, data any) ([]models.TagModel, error) {
	tagProperty := data.(models.TagPropery)
	rows, err := s.sourceDBTX.DB.QueryContext(ctx, `
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

func (s *SQLSource) SaveOneComment(ctx context.Context, data any) (uint, error) {
	commentModel := data.(models.CommentModel)
	insertComment := func(ctx context.Context) error {
		return s.sourceDBTX.DB.QueryRowContext(ctx, `
WITH to_comments AS(
    INSERT INTO comments(id_article,id_autor,created_at)
        VALUES((SELECT id
                FROM articles
                WHERE slug = $1
                LIMIT 1),$2,$3)
        RETURNING id
), to_comments_body AS (
    INSERT INTO comments_body (id_comment, body)
        VALUES ((SELECT id FROM to_comments), $4)
)
SELECT id FROM to_comments;`,
			commentModel.ArticleSlug, //1
			commentModel.AutorID,     //2
			commentModel.Body,        //3
			commentModel.CreatedAt,   //4
		).Scan(&commentModel.ID)
	}
	return commentModel.ID, s.sourceDBTX.Transaction(ctx, insertComment)
}

func (s *SQLSource) NewDataComment(ctx context.Context, data any) error {
	commentModel := data.(models.CommentModel)
	updateComment := func(ctx context.Context) error {
		approve := false
		err := s.sourceDBTX.Tx.QueryRowContext(ctx, `
WITH to_comments_body AS(
    UPDATE comments_body
        SET body = $3
        WHERE id_comment = $1
           AND ($2 = (SELECT id_autor --check autor
                     FROM comments
                     WHERE id = $1)
                    OR  (SELECT access --check admin status
                         FROM users_access
                         WHERE id_user = $2) > '3' --admin status start with 4 
               )
), to_comments AS(
    UPDATE comments
        SET updated_at = $4
        WHERE id = $1
           AND ($2 = (SELECT id_autor
                     FROM comments
                     WHERE id = $1)
                    OR  (SELECT access
                         FROM users_access
                         WHERE id_user = $2) > '3' 
               )
           RETURNING *
)
SELECT EXISTS(SELECT updated_at FROM to_comments);`,
			commentModel.ID,        //1
			commentModel.AutorID,   //2
			commentModel.Body,      //3
			commentModel.UpdatedAt, //4
		).Scan(&approve)
		if err != nil || !approve {
			return ErrSourceNoUpdate
		}
		return nil
	}
	return s.sourceDBTX.Transaction(ctx, updateComment)
}

func (s *SQLSource) EndCommentLife(ctx context.Context, data any) error {
	commentModel := data.(models.CommentModel)
	deleteComment := func(ctx context.Context) error {
		delCommentID := uint(0)
		err := s.sourceDBTX.Tx.QueryRowContext(ctx, `
DELETE FROM comments_bodY
WHERE id_comment = $1
  AND ($2 = (SELECT id_autor
             FROM comments
             WHERE id = $1)
           OR  (SELECT access
                FROM users_access
                WHERE id_user = $2) > '3' 
               )
RETURNING id_comment;`, commentModel.ID, commentModel.AutorID).Scan(&delCommentID)
		if err != nil || delCommentID != commentModel.ID {
			return ErrSourceNoUpdate
		}
		err = s.sourceDBTX.Tx.QueryRowContext(ctx, `
DELETE FROM comments
WHERE id = $1
  AND (id_autor = $2             
           OR  (SELECT access
                FROM users_access
                WHERE id_user = $2) > '3' 
               )
RETURNING id;
`, commentModel.ID, commentModel.AutorID).Scan(&delCommentID)
		if err != nil || delCommentID != commentModel.ID {
			return ErrSourceNoUpdate
		}
		return nil
	}
	return s.sourceDBTX.Transaction(ctx, deleteComment)
}

func (s *SQLSource) FindOneComment(ctx context.Context, data any) (models.CommentModel, error) {
	commentID := data.(uint)
	row := s.sourceDBTX.DB.QueryRowContext(ctx, `
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

func (s *SQLSource) FindCommentList(ctx context.Context, data any) ([]models.CommentModel, error) {
	commentProperty := data.(models.CommentProperty)
	rows, err := s.sourceDBTX.DB.QueryContext(ctx, `
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
