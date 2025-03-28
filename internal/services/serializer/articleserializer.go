package serializer

import (
	"context"
	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	vr "github.com/Ekvo/golang-gin-postgres-api/internal/variables"
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

	Autor ProfileResponse `json:"autor_profile"`

	NumberOfLikes uint `json:"number_of_likes"`

	Favorite bool `json:"favorite"`
}

func (as *ArcticleSerializer) Response(db models.ArticleWithAutor) (ArcticleResponse, error) {
	ctx := as.C.Request.Context()
	userID := as.C.MustGet(flag.KeyUserID).(uint)
	favorite, err := db.IsArticleFavorite(ctx, []uint{as.ID, userID})
	if err != nil {
		return ArcticleResponse{}, err
	}
	ctx = context.WithValue(ctx, flag.KeyFlagFiled, flag.FlagID)
	autor, err := db.FindOneUserByField(ctx, models.UserModel{ID: as.AutorID})
	if err != nil {
		return ArcticleResponse{}, err
	}
	autorProfile := ProfileSerializer{as.C, autor}
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
		CreatedAt:     as.CreatedAt.UTC().Format(vr.RFC3339Milli),
		Autor:         autorResponse,
		NumberOfLikes: as.NumberOfLikes,
		Favorite:      favorite,
	}
	if as.UpdatedAt != nil {
		articleResponse.UpdatedAt = as.UpdatedAt.Format(vr.RFC3339Milli)
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
