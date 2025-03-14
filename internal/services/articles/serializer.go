package articles

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/users"
)

type ArcticleSerializer struct {
	C *gin.Context
	models.ArticleModel
}

type ArcticleResponse struct {
	ID          uint     `json:"id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	Tags        []string `json:"tag_list"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at,omitempty"`

	Autor users.ProfileResponse `json:"autor_profile"`

	NumberOfLikes uint `json:"number_of_likes"`

	Favorite bool `json:"favorite"`
}

func (as *ArcticleSerializer) Response(db models.ArticleWithAutor) (ArcticleResponse, error) {
	ctx := as.C.Request.Context()
	userID := as.C.MustGet(models.KeyUserID).(uint)
	favorite, err := db.IsArticleFavorite(ctx, []uint{as.ID, userID})
	if err != nil {
		return ArcticleResponse{}, err
	}
	ctx = context.WithValue(ctx, models.KeyFlagFiled, models.FlagID)
	autor, err := db.FindOneUserByField(ctx, models.UserModel{ID: as.AutorID})
	if err != nil {
		return ArcticleResponse{}, err
	}
	autorProfile := users.ProfileSerializer{as.C, autor}
	autorResponse, err := autorProfile.Response(db)
	if err != nil {
		return ArcticleResponse{}, err
	}
	articleResponse := ArcticleResponse{
		ID:            as.ID,
		Slug:          as.Slug,
		Title:         as.Title,
		Description:   as.Description,
		Body:          as.Body,
		Tags:          as.Tags,
		CreatedAt:     as.CreatedAt.UTC().Format(time.RFC3339Nano),
		Autor:         autorResponse,
		NumberOfLikes: as.NumberOfLikes,
		Favorite:      favorite,
	}
	if as.UpdatedAt != nil {
		articleResponse.UpdatedAt = as.UpdatedAt.Format(time.RFC3339Nano)
	}
	return articleResponse, nil
}

type ArcticleListSerializer struct {
	C           *gin.Context
	ArticleList []models.ArticleModel
}

func (als *ArcticleListSerializer) Response(db models.ArticleWithAutor) ([]ArcticleResponse, error) {
	arrArticleResponse := make([]ArcticleResponse, 0, len(als.ArticleList))
	for _, article := range als.ArticleList {
		serialize := ArcticleSerializer{als.C, article}
		articleResponse, err := serialize.Response(db)
		if err != nil {
			return nil, err
		}
		arrArticleResponse = append(arrArticleResponse, articleResponse)
	}
	return arrArticleResponse, nil
}

type TagSerializer struct {
	C *gin.Context
	models.TagModel
}

type TagResponse struct {
	ID        uint                  `json:"id"`
	Name      string                `json:"name"`
	Autor     users.ProfileResponse `json:"autor_profile"`
	CreatedAt string                `json:"created_at"`
}

func (ts *TagSerializer) Response(db models.UserApproveFollowing) (TagResponse, error) {
	ctx := context.WithValue(ts.C.Request.Context(), models.KeyFlagFiled, models.FlagID)
	autor, err := db.FindOneUserByField(ctx, models.UserModel{ID: ts.AutorID})
	if err != nil {
		return TagResponse{}, err
	}
	autorProfile := users.ProfileSerializer{ts.C, autor}
	autorResponse, err := autorProfile.Response(db)
	if err != nil {
		return TagResponse{}, err
	}
	return TagResponse{
		ID:        ts.ID,
		Name:      ts.Name,
		Autor:     autorResponse,
		CreatedAt: ts.CreatedAt.Format(time.RFC3339Nano),
	}, nil
}

type TagListSerializer struct {
	C    *gin.Context
	Tags []models.TagModel
}

func (tls *TagListSerializer) Response(db models.UserApproveFollowing) ([]TagResponse, error) {
	arrTagResponse := make([]TagResponse, 0, len(tls.Tags))
	for _, tag := range tls.Tags {
		serialize := TagSerializer{tls.C, tag}
		tagResponse, err := serialize.Response(db)
		if err != nil {
			return nil, err
		}
		arrTagResponse = append(arrTagResponse, tagResponse)
	}
	return arrTagResponse, nil
}

type CommentSerialize struct {
	C *gin.Context
	models.CommentModel
}

type CommentResponse struct {
	ID        uint                  `json:"id"`
	Body      string                `json:"body"`
	CreatedAt string                `json:"created_at"`
	UpdatedAt string                `json:"updated_at,omitempty"`
	Autor     users.ProfileResponse `json:"autor_profile"`
}

func (cs *CommentSerialize) Response(db models.UserApproveFollowing) (CommentResponse, error) {
	ctx := context.WithValue(cs.C.Request.Context(), models.KeyFlagFiled, models.FlagID)
	autor, err := db.FindOneUserByField(ctx, models.UserModel{ID: cs.AutorID})
	if err != nil {
		return CommentResponse{}, err
	}
	serialize := users.ProfileSerializer{cs.C, autor}
	autorResponse, err := serialize.Response(db)
	if err != nil {
		return CommentResponse{}, err
	}
	commentResponse := CommentResponse{
		ID:        cs.ID,
		Body:      cs.Body,
		CreatedAt: cs.CreatedAt.Format(time.RFC3339Nano),
		Autor:     autorResponse,
	}
	if cs.UpdatedAt != nil {
		commentResponse.UpdatedAt = cs.UpdatedAt.Format(time.RFC3339Nano)
	}
	return commentResponse, nil
}

type CommentListSerialize struct {
	C        *gin.Context
	Comments []models.CommentModel
}

func (cls *CommentListSerialize) Response(db models.UserApproveFollowing) ([]CommentResponse, error) {
	arrCommentResponse := make([]CommentResponse, 0, len(cls.Comments))
	for _, comment := range cls.Comments {
		serialize := CommentSerialize{cls.C, comment}
		commentResponse, err := serialize.Response(db)
		if err != nil {
			return nil, err
		}
		arrCommentResponse = append(arrCommentResponse, commentResponse)
	}
	return arrCommentResponse, nil
}
