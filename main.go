package main

import (
	"database/sql"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/Ekvo/golang-gin-postgres-api/internal/users"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
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

	router := gin.Default()

	first := router.Group("/zephyr")
	users.UserBeforeRegister(first.Group("/connect"), store)
	first.Use(users.Autorization(store))
	users.UserAfterRegister(first.Group("/user"), store)
	users.SpeakerFolower(first.Group("/profile"), store)

	srv := &http.Server{
		Addr:         "127.0.0.1:8000",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error - %w", err)
	}
}
