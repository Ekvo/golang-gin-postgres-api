package main

import (
	"database/sql"
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

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS NEW1
(
    id serial
);`)
	if err != nil {
		return
	}

	store := source.NewSQLSource(db)

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
