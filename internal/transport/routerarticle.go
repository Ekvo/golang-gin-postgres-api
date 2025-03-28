package transport

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	mod "github.com/Ekvo/golang-gin-postgres-api/internal/models"
	art "github.com/Ekvo/golang-gin-postgres-api/internal/services/articles"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/users/access"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/users/flag"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

// ErrArticlesParam - mark for ':id' router.PUT("/slug:/comments/:id", CommentUpdate(storeDB))
var ErrArticlesParam = errors.New("invalid param")

func ArcticleWrite(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.POST("/", ArcticleCreate(storeDB))
	router.PUT("/:slug", ArcticleUpdate(storeDB))
	router.DELETE("/:slug", ArcticleRemove(storeDB))

	router.POST("/:slug/favorite", ArcricleToFaorite(storeDB))
	router.DELETE("/:slug/favorite", ArcticleUnFavorite(storeDB))

	router.POST("/slug:/comments", CommentCreate(storeDB))
	router.PUT("/slug:/comments/:id", CommentUpdate(storeDB))
	router.DELETE("/slug:/comments/:id", CommentRemove(storeDB))
}

func ArcticleRead(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.GET("/", ArcticleListRetrive(storeDB))
	router.GET("/:slug", ArcticleRetrive(storeDB))
	router.GET("/:slug/comments", CommentListRetrive(storeDB))
}

func TagWrite(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.POST("/", TagCreate(storeDB))
}

func TagRead(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.GET("/:tagname", TagRetrive(storeDB))
	router.GET("/", TagListRetrive(storeDB))
}

func ArcticleCreate(db mod.ArticleWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !access.WriteAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewArticleCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator)
			return
		}
		var err error = nil
		articleModel := modelValidator.Model()
		ctx := c.Request.Context()
		articleModel.ID, err = db.SaveOneArticle(ctx, articleModel)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serializer := art.ArcticleSerializer{c, articleModel}
		articleResponse, err := serializer.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"article": articleResponse})
	}
}

func ArcticleUpdate(db mod.ArticleWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAccess := c.MustGet(flag.KeyUserAccess).(string)
		if !access.WriteAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewArticleCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		slug := c.Param("slug")
		ctx := c.Request.Context()
		oldArticleData, err := db.FindOneArticle(ctx, slug)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("article", source.ErrSourceNotFound))
			return
		}
		newArticleData := modelValidator.Model()
		if newArticleData.AutorID != oldArticleData.AutorID && !access.AdminAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		newArticleData.ID = oldArticleData.ID
		// если измениея вносит админ
		newArticleData.AutorID = oldArticleData.AutorID
		newArticleData.UpdatedAt = new(time.Time)
		*newArticleData.UpdatedAt = newArticleData.CreatedAt
		newArticleData.CreatedAt = oldArticleData.CreatedAt
		if err := db.NewDataArticle(ctx, newArticleData); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("article", source.ErrSourceNoUpdate))
			return
		}
		serializer := art.ArcticleSerializer{c, newArticleData}
		articleResponse, err := serializer.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"article": articleResponse})
	}
}
func ArcticleRemove(db mod.ArticleUpdateFind) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAccess := c.MustGet(flag.KeyUserAccess).(string)
		if !access.WriteAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		slug := c.Param("slug")
		ctx := c.Request.Context()
		article, err := db.FindOneArticle(ctx, slug)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("article", source.ErrSourceNotFound))
			return
		}
		userID := c.MustGet(flag.KeyUserID).(uint)
		if userID != article.AutorID && !access.AdminAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		if err := db.EndArticleLife(ctx, slug); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"article": "status - deleted"})
	}
}

func ArcricleToFaorite(db mod.ArticleWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !access.WriteAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		slug := c.Param("slug")
		ctx := context.WithValue(c.Request.Context(), flag.KeyUserID, c.MustGet(flag.KeyUserID).(uint))
		if err := db.ArticleToFavorite(ctx, slug); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("article", source.ErrSourceNotFound))
			return
		}
		article, err := db.FindOneArticle(ctx, slug)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serialize := art.ArcticleSerializer{c, article}
		articleResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"article": articleResponse})
	}
}

func ArcticleUnFavorite(db mod.ArticleWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !access.WriteAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		slug := c.Param("slug")
		ctx := context.WithValue(c.Request.Context(), flag.KeyUserID, c.MustGet(flag.KeyUserID).(uint))
		if err := db.ArticleUnFovarite(ctx, slug); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("article", source.ErrSourceNotFound))
			return
		}
		article, err := db.FindOneArticle(ctx, slug)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serialize := art.ArcticleSerializer{c, article}
		articleResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"article": articleResponse})
	}
}

func CommentCreate(db mod.CommentWihtAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !access.WriteAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewCommentCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		var err error = nil
		comment := modelValidator.Model()
		ctx := c.Request.Context()
		comment.ID, err = db.SaveOneComment(ctx, comment)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serialize := art.CommentSerialize{c, comment}
		commentResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"comment": commentResponse})
	}
}

