package validator

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type ArticleCreateValidator struct {
	Article struct {
		Title       string   `from:"title" json:"title" binding:"required,alphanum,min=1,max=50"`
		Description string   `form:"description" json:"description" binding:"required,min=1,max=2048"`
		Body        string   `form:"body" json:"body" binding:"required,min=1,max=2048"`
		Tags        []string `form:"tag_list" json:"tag_list"`
	} `json:"article_update"`
	aModel models.ArticleModel `json:"-"`
}

func NewArticleCreateValidator() ArticleCreateValidator {
	return ArticleCreateValidator{}
}

func (acv *ArticleCreateValidator) Model() models.ArticleModel {
	return acv.aModel
}

func (acv *ArticleCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, acv); err != nil {
		return err
	}
	acv.aModel.Slug = slug.Make(acv.Article.Title)
	acv.aModel.Title = acv.Article.Title
	acv.aModel.AutorID = c.MustGet(flag.KeyUserID).(uint)
	acv.aModel.Description = acv.Article.Description
	acv.aModel.Body = acv.Article.Body
	acv.aModel.Tags = acv.Article.Tags
	acv.aModel.CreatedAt = time.Now()
	return nil
}
