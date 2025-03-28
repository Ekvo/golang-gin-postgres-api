package serializer

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	vr "github.com/Ekvo/golang-gin-postgres-api/internal/variables"
)

type TagSerializer struct {
	C *gin.Context
	models.TagModel
}

type TagResponse struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	Autor     ProfileResponse `json:"autor_profile"`
	CreatedAt string          `json:"created_at"`
}

func (ts *TagSerializer) Response(db models.UserApproveFollowing) (TagResponse, error) {
	ctx := context.WithValue(ts.C.Request.Context(), flag.KeyFlagFiled, flag.FlagID)
	autor, err := db.FindOneUserByField(ctx, models.UserModel{ID: ts.AutorID})
	if err != nil {
		return TagResponse{}, err
	}
	autorProfile := ProfileSerializer{ts.C, autor}
	autorResponse, err := autorProfile.Response(db)
	if err != nil {
		return TagResponse{}, err
	}
	return TagResponse{
		ID:        ts.ID,
		Name:      ts.Name,
		Autor:     autorResponse,
		CreatedAt: ts.CreatedAt.Format(vr.RFC3339Milli),
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