func CommentUpdate(db mod.CommentWihtAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAccess := c.MustGet(flag.KeyUserAccess).(string)
		if !access.WriteAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewCommentCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		commentID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, common.NewError("param", ErrArticlesParam))
			return
		}
		ctx := c.Request.Context()
		oldComment, err := db.FindOneComment(ctx, uint(commentID))
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("comment", source.ErrSourceNotFound))
			return
		}
		newComment := modelValidator.Model()
		if newComment.AutorID != oldComment.AutorID && !access.AdminAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		newComment.ID = oldComment.ID
		// если измения вносит админ
		newComment.AutorID = oldComment.AutorID
		newComment.UpdatedAt = new(time.Time)
		*newComment.UpdatedAt = newComment.CreatedAt
		newComment.CreatedAt = oldComment.CreatedAt
		if err := db.NewDataComment(ctx, newComment); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serialize := art.CommentSerialize{c, newComment}
		commentResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"comment": commentResponse})
	}
}

func CommentRemove(db mod.CommentWihtAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAccess := c.MustGet(flag.KeyUserAccess).(string)
		if !access.WriteAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		commentID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrArticlesParam)
			return
		}
		ctx := c.Request.Context()
		comment, err := db.FindOneComment(ctx, uint(commentID))
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("comment", source.ErrSourceNotFound))
			return
		}
		userID := c.MustGet(flag.KeyUserID).(uint)
		if userID != comment.AutorID && !access.AdminAccess(userAccess) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		if err := db.EndCommentLife(ctx, comment.ID); err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"comment": "status - deleted"})
	}
}

func ArcticleRetrive(db mod.ArticleWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !access.ReadAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewArticlePropertyValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		ctx := c.Request.Context()
		article, err := db.FindOneArticle(ctx, modelValidator.Model())
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("article", source.ErrSourceNotFound))
			return
		}
		serialize := art.ArcticleSerializer{c, article}
		articelResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"article": articelResponse})
	}
}

func CommentListRetrive(db mod.CommentWihtAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !access.ReadAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewCommentPropertyValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		ctx := c.Request.Context()
		commentList, err := db.FindCommentList(ctx, modelValidator.Model())
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("comment_list", source.ErrSourceNotFound))
			return
		}
		serialize := art.CommentListSerialize{c, commentList}
		commentListResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"commet_list": commentListResponse})
	}
}

func ArcticleListRetrive(db mod.ArticleWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !access.ReadAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewArticlePropertyValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		ctx := c.Request.Context()
		articleList, err := db.FindArticleList(ctx, modelValidator.Model())
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusNotFound, common.NewError("access", source.ErrSourceNotFound))
			return
		}
		serialize := art.ArcticleListSerializer{c, articleList}
		articleListresponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"article_tils": articleListresponse})
	}
}

func TagCreate(db mod.TagWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !(access.AdminAccess(c.MustGet(flag.KeyUserAccess).(string))) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewTagCreateValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		var err error = nil
		tag := modelValidator.Model()
		ctx := c.Request.Context()
		tag.ID, err = db.SaveOneTag(ctx, tag)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("data_base", err))
			return
		}
		serialize := art.TagSerializer{c, tag}
		tagResponse, err := serialize.Response(db)
		if err != nil {
			if err == context.DeadlineExceeded {
				return
			}
			c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"tag": tagResponse})
	}
}

func TagRetrive(db mod.TagWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		if !access.ReadAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		tagName := c.Param("tagname")
		tag, err := db.FindOneTag(ctx, tagName)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("tag", source.ErrSourceNotFound))
			}
			return
		}
		serialize := art.TagSerializer{c, tag}
		tagResponse, err := serialize.Response(db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"tag": tagResponse})
	}
}

func TagListRetrive(db mod.TagWithAutor) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if ctx.Err() != nil {
			return
		}
		if !access.ReadAccess(c.MustGet(flag.KeyUserAccess).(string)) {
			c.JSON(http.StatusForbidden, common.NewError("access", access.ErrServicesUsersAccessDenied))
			return
		}
		modelValidator := art.NewTagPropertyValidator()
		if err := modelValidator.Bind(c); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewDataErrorValidator(err))
			return
		}
		tagList, err := db.TagsList(ctx, modelValidator.Model())
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusNotFound, common.NewError("tag_list", source.ErrSourceNotFound))
			}
			return
		}
		serialize := art.TagListSerializer{c, tagList}
		tagListResponse, err := serialize.Response(db)
		if err != nil {
			if err != context.DeadlineExceeded {
				c.JSON(http.StatusInternalServerError, common.NewError("serializer", err))
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"tag_list": tagListResponse})
	}
}
