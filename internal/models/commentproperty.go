package models

import "github.com/Ekvo/golang-gin-postgres-api/pkg/common"

type CommentProperty struct {
	ArcticleSlug string

	//автор комментария
	AutorName string

	// см. 'ArticleProperty'
	common.TimeRange

	// см. 'ArticleProperty'
	common.LimitOffset
}
