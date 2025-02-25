package main

import (
	"github.com/Ekvo/golang-gin-postgres-api/internal/common"
	"github.com/Ekvo/golang-gin-postgres-api/internal/users"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()

	r.POST("/zephyr", func(c *gin.Context) {
		m := users.NewUserCreateValidator()
		if err := m.Bind(c); err != nil {
			c.JSON(http.StatusBadRequest, common.NewDataErrorValidator(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": m.D()})
	})

	if err := r.Run("127.0.0.1:8000"); err != nil {
		log.Fatalf("server error - %w", err)
	}
}
