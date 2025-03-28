package models

import "github.com/Ekvo/golang-gin-postgres-api/pkg/common"

// TagPropery - свойсва для поиска набора 'TagModel'
type TagPropery struct {
	AutorName string

	common.TimeRange

	common.LimitOffset
}
