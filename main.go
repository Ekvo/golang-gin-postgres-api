package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"

	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/Ekvo/golang-gin-postgres-api/internal/users"
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

	s := source.NewSQLSource(db)

	r := gin.Default()
	first := r.Group("/zephyr")
	users.UserBeforeRegister(first.Group("/connect"), s)
	/*
		r.GET("/", func(c *gin.Context) {
			id := c.Request.URL.Query().Get("id")
			idInt, _ := strconv.Atoi(id)
			ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
			defer cancel()

			u, err := s.FindOneUser(ctx, source.UserModel{ID: uint(idInt)})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
			c.JSON(http.StatusOK, gin.H{"user": u})
		})
	*/
	r.POST("/zephyr", func(c *gin.Context) {
		m := users.NewUserCreateValidator()
		if err := m.Bind(c); err != nil {
			c.JSON(http.StatusBadRequest, common.NewDataErrorValidator(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": m.User})
	})

	if err := r.Run("127.0.0.1:8000"); err != nil {
		log.Fatalf("server error - %w", err)
	}
}
