package models

import (
	"github.com/gin-gonic/gin"
)

// Models - типы структур используемые в 'ValidatorModel' см. ниже
type Models interface {
	UserModel |
		ArticleModel | ArticleProperty |
		TagModel |
		CommentModel | CommentProperty
}

// ValidatorModel - описывает свойсва валидации объектов
type ValidatorModel[M Models] interface {
	Bind(c *gin.Context) error
	Model() M
}
