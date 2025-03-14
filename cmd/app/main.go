package main

import (
	"bytes"
	"database/sql"
	"github.com/Ekvo/golang-gin-postgres-api/internal/models"
	"github.com/Ekvo/golang-gin-postgres-api/internal/services/users"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/Ekvo/golang-gin-postgres-api/internal/transport"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
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
	//t := time.Now()
	m := models.CommentModel{
		ID: 7,
	}
	m.AutorID = 15

	//ctx := context.WithValue(context.Background(), services.UserID, uint(1))
	//err := store.EndCommentLife(ctx, m)
	//if err != nil {
	//	return
	//}

	router := gin.Default()

	first := router.Group("/zephyr")
	first.Use(common.ContextMiddleware(transport.CTXUsersTimeRequest))
	transport.UserBeforeRegister(first.Group("/connect"), store)

	first.Use(users.Autorization(store))
	transport.UserAfterRegister(first.Group("/user"), store)
	transport.SpeakerFolower(first.Group("/profile"), store)

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
