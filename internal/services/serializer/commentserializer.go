package serializer

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	vr "github.com/Ekvo/golang-gin-postgres-api/internal/variables"
)

type CommentSerialize struct {
	C *gin.Context
	models.CommentModel
}

type CommentResponse struct {
	ID        uint            `json:"id"`
	Body      string          `json:"body"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at,omitempty"`
	Autor     ProfileResponse `json:"autor_profile"`
}

func (cs *CommentSerialize) Response(db models.UserApproveFollowing) (CommentResponse, error) {
	ctx := context.WithValue(cs.C.Request.Context(), flag.KeyFlagFiled, flag.FlagID)
	autor, err := db.FindOneUserByField(ctx, models.UserModel{ID: cs.AutorID})
	if err != nil {
		return CommentResponse{}, err
	}
	serialize := ProfileSerializer{cs.C, autor}
	autorResponse, err := serialize.Response(db)
	if err != nil {
		return CommentResponse{}, err
	}
	commentResponse := CommentResponse{
		ID:        cs.ID,
		Body:      cs.Body,
		CreatedAt: cs.CreatedAt.Format(vr.RFC3339Milli),
		Autor:     autorResponse,
	}
	if cs.UpdatedAt != nil {
		commentResponse.UpdatedAt = cs.UpdatedAt.Format(vr.RFC3339Milli)
	}
	return commentResponse, nil
}

type CommentListSerialize struct {
	C        *gin.Context
	Comments []models.CommentModel
}

func (cls *CommentListSerialize) Response(db models.UserApproveFollowing) ([]CommentResponse, error) {
	arrCommentResponse := make([]CommentResponse, 0, len(cls.Comments))
	for _, comment := range cls.Comments {
		serialize := CommentSerialize{cls.C, comment}
		commentResponse, err := serialize.Response(db)
		if err != nil {
			return nil, err
		}
		arrCommentResponse = append(arrCommentResponse, commentResponse)
	}
	return arrCommentResponse, nil
}
