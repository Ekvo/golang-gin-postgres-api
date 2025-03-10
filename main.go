package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/users"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	users2 "github.com/Ekvo/golang-gin-postgres-api/internal/transport"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func main() {
	db, errDB := sql.Open("postgres", `
host=127.0.0.1 port=5432 user=postgres password=1234567 dbname=zephyr sslmode=disable`)
	if errDB != nil {
		log.Fatalf("db error - %v", errDB)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("db.Colse error - %v", err)
		}
	}()

	store := source.NewSQLSource(db)

	ctx := context.WithValue(context.Background(), services.UserID, uint(1))
	//data := models.ArticleProperty{
	//	Tags:      models.ArcticleTags{"#sport"},
	//	AutorName: "dd",
	//	Favorited: false,
	//	Limit:     10,
	//	Offset:    0,
	//}
	//
	//tags, err := store.FindArticleList(ctx, data)
	//if err != nil {
	//	return
	//}
	//fmt.Printf("%v", tags)
	start := time.Time{}
	end := time.Now()

	comment, err := store.FindCommentList(ctx, models.CommentProperty{
		ArcticleSlug: "second",
		TimeRange: common.TimeRange{
			StartDate: start,
			EndDate:   end,
		},
		LimitOffset: common.LimitOffset{
			Limit:  10,
			Offset: 0,
		},
	})
	if err != nil {
		return
	}
	fmt.Printf("%v", comment)

	/*
		type CommentProperty struct {
			ArcticleSlug string

			//автор комментария
			AutorName string

			StartDate *time.Time
			EndDate   *time.Time

			Limit  uint
			Offset uint
		}
	*/

	router := gin.Default()

	first := router.Group("/zephyr")
	first.Use(common.ContextMiddleware(users2.CTXUsersTimeRequest))
	users2.UserBeforeRegister(first.Group("/connect"), store)

	first.Use(users.Autorization(store))
	users2.UserAfterRegister(first.Group("/user"), store)
	users2.SpeakerFolower(first.Group("/profile"), store)

	if err := router.Run("127.0.0.1:8000"); err != nil {
		log.Fatalf("server error - %w", err)
	}
}

func arrayLine(data []string) string {
	var buff bytes.Buffer

	for i := 0; i < len(data); i++ {
		buff.WriteByte('\'')
		buff.WriteString(data[i])
		buff.Write([]byte{'\'', ','})
	}
	if n := buff.Len(); n > 0 {
		buff.Truncate(n - 1)
	}
	return buff.String()
}
