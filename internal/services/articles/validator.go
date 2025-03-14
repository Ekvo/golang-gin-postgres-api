package articles

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
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

func NewArticleCreateValidator() *ArticleCreateValidator {
	return &ArticleCreateValidator{}
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
	acv.aModel.AutorID = c.MustGet(models.KeyUserID).(uint)
	acv.aModel.Description = acv.Article.Description
	acv.aModel.Body = acv.Article.Body
	acv.aModel.Tags = acv.Article.Tags
	acv.aModel.CreatedAt = time.Now()
	return nil
}

type TagCreateValidator struct {
	Tag struct {
		Name string `from:"tag_name" json:"tag_name" binding:"required,startswith=#,min=2,max=26"`
	} `json:"tag_update"`
	tModel models.TagModel `json:"-"`
}

func NewTagCreateValidator() *TagCreateValidator {
	return &TagCreateValidator{}
}

func (tcv *TagCreateValidator) Model() models.TagModel {
	return tcv.tModel
}

func (tcv *TagCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, tcv); err != nil {
		return err
	}
	tcv.tModel.Name = tcv.Tag.Name
	tcv.tModel.AutorID = c.MustGet(models.KeyUserID).(uint)
	tcv.tModel.CreatedAt = time.Now()
	return nil
}

type CommentCreateValidator struct {
	Comment struct {
		Body string `form:"body" json:"body" binding:"min=1,max=2048"`
	} `json:"comment_update"`
	cModel models.CommentModel `json:"-"`
}

func NewCommentCreateValidator() *CommentCreateValidator {
	return &CommentCreateValidator{}
}

func (ccv *CommentCreateValidator) Model() models.CommentModel {
	return ccv.cModel
}

func (ccv *CommentCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, ccv); err != nil {
		return err
	}
	ccv.cModel.AutorID = c.MustGet(models.KeyUserID).(uint)
	ccv.cModel.Body = ccv.Comment.Body
	ccv.cModel.CreatedAt = time.Now()
	return nil
}

type ArticlePropertyValidator struct {
	Property struct {
		Tags      []string  `from:"tags" json:"tags"`
		AutorName string    `form:"autor_name" json:"autor_name" binding:"omitempty,alphanum,min=1,max=128"'`
		Favorited bool      `form:"favorited" json:"favorited"`
		StartDate time.Time `from:"start_date" json:"start_date"`
		EndDate   time.Time `form:"end_date" json:"end_date"`
		Limit     uint      `form:"limit" json:"limit" binding:"omitempty,numeric"`
		Offset    uint      `form:"offset" json:"offset" binding:"omitempty,numeric"`
	} `json:"artcicle_property"`
	aProperty models.ArticleProperty `json:"-"`
}

func NewArticlePropertyValidator() *ArticlePropertyValidator {
	return &ArticlePropertyValidator{}
}

func (apv *ArticlePropertyValidator) Model() models.ArticleProperty {
	return apv.aProperty
}

func (apv *ArticlePropertyValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, apv); err != nil {
		return err
	}
	apv.aProperty.Tags = apv.Property.Tags
	apv.aProperty.AutorName = apv.Property.AutorName
	apv.aProperty.Favorited = apv.Property.Favorited
	if apv.Property.Limit == 0 {
		apv.Property.Limit = 20
	}
	apv.aProperty.Limit = apv.Property.Limit
	apv.aProperty.Offset = apv.Property.Offset

	if start := apv.Property.StartDate; !start.IsZero() {
		apv.aProperty.StartDate = start
	}
	if end := apv.Property.EndDate; !end.IsZero() {
		apv.aProperty.EndDate = end
	} else {
		apv.aProperty.EndDate = time.Now()
	}
	return nil
}

type TagPropertyValidator struct {
	Property struct {
		AutorName string    `form:"autor_name" json:"autor_name" binding:"omitempty,alphanum,min=1,max=128"`
		StartDate time.Time `from:"start_date" json:"start_date"`
		EndDate   time.Time `form:"end_date" json:"end_date"`
		Limit     uint      `form:"limit" json:"limit" binding:"omitempty,numeric"`
		Offset    uint      `form:"offset" json:"offset" binding:"omitempty,numeric"`
	} `json:"tag_property"`
	tProperty models.TagPropery `json:"-"`
}

func NewTagPropertyValidator() *TagPropertyValidator {
	return &TagPropertyValidator{}
}

func (tpv *TagPropertyValidator) Model() models.TagPropery {
	return tpv.tProperty
}

func (tpv *TagPropertyValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, tpv); err != nil {
		return err
	}
	tpv.tProperty.AutorName = tpv.Property.AutorName
	if tpv.Property.Limit == 0 {
		tpv.Property.Limit = 20
	}
	tpv.tProperty.Limit = tpv.Property.Limit
	tpv.tProperty.Offset = tpv.Property.Offset
	if start := tpv.Property.StartDate; !start.IsZero() {
		tpv.tProperty.StartDate = start
	}
	if end := tpv.Property.EndDate; !end.IsZero() {
		tpv.tProperty.EndDate = end
	} else {
		tpv.tProperty.EndDate = time.Now()
	}
	return nil
}

type CommentPropertyValidator struct {
	Property struct {
		ArcticleSlug string    `form:"arcticle_slug" json:"arcticle_slug" binding:"required,aphanum,min=1,max=50"`
		AutorName    string    `form:"autor_name" json:"autor_name" binding:"omitempty,alphanum,min=1,max=128"`
		StartDate    time.Time `from:"start_date" json:"start_date"`
		EndDate      time.Time `form:"end_date" json:"end_date"`
		Limit        uint      `form:"limit" json:"limit" binding:"omitempty,numeric"`
		Offset       uint      `form:"offset" json:"offset" binding:"omitempty,numeric"`
	} `json:"comment_property"`
	cProperty models.CommentProperty `json:"-"`
}

func NewCommentPropertyValidator() *CommentPropertyValidator {
	return &CommentPropertyValidator{}
}

func (cpv *CommentPropertyValidator) Model() models.CommentProperty {
	return cpv.cProperty
}

func (cpv *CommentPropertyValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, cpv); err != nil {
		return err
	}
	cpv.cProperty.ArcticleSlug = cpv.Property.ArcticleSlug
	cpv.cProperty.AutorName = cpv.Property.AutorName
	if cpv.Property.Limit == 0 {
		cpv.Property.Limit = 20
	}
	cpv.cProperty.Limit = cpv.Property.Limit
	cpv.cProperty.Offset = cpv.Property.Offset
	if start := cpv.Property.StartDate; !start.IsZero() {
		cpv.cProperty.StartDate = start
	}
	if end := cpv.Property.EndDate; !end.IsZero() {
		cpv.cProperty.EndDate = end
	} else {
		cpv.cProperty.EndDate = time.Now()
	}
	return nil
}
