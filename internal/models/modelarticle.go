package models

import (
	"context"
	"time"

	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
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

// ArticleProperty - содержит набот свойсв - для поиска множества - 'ArticleModel'
type ArticleProperty struct {
	// список тегов
	// пустой строка -> поиск по всем тегам
	Tags ArcticleTags

	// автор в единсвенном числе
	// пустая строка -> поиск по всем авторам
	AutorName string

	// находится ли статья в избранном списке для текущего пользователя
	Favorited bool

	// отрезок времени, в течение которого была создан объект
	common.TimeRange

	// см. пакет pkg/common
	common.LimitOffset
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

// TagModel - свойсва тега в базе
type TagModel struct {
	ID uint

	// максимальный размер - 25
	Name string

	AutorID uint

	CreatedAt time.Time
}

// TagPropery - свойсва для поиска набора 'TagModel'
type TagPropery struct {
	AutorName string

	common.TimeRange

	common.LimitOffset
}

type TagUpdate interface {
	SaveOneTag(ctx context.Context, data any) (uint, error)
}

type TagFind interface {
	FindOneTag(ctx context.Context, data any) (TagModel, error)

	TagsList(ctx context.Context, data any) ([]TagModel, error)
}

type TagUpdateFind interface {
	TagUpdate
	TagFind
}

type TagWithAutor interface {
	TagUpdateFind
	UserApproveFollowing
}

type CommentModel struct {
	ID uint

	ArticleID   uint
	ArticleSlug string

	AutorID uint

	CreatedAt time.Time
	UpdatedAt *time.Time

	//maxlenght 2048
	Body string
}

type CommentProperty struct {
	ArcticleSlug string

	//автор комментария
	AutorName string

	// см. 'ArticleProperty'
	common.TimeRange

	// см. 'ArticleProperty'
	common.LimitOffset
}

type CommentUpdate interface {
	SaveOneComment(ctx context.Context, data any) (uint, error)

	NewDataComment(ctx context.Context, data any) error

	EndCommentLife(ctx context.Context, data any) error
}

type CommentFind interface {
	FindOneComment(ctx context.Context, data any) (CommentModel, error)
	FindCommentList(ctx context.Context, data any) ([]CommentModel, error)
}

type CommentUpdateFind interface {
	CommentUpdate
	CommentFind
}
type CommentWihtAutor interface {
	CommentUpdateFind
	UserApproveFollowing
}
