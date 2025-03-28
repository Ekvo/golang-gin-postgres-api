package models

import (
	"context"
	"time"
)

// MaxLenSlug - для проверки во время записи в базу данных
// slug.Make(title) - в разных языках дает несоразмерную длину при переводе на анлл.
// см. 'arcticle/validator' - 'func (acv *ArticleCreateValidator) Bind(c *gin.Context) error'
const MaxLenSlug = 50

// ArcticleTags - названия тегов связанных с определенной статьей 'ArticleModel'
type ArcticleTags []string

type ArticleModel struct {
	ID uint

	// максимальный размер - 50
	Slug string

	// максимальный размер - 50
	Title string

	AutorID uint

	CreatedAt time.Time
	UpdatedAt *time.Time

	// максимальный размер - 2048
	Description string

	// максимальный размер - 2048
	Body string

	Tags ArcticleTags

	// Количесво уникальных пользователей  - отметивших статью как избранную
	NumberOfLikes uint
}

// ArticleUpdate - описывает жизненный цикл 'ArticleModel'
type ArticleUpdate interface {
	SaveOneArticle(ctx context.Context, data any) (uint, error)

	// NewDataArticle -
	NewDataArticle(ctx context.Context, data any) error

	// EndArticleLife - удаление данных связанных с 'ArticleModel'
	// инофрмация: комментраии, связь тегов со статьей и непосредсвенно сама статья.
	EndArticleLife(ctx context.Context, data any) error
}

// ArticleFind - шаблон для поиска 'ArticleModel'
type ArticleFind interface {
	FindOneArticle(ctx context.Context, data any) (ArticleModel, error)
	FindArticleList(ctx context.Context, data any) ([]ArticleModel, error)
}

type ArticleUpdateFind interface {
	ArticleUpdate
	ArticleFind
}

// ArticleFavorite - добавление удаление подписки пользователя на статью
type ArticleFavorite interface {
	ArticleToFavorite(ctx context.Context, data any) error
	IsArticleFavorite(ctx context.Context, data any) (bool, error)
	ArticleUnFovarite(ctx context.Context, data any) error
}

type ArticleWithAutor interface {
	ArticleUpdateFind
	ArticleFavorite
	UserApproveFollowing
}
