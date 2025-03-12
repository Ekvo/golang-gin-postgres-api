package models

import (
	"github.com/gin-gonic/gin"
)

// ValidatorModel - описывает свойсва валидации объектов
type ValidatorModel[V any, M any] interface {
	Bind(c *gin.Context) error
	Model() M
}
