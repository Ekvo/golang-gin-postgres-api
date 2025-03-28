package models

import "github.com/Ekvo/golang-gin-postgres-api/pkg/common"

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
