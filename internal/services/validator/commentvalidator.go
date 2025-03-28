package validator

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/flag"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

type CommentCreateValidator struct {
	Comment struct {
		Body string `form:"body" json:"body" binding:"min=1,max=2048"`
	} `json:"comment_update"`
	cModel models.CommentModel `json:"-"`
}

func NewCommentCreateValidator() CommentCreateValidator {
	return CommentCreateValidator{}
}

func (ccv *CommentCreateValidator) Model() models.CommentModel {
	return ccv.cModel
}

func (ccv *CommentCreateValidator) Bind(c *gin.Context) error {
	if err := common.Bind(c, ccv); err != nil {
		return err
	}
	ccv.cModel.AutorID = c.MustGet(flag.KeyUserID).(uint)
	ccv.cModel.Body = ccv.Comment.Body
	ccv.cModel.CreatedAt = time.Now()
	return nil
}
