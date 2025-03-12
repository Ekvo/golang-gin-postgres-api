package transport

import (
	"github.com/Ekvo/golang-gin-postgres-api/internal/source"
	"github.com/gin-gonic/gin"
)

func ArcticleAfterRegister(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.POST("/", ArcticleCreate(storeDB))
	router.PUT("/:slug", ArcticleUpdate(storeDB))
	router.DELETE("/:slug", ArcticleRemove(storeDB))

	router.POST("/:slug/favorite", ArcricleToFaorite(storeDB))
	router.DELETE("/:slug/favorite", ArcticleUnFavorite(storeDB))

	router.POST("/slug:/commentss", CommentCreate(storeDB))
	router.PUT("/slug:/comments/:id", CommentUpdate(storeDB))
	router.DELETE("/slug:/comments/:id", CommentRemove(storeDB))
}

func ArcticleBeforeRegister(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.GET("/:slug", ArcticleRetrive(storeDB))
	router.POST("/", ArcticleListretrive(storeDB))
	router.GET("/:slug/comments", CommentListRetrive)
}

func TagAfterRegister(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.POST("/tag", TagCreate(storeDB))
}

func TagBeforeRegister(router *gin.RouterGroup, storeDB *source.SQLSource) {
	router.POST("/tags", TagListRetrive(storeDB))
}
