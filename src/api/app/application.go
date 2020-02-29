package app

import (
	"github.com/gin-gonic/gin"
	log "github.com/tajud99n/go-micro/src/api/log/zap"
)

var (
	router *gin.Engine
)

func init() {
	router = gin.Default()
}

func StartApp() {
	log.Info("about to map the urls")
	mapUrls()
	log.Info("urls successfully mapped")

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
