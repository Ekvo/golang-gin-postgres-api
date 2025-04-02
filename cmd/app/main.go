package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Ekvo/golang-gin-postgres-api/internal/services/autorization"
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/Ekvo/golang-gin-postgres-api/internal/transport/rest"
	"github.com/Ekvo/golang-gin-postgres-api/pkg/common"
)

func main() {
	ctx := context.Background()
	pool, err := source.InitDB(ctx, ".env")
	if err != nil {
		log.Fatalf("Pool: error - %v", err)
	}
	defer pool.Close()
	store := source.NewSQLSource(pool)
	router := gin.Default()

	first := router.Group("/zephyr")
	first.Use(common.ContextMiddleware(rest.CTXUsersTimeRequest))
	rest.UserBeforeRegister(first.Group("/connect"), store)

	first.Use(autorization.Autorization(store))
	rest.UserAfterRegister(first.Group("/user"), store)
	rest.SpeakerFolower(first.Group("/profile"), store)

	if err := router.Run("127.0.0.1:8000"); err != nil {
		log.Fatalf("server error - %v", err)
	}
}
